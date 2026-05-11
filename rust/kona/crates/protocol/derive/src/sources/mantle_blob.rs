//! Mantle Blob Data Source
//!
//! This source provides blob data with Mantle blob decoding format.
//! Mantle blob decoding concatenates all decoded blobs and then RLP decodes them.

use crate::{
    BlobData, BlobProvider, BlobProviderError, ChainProvider, DataAvailabilityProvider,
    PipelineError, PipelineResult,
};
use alloc::{boxed::Box, format, string::ToString, vec::Vec};
use alloy_consensus::{
    Transaction, TxEip4844Variant, TxEnvelope, TxType, transaction::SignerRecoverable,
};
use alloy_eips::eip4844::IndexedBlobHash;
use alloy_primitives::{Address, Bytes};
use alloy_rlp::Decodable;
use async_trait::async_trait;
use kona_protocol::BlockInfo;

/// A wrapper for RLP decoding of Vec<Bytes>
#[derive(Debug, Clone)]
struct VecOfBytes(pub Vec<Bytes>);

impl Decodable for VecOfBytes {
    fn decode(buf: &mut &[u8]) -> alloy_rlp::Result<Self> {
        let vec: Vec<Bytes> = Decodable::decode(buf)?;
        Ok(VecOfBytes(vec))
    }
}

/// A data iterator that reads from a blob using Mantle's blob decoding format.
#[derive(Debug, Clone)]
pub struct MantleBlobSource<F, B>
where
    F: ChainProvider + Send,
    B: BlobProvider + Send,
{
    /// Chain provider.
    pub chain_provider: F,
    /// Fetches blobs.
    pub blob_fetcher: B,
    /// The address of the batcher contract.
    pub batcher_address: Address,
    /// Data.
    pub data: Vec<BlobData>,
    /// Whether the source is open.
    pub open: bool,
    /// Whether Mantle RLP format decode has failed, signaling transition to standard blob format.
    /// Matches Go's `blobSourceChanged` toggle in `DataSourceFactory`.
    /// This field persists across `clear()` calls — once set, it remains true for all subsequent
    /// blocks, ensuring that failed Mantle RLP decoding is not retried.
    pub(crate) mantle_format_failed: bool,
}

impl<F, B> MantleBlobSource<F, B>
where
    F: ChainProvider + Send,
    B: BlobProvider + Send,
{
    /// Creates a new Mantle blob source.
    pub const fn new(chain_provider: F, blob_fetcher: B, batcher_address: Address) -> Self {
        Self {
            chain_provider,
            blob_fetcher,
            batcher_address,
            data: Vec::new(),
            open: false,
            mantle_format_failed: false,
        }
    }

    fn decode_mantle_rlp_frames(bytes: &[u8]) -> Option<Vec<Bytes>> {
        let mut rlp_slice = bytes;
        let decoded = VecOfBytes::decode(&mut rlp_slice).ok()?;
        if rlp_slice.is_empty() { Some(decoded.0) } else { None }
    }

    /// Extracts blob data and tracks which blobs belong to which transaction.
    /// Returns: (all_blob_data, all_blob_hashes, tx_blob_counts)
    /// - all_blob_data: all BlobData in order (calldata or blob placeholders)
    /// - all_blob_hashes: all IndexedBlobHash
    /// - tx_blob_counts: number of blobs per transaction (0 for calldata tx, N for blob tx with N blobs)
    fn extract_blob_data(
        &self,
        txs: Vec<TxEnvelope>,
        batcher_address: Address,
    ) -> (Vec<BlobData>, Vec<IndexedBlobHash>, Vec<usize>) {
        let mut index: u64 = 0;
        let mut data = Vec::new();
        let mut hashes = Vec::new();
        let mut tx_blob_counts = Vec::new();
        for tx in txs {
            let (tx_kind, calldata, blob_hashes) = match &tx {
                TxEnvelope::Legacy(tx) => (tx.tx().to(), tx.tx().input.clone(), None),
                TxEnvelope::Eip2930(tx) => (tx.tx().to(), tx.tx().input.clone(), None),
                TxEnvelope::Eip1559(tx) => (tx.tx().to(), tx.tx().input.clone(), None),
                TxEnvelope::Eip4844(blob_tx_wrapper) => match blob_tx_wrapper.tx() {
                    TxEip4844Variant::TxEip4844(tx) => {
                        (tx.to(), tx.input.clone(), Some(tx.blob_versioned_hashes.clone()))
                    }
                    TxEip4844Variant::TxEip4844WithSidecar(tx) => {
                        let tx = tx.tx();
                        (tx.to(), tx.input.clone(), Some(tx.blob_versioned_hashes.clone()))
                    }
                },
                _ => continue,
            };
            let Some(to) = tx_kind else { continue };

            if to != self.batcher_address {
                index += blob_hashes.map_or(0, |h| h.len() as u64);
                continue;
            }
            if tx.recover_signer().unwrap_or_default() != batcher_address {
                index += blob_hashes.map_or(0, |h| h.len() as u64);
                continue;
            }
            if tx.tx_type() != TxType::Eip4844 {
                let blob_data = BlobData { data: None, calldata: Some(calldata.to_vec().into()) };
                data.push(blob_data);
                tx_blob_counts.push(0); // Calldata tx has 0 blobs
                continue;
            }
            if !calldata.is_empty() {
                let hash = match &tx {
                    TxEnvelope::Legacy(tx) => Some(tx.hash()),
                    TxEnvelope::Eip2930(tx) => Some(tx.hash()),
                    TxEnvelope::Eip1559(tx) => Some(tx.hash()),
                    TxEnvelope::Eip4844(blob_tx_wrapper) => Some(blob_tx_wrapper.hash()),
                    _ => None,
                };
                warn!(target: "mantle_blob_source", "Blob tx has calldata, which will be ignored: {hash:?}");
            }
            let blob_hashes = if let Some(b) = blob_hashes {
                b
            } else {
                continue;
            };
            let tx_blob_count = blob_hashes.len();
            for hash in blob_hashes {
                let indexed = IndexedBlobHash { hash, index };
                hashes.push(indexed);
                data.push(BlobData::default());
                index += 1;
            }
            tx_blob_counts.push(tx_blob_count);
        }
        #[cfg(feature = "metrics")]
        metrics::gauge!(
            crate::metrics::Metrics::PIPELINE_DATA_AVAILABILITY_PROVIDER,
            "source" => "mantle_blobs",
        )
        .increment(data.len() as f64);
        (data, hashes, tx_blob_counts)
    }

    /// Loads blob data into the source if it is not open.
    async fn load_blobs(
        &mut self,
        block_ref: &BlockInfo,
        batcher_address: Address,
    ) -> Result<(), BlobProviderError> {
        if self.open {
            return Ok(());
        }

        let info = self
            .chain_provider
            .block_info_and_transactions_by_hash(block_ref.hash)
            .await
            .map_err(|e| BlobProviderError::Backend(e.to_string()))?;

        let (mut data, blob_hashes, tx_blob_counts) = self.extract_blob_data(info.1, batcher_address);

        // If there are no hashes, set the calldata and return.
        if blob_hashes.is_empty() {
            self.open = true;
            self.data = data;
            return Ok(());
        }

        let blobs =
            self.blob_fetcher.get_and_validate_blobs(block_ref, &blob_hashes).await.map_err(
                |e| {
                    warn!(target: "mantle_blob_source", "Failed to fetch blobs: {e}");
                    BlobProviderError::Backend(e.to_string())
                },
            )?;

        // Process each transaction's blobs separately
        let mut result_data = Vec::new();
        let mut blob_index = 0;
        let mut data_index = 0;

        for tx_blob_count in tx_blob_counts {
            if tx_blob_count == 0 {
                // Calldata tx: keep as-is
                result_data.push(data[data_index].clone());
                data_index += 1;
                continue;
            }

            // EIP4844 tx: try Mantle format (RLP list) then fallback to standard format
            let tx_data_range = data_index..data_index + tx_blob_count;
            let tx_blobs = &mut data[tx_data_range.clone()];
            let mut fallback_result = Vec::new();
            let mut tx_whole_blob_data = Vec::new();
            let mut all_blobs_valid = true;

            // Fill and decode each blob for this transaction.
            // Matches Go's processTxBlobs: track allBlobsValid to only attempt
            // Mantle RLP when every blob decoded successfully.
            for blob_data in tx_blobs.iter_mut() {
                let current_blob_index = blob_index;
                match blob_data.fill(&blobs, blob_index) {
                    Ok(should_increment) => {
                        if should_increment {
                            blob_index += 1;
                        }
                    }
                    Err(e) => {
                        return Err(e.into());
                    }
                }
                // Decode blob data (EIP-4844 standard decoding).
                // If this fails, mark all_blobs_valid = false and skip,
                // matching Go's `allBlobsValid = false; continue` pattern.
                let decoded = match blob_data.decode() {
                    Ok(decoded) => decoded,
                    Err(e) => {
                        warn!(
                            target: "mantle_blob_source",
                            "Blob decode failed at index {}: {:?}, skipping blob",
                            current_blob_index,
                            e
                        );
                        all_blobs_valid = false;
                        continue;
                    }
                };
                fallback_result.push(BlobData { data: Some(decoded.clone()), calldata: None });
                tx_whole_blob_data.extend_from_slice(&decoded);
            }

            // Try Mantle format if all blobs are valid, matching Go's:
            //   if allBlobsValid && len(txBlobData) > 0 { ... }
            if !self.mantle_format_failed && all_blobs_valid && !tx_whole_blob_data.is_empty() {
                if let Some(rlp_frames) = Self::decode_mantle_rlp_frames(&tx_whole_blob_data) {
                    for bytes in rlp_frames {
                        result_data.push(BlobData { data: Some(bytes), calldata: None });
                    }
                    data_index += tx_blob_count;
                    continue;
                }
                // Mantle RLP decode failed — fire the toggle unconditionally,
                // matching Go's `ds.blobToggle()` call.
                self.mantle_format_failed = true;
                warn!(
                    target: "mantle_blob_source",
                    "Mantle format decode failed, falling back to standard blob format"
                );
            }
            // Fallback: return each valid blob's data individually (standard format).
            result_data.extend(fallback_result);

            data_index += tx_blob_count;
        }

        self.open = true;
        self.data = result_data;
        Ok(())
    }

    /// Extracts the next data from the source.
    fn next_data(&mut self) -> PipelineResult<BlobData> {
        if self.data.is_empty() {
            return Err(PipelineError::Eof.temp());
        }

        Ok(self.data.remove(0))
    }
}

#[async_trait]
impl<F, B> DataAvailabilityProvider for MantleBlobSource<F, B>
where
    F: ChainProvider + Sync + Send,
    B: BlobProvider + Sync + Send,
{
    type Item = Bytes;

    async fn next(
        &mut self,
        block_ref: &BlockInfo,
        batcher_address: Address,
    ) -> PipelineResult<Self::Item> {
        self.load_blobs(block_ref, batcher_address).await?;

        let next_data = self.next_data()?;
        if let Some(c) = next_data.calldata {
            return Ok(c);
        }

        // Return the already decoded frame data directly
        // (it was RLP decoded in load_blobs)
        next_data.data.ok_or_else(|| PipelineError::Eof.temp())
    }

    fn clear(&mut self) {
        self.data.clear();
        self.open = false;
    }

    fn reset(&mut self) {
        self.clear();
        self.mantle_format_failed = false;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{
        errors::PipelineErrorKind,
        sources::blob_data::BlobData,
        test_utils::{TestBlobProvider, TestChainProvider},
    };
    use alloc::vec;
    use alloy_eips::eip4844::Blob;
    use alloy_primitives::{B256, address, b256, hex};
    use alloy_rlp::{Decodable, Encodable};

    fn default_test_mantle_blob_source() -> MantleBlobSource<TestChainProvider, TestBlobProvider> {
        let chain_provider = TestChainProvider::default();
        let blob_fetcher = TestBlobProvider::default();
        let batcher_address = Address::default();
        MantleBlobSource::new(chain_provider, blob_fetcher, batcher_address)
    }

    /// https://sepolia.etherscan.io/tx/0x468f0d2b209ad680e147f093f430ffa31f453a14a183d248267b3aaa21a624da
    fn valid_mantle_blob_tx() -> (TxEnvelope, Address, Address, [B256; 3]) {
        let raw_tx_hex = "0x03f8d783aa36a7822aea830f4240830f427082520894ffeeddccbbaa00000000000000000000000000008080c0843b9aca00f863a001d402363affae0d61efd3811cfa5d482e2d3700f20ce1a7934add3c6795f2dea0019f084796dbf6a7ba47b2c1a50eff0c1e32baad480bf5e57b89eafeb418dd76a0011f403cfe025351f08e1fb5b31651012a53c247daf632d3ca51bbd10f82cf2601a0ee6af3b8596947879b181f00119f50fc88cdcf935720e200d81f240a61913c32a02098664ffc7b20db92a6b842ecf3dc3ec4efd232f59896d82939fc97bef363c5";
        let raw_tx_bytes = hex::decode(raw_tx_hex.strip_prefix("0x").unwrap()).unwrap();
        let tx = TxEnvelope::decode(&mut raw_tx_bytes.as_slice()).unwrap();

        let batcher_address = address!("0xFFEEDDCcBbAA0000000000000000000000000000");
        let signer = tx.recover_signer().expect("Should recover signer from Mantle tx");

        let blob_hashes = [
            b256!("01d402363affae0d61efd3811cfa5d482e2d3700f20ce1a7934add3c6795f2de"),
            b256!("019f084796dbf6a7ba47b2c1a50eff0c1e32baad480bf5e57b89eafeb418dd76"),
            b256!("011f403cfe025351f08e1fb5b31651012a53c247daf632d3ca51bbd10f82cf26"),
        ];

        (tx, batcher_address, signer, blob_hashes)
    }

    #[tokio::test]
    async fn test_load_blobs_open() {
        let mut source = default_test_mantle_blob_source();
        source.open = true;
        assert!(source.load_blobs(&BlockInfo::default(), Address::ZERO).await.is_ok());
    }

    #[tokio::test]
    async fn test_load_blobs_chain_provider_err() {
        let mut source = default_test_mantle_blob_source();
        assert!(matches!(
            source.load_blobs(&BlockInfo::default(), Address::ZERO).await,
            Err(BlobProviderError::Backend(_))
        ));
    }

    #[tokio::test]
    async fn test_load_blobs_empty_txs() {
        let mut source = default_test_mantle_blob_source();
        let block_info = BlockInfo::default();
        source.chain_provider.insert_block_with_transactions(0, block_info, Vec::new());
        assert!(!source.open);
        assert!(source.load_blobs(&BlockInfo::default(), Address::ZERO).await.is_ok());
        assert!(source.data.is_empty());
        assert!(source.open);
    }

    #[tokio::test]
    async fn test_next_empty_data_eof() {
        let mut source = default_test_mantle_blob_source();
        source.open = true;

        let err = source.next(&BlockInfo::default(), Address::ZERO).await.unwrap_err();
        assert!(matches!(err, PipelineErrorKind::Temporary(PipelineError::Eof)));
    }

    #[tokio::test]
    async fn test_next_calldata() {
        let mut source = default_test_mantle_blob_source();
        source.open = true;
        source
            .data
            .push(BlobData { data: None, calldata: Some(Bytes::from(vec![0x01, 0x02, 0x03])) });

        let data = source.next(&BlockInfo::default(), Address::ZERO).await.unwrap();
        assert_eq!(data, Bytes::from(vec![0x01, 0x02, 0x03]));
    }

    #[tokio::test]
    async fn test_verify_blob_data_rlp_decode() {
        // Verify RLP encoding/decoding of Vec<Bytes> (Mantle's blob format)
        // This test verifies the RLP structure used by Mantle
        let test_batches =
            vec![Bytes::from(vec![0x00, 0x01, 0x02]), Bytes::from(vec![0x03, 0x04, 0x05])];

        // RLP encode
        let mut rlp_encoded = Vec::new();
        test_batches.encode(&mut rlp_encoded);

        // RLP decode
        let mut rlp_slice = rlp_encoded.as_slice();
        let decoded: Vec<Bytes> = Decodable::decode(&mut rlp_slice).unwrap();

        assert_eq!(decoded.len(), 2);
        assert_eq!(decoded[0], Bytes::from(vec![0x00, 0x01, 0x02]));
        assert_eq!(decoded[1], Bytes::from(vec![0x03, 0x04, 0x05]));
    }

    #[tokio::test]
    async fn test_clear() {
        let mut source = default_test_mantle_blob_source();
        source.open = true;
        source.data.push(BlobData { data: None, calldata: Some(Bytes::from(vec![0x01, 0x02])) });

        source.clear();

        assert!(!source.open, "Source should be closed after clear");
        assert!(source.data.is_empty(), "Data should be empty after clear");
    }

    #[tokio::test]
    async fn test_multiple_next_calls() {
        let mut source = default_test_mantle_blob_source();
        source.open = true;
        source.data.push(BlobData { data: None, calldata: Some(Bytes::from(vec![0x01])) });
        source.data.push(BlobData { data: None, calldata: Some(Bytes::from(vec![0x02])) });
        source.data.push(BlobData { data: None, calldata: Some(Bytes::from(vec![0x03])) });

        let data1 = source.next(&BlockInfo::default(), Address::ZERO).await.unwrap();
        assert_eq!(data1, Bytes::from(vec![0x01]));

        let data2 = source.next(&BlockInfo::default(), Address::ZERO).await.unwrap();
        assert_eq!(data2, Bytes::from(vec![0x02]));

        let data3 = source.next(&BlockInfo::default(), Address::ZERO).await.unwrap();
        assert_eq!(data3, Bytes::from(vec![0x03]));

        // Should return EOF after all data consumed
        let err = source.next(&BlockInfo::default(), Address::ZERO).await.unwrap_err();
        assert!(matches!(err, PipelineErrorKind::Temporary(PipelineError::Eof)));
    }

    #[tokio::test]
    async fn test_wrong_batcher_address() {
        let mut source = default_test_mantle_blob_source();
        let correct_batcher = address!("0xFFEEDDCcBbAA0000000000000000000000000000");
        source.batcher_address = correct_batcher;

        let block_info = BlockInfo::default();
        source.chain_provider.insert_block_with_transactions(0, block_info, Vec::new());

        let result = source.load_blobs(&BlockInfo::default(), correct_batcher).await;
        assert!(result.is_ok());
        assert!(source.data.is_empty(), "No data should be extracted from wrong batcher");
    }

    #[tokio::test]
    async fn test_mantle_valid_blob_decode() {
        let (tx, batcher_address, signer, blob_hashes) = valid_mantle_blob_tx();

        let mut source = default_test_mantle_blob_source();
        source.batcher_address = batcher_address;

        let blob_hexes = [
            include_str!("testdata/mantle_sepolia_block_10001504_blob_0.hex"),
            include_str!("testdata/mantle_sepolia_block_10001504_blob_1.hex"),
            include_str!("testdata/mantle_sepolia_block_10001504_blob_2.hex"),
        ];

        for (i, blob_hex) in blob_hexes.iter().enumerate() {
            let blob_bytes =
                hex::decode(blob_hex.trim().strip_prefix("0x").unwrap_or(blob_hex.trim())).unwrap();
            assert_eq!(blob_bytes.len(), 131072, "Each blob should be 131072 bytes");
            let blob = Blob::try_from(blob_bytes.as_slice()).unwrap();
            source.blob_fetcher.insert_blob(blob_hashes[i], blob);
        }

        let block_info = BlockInfo::default();
        source.chain_provider.insert_block_with_transactions(1, block_info, vec![tx]);

        source.load_blobs(&BlockInfo::default(), signer).await.unwrap();

        assert!(source.open, "Source should be open after load_blobs");
        assert!(!source.data.is_empty(), "Should have decoded frames from Mantle blobs");

        for (i, blob_data) in source.data.iter().enumerate() {
            assert!(blob_data.data.is_some(), "Frame {} should have data", i);
            let data = blob_data.data.as_ref().unwrap();
            assert!(!data.is_empty(), "Frame {} data should not be empty", i);
        }
    }

    #[tokio::test]
    async fn test_rlp_decode_valid_mantle_format() {
        // Verify the exact RLP format used by Mantle matches our implementation
        // Create sample frames as Mantle does
        let frame1 = Bytes::from(vec![0x00, 0x01, 0x02, 0x03]);
        let frame2 = Bytes::from(vec![0x04, 0x05, 0x06]);
        let frames = vec![frame1.clone(), frame2.clone()];

        // RLP encode (this is what Mantle does before splitting into blobs)
        let mut rlp_encoded = Vec::new();
        frames.encode(&mut rlp_encoded);

        // Verify we can decode it back
        let mut rlp_slice = rlp_encoded.as_slice();
        let decoded: Vec<Bytes> = Decodable::decode(&mut rlp_slice).unwrap();

        assert_eq!(decoded.len(), 2);
        assert_eq!(decoded[0], frame1);
        assert_eq!(decoded[1], frame2);

        // Verify the RLP structure
        assert!(rlp_encoded[0] >= 0xc0, "Should start with RLP list marker");
    }

    #[test]
    fn test_decode_mantle_rlp_frames_requires_full_consumption() {
        let frames = vec![Bytes::from(vec![0x01]), Bytes::from(vec![0x02])];
        let mut encoded = Vec::new();
        frames.encode(&mut encoded);

        let decoded =
            MantleBlobSource::<TestChainProvider, TestBlobProvider>::decode_mantle_rlp_frames(
                &encoded,
            )
            .unwrap();
        assert_eq!(decoded, frames);

        let mut with_trailing = encoded;
        with_trailing.extend_from_slice(&[0xff, 0xee]);
        let decoded_with_trailing =
            MantleBlobSource::<TestChainProvider, TestBlobProvider>::decode_mantle_rlp_frames(
                &with_trailing,
            );
        assert!(
            decoded_with_trailing.is_none(),
            "RLP payload with trailing bytes must be rejected"
        );
    }

    #[tokio::test]
    async fn test_parse_blob_count_from_transaction() {
        let (tx, _, _, expected_hashes) = valid_mantle_blob_tx();

        // Extract blob_versioned_hashes from EIP-4844 transaction
        let (blob_hashes, tx_type) = match &tx {
            TxEnvelope::Eip4844(wrapper) => {
                let hashes = match wrapper.tx() {
                    TxEip4844Variant::TxEip4844(tx) => &tx.blob_versioned_hashes,
                    TxEip4844Variant::TxEip4844WithSidecar(tx) => &tx.tx().blob_versioned_hashes,
                };
                (hashes, "EIP-4844")
            }
            TxEnvelope::Eip1559(_) => panic!("Not a blob transaction"),
            _ => panic!("Unexpected transaction type"),
        };

        // Verify blob count
        assert_eq!(blob_hashes.len(), 3, "Should have 3 blob hashes");
        assert_eq!(tx_type, "EIP-4844", "Should be EIP-4844 transaction");

        // Verify the actual blob hashes
        for (i, hash) in blob_hashes.iter().enumerate() {
            let hash_str = format!("0x{:x}", hash);
            let expected_str = format!("0x{:x}", expected_hashes[i]);
            assert_eq!(hash_str, expected_str, "Blob hash {} mismatch", i);
        }
    }

    #[tokio::test]
    async fn test_signer_verification() {
        let (tx, batcher_address, actual_signer, _) = valid_mantle_blob_tx();

        let mut source = default_test_mantle_blob_source();
        source.batcher_address = batcher_address;

        // Test 1: Transaction with correct signer should be accepted
        let (data, hashes, tx_counts) = source.extract_blob_data(vec![tx.clone()], actual_signer);
        assert!(!data.is_empty(), "Should extract data when signer matches (Mantle has 3 blobs)");
        assert_eq!(hashes.len(), 3, "Should extract 3 blob hashes from Mantle transaction");
        assert_eq!(tx_counts, vec![3], "Should have 1 transaction with 3 blobs");

        // Test 2: Transaction with wrong signer should be rejected
        let wrong_signer = Address::ZERO;
        let (data2, hashes2, tx_counts2) = source.extract_blob_data(vec![tx], wrong_signer);
        assert!(data2.is_empty(), "Should not extract data when signer doesn't match");
        assert!(hashes2.is_empty(), "Should not extract blob hashes when signer doesn't match");
        assert!(tx_counts2.is_empty(), "Should have no transactions extracted");
    }

    /// Helper function: decode blob data and check if it's RLP encoded
    fn decode_blob_and_check_rlp(blob_bytes: &[u8]) -> (Bytes, bool, usize) {
        use crate::sources::blob_data::BlobData as BD;

        let mut blob_data = BD::default();
        let blob = Box::new(Blob::try_from(blob_bytes).unwrap());
        blob_data.fill(&[blob], 0).unwrap();
        let decoded = blob_data.decode().unwrap();

        let mut rlp_slice = &decoded[..];
        match VecOfBytes::decode(&mut rlp_slice) {
            Ok(vec) => (decoded, true, vec.0.len()),
            Err(_) => (decoded, false, 0),
        }
    }

    /// Helper function: parse transaction from hex string
    fn parse_tx_from_hex(hex_str: &str) -> TxEnvelope {
        let bytes = hex::decode(hex_str.strip_prefix("0x").unwrap()).unwrap();
        TxEnvelope::decode(&mut bytes.as_slice()).unwrap()
    }

    /// Helper function: load blob from hex file and insert into provider
    fn load_blob_from_hex(
        hex_content: &str,
        blob_hash: B256,
        provider: &mut TestBlobProvider,
    ) -> Vec<u8> {
        let blob_bytes = hex::decode(hex_content.trim().strip_prefix("0x").unwrap_or(hex_content.trim())).unwrap();
        assert_eq!(blob_bytes.len(), 131072, "Blob should be 131072 bytes");
        let blob = Blob::try_from(blob_bytes.as_slice()).unwrap();
        provider.insert_blob(blob_hash, blob);
        blob_bytes
    }

    #[tokio::test]
    async fn test_load_blobs_with_mantle_and_op_blob() {
        // Parse transactions (Mantle blob tx and OP blob tx)
        // https://sepolia.etherscan.io/tx/0x59f091996b73b63989bb0fb8e1e9bf099e4197f3aee91fb7d3f4e183e49ab983
        let mantle_tx = parse_tx_from_hex(
            "0x03f89483aa36a70b8407b5fd4b8509020d9db982520894ffeeddccbbaa00000000000000000000000000008080c0848ac83386e1a001a5e6832cc5b2d89a9dd8ca09ccbdfa9f41a83f8ee4c0a8ca6b63ee693f9fb580a04c7450151bca6b9731fed99fe2ef526fbee62d03cbf153ad4f8fe12ee42c5d719f6a4273bda34e0236f13ce0d516bee9003fdc2ede2642aebe26c9a6b2c1a2f9"
        );
        // https://sepolia.etherscan.io/tx/0x1d9574cc7efa12cf7d3d5a8e0ee1078902158e1aa25d079def7fe65413f51d1b
        let op_tx = parse_tx_from_hex(
            "0x03f89683aa36a7820c99843b9aca0084ba3580c682520894ffeeddccbbaa00000000000000000000000000008080c0843b9aca00e1a001a1de70ef5f8e5f451d2b054df35767bcfe7c1a5d58616ce58742ee9f968dc101a02b985d5ff8834908927adeddeecf37332312e92e25ad913a0fb7aa235b68b49da02c2fd63671226bdf1a6edfc622ef1b5c2587e988cd12a8c44cd6c1fad1ed46ac"
        );

        let batcher_address = address!("0xFFEEDDCcBbAA0000000000000000000000000000");
        let signer = address!("0x008424f79C72a81fE32bf09b0D8A10F2617A5B57");

        let mut source = default_test_mantle_blob_source();
        source.batcher_address = batcher_address;

        // Load blobs and verify their encoding format
        let mantle_blob_hash = b256!("0x01a5e6832cc5b2d89a9dd8ca09ccbdfa9f41a83f8ee4c0a8ca6b63ee693f9fb5");
        let mantle_blob_bytes = load_blob_from_hex(
            include_str!("testdata/mantle_sepolia_mantle_blob.hex"),
            mantle_blob_hash,
            &mut source.blob_fetcher,
        );

        let op_blob_hash = b256!("0x01a1de70ef5f8e5f451d2b054df35767bcfe7c1a5d58616ce58742ee9f968dc1");
        load_blob_from_hex(
            include_str!("testdata/mantle_sepolia_blob.hex"),
            op_blob_hash,
            &mut source.blob_fetcher,
        );

        // Verify encoding formats: Mantle should be RLP encoded, OP should be standard format
        let (_, mantle_is_rlp, mantle_frame_count) = decode_blob_and_check_rlp(&mantle_blob_bytes);
        assert!(mantle_is_rlp, "Mantle blob should be RLP encoded");
        assert!(mantle_frame_count >= 1, "Mantle blob should contain at least 1 frame");

        // Create block with both transactions and load blobs
        let block_info = BlockInfo::default();
        source.chain_provider.insert_block_with_transactions(1, block_info, vec![mantle_tx, op_tx]);
        source.load_blobs(&BlockInfo::default(), signer).await.unwrap();

        // Verify extraction results
        assert!(source.open, "Source should be open after load_blobs");
        assert!(!source.data.is_empty(), "Should have extracted data from both transactions");

        let extracted_frames = source.data.len();
        let expected_min_frames = mantle_frame_count + 1;

        assert!(
            source.data.iter().all(|bd| bd.data.is_some() && !bd.data.as_ref().unwrap().is_empty()),
            "All frames should have non-empty data"
        );

        assert!(
            extracted_frames >= expected_min_frames,
            "Expected at least {} frames, got {}",
            expected_min_frames,
            extracted_frames
        );
    }

    #[tokio::test]
    async fn test_load_blobs_skips_malformed_blob_decode() {
        use crate::sources::blobs::tests::valid_blob_txs;
        use alloy_consensus::{Blob, transaction::SignerRecoverable};

        let txs = valid_blob_txs();
        let signer = txs[0].recover_signer().unwrap_or_default();

        let mut source = default_test_mantle_blob_source();
        source.batcher_address = address!("11E9CA82A3a762b4B5bd264d4173a242e7a77064");

        let block_info = BlockInfo::default();
        source.chain_provider.insert_block_with_transactions(1, block_info, txs);

        // These blobs are intentionally malformed for OP blob decoding and should be skipped
        // without failing the full transaction/block processing path.
        let hashes = [
            b256!("012ec3d6f66766bedb002a190126b3549fce0047de0d4c25cffce0dc1c57921a"),
            b256!("0152d8e24762ff22b1cfd9f8c0683786a7ca63ba49973818b3d1e9512cd2cec4"),
            b256!("013b98c6c83e066d5b14af2b85199e3d4fc7d1e778dd53130d180f5077e2d1c7"),
            b256!("01148b495d6e859114e670ca54fb6e2657f0cbae5b08063605093a4b3dc9f8f1"),
            b256!("011ac212f13c5dff2b2c6b600a79635103d6f580a4221079951181b25c7e6549"),
        ];
        for hash in hashes {
            source.blob_fetcher.insert_blob(hash, Blob::with_last_byte(1u8));
        }

        source.load_blobs(&BlockInfo::default(), signer).await.unwrap();
        assert!(source.open);
        assert!(source.data.is_empty(), "Malformed blobs should be skipped, not hard-fail");
    }

    /// Test that once Mantle RLP decode fails, the result is cached and subsequent blocks
    /// skip the Mantle RLP attempt entirely, matching Go's `blobSourceChanged` toggle behavior.
    #[tokio::test]
    async fn test_mantle_format_failed_cached_across_blocks() {
        use crate::sources::blobs::tests::valid_blob_txs;
        use alloy_consensus::{Blob, transaction::SignerRecoverable};

        let txs = valid_blob_txs();
        let signer = txs[0].recover_signer().unwrap_or_default();

        let mut source = default_test_mantle_blob_source();
        source.batcher_address = address!("11E9CA82A3a762b4B5bd264d4173a242e7a77064");

        let block_info = BlockInfo::default();
        source.chain_provider.insert_block_with_transactions(1, block_info, txs.clone());

        let blob_hashes = [
            b256!("012ec3d6f66766bedb002a190126b3549fce0047de0d4c25cffce0dc1c57921a"),
            b256!("0152d8e24762ff22b1cfd9f8c0683786a7ca63ba49973818b3d1e9512cd2cec4"),
            b256!("013b98c6c83e066d5b14af2b85199e3d4fc7d1e778dd53130d180f5077e2d1c7"),
            b256!("01148b495d6e859114e670ca54fb6e2657f0cbae5b08063605093a4b3dc9f8f1"),
            b256!("011ac212f13c5dff2b2c6b600a79635103d6f580a4221079951181b25c7e6549"),
        ];

        // Use standard OP-format blobs (not Mantle RLP format)
        let minimal_blob = {
            let mut data = vec![0u8; 131072];
            data[0] = crate::sources::blob_data::BLOB_ENCODING_VERSION;
            data[2] = 0;
            data[3] = 0;
            data[4] = 1;
            Blob::try_from(data.as_slice()).unwrap()
        };
        for hash in blob_hashes {
            source.blob_fetcher.insert_blob(hash, minimal_blob);
        }

        // Block 1: mantle_format_failed starts false, RLP decode fails, toggle fires
        assert!(!source.mantle_format_failed);
        source.load_blobs(&BlockInfo::default(), signer).await.unwrap();
        assert!(source.mantle_format_failed, "Toggle should fire after Mantle RLP decode failure");

        // Simulate next block: clear data but mantle_format_failed persists
        source.clear();
        assert!(source.mantle_format_failed, "Toggle must persist across clear()");

        // Block 2: re-insert same block data
        source.chain_provider.insert_block_with_transactions(2, block_info, txs);
        for hash in blob_hashes {
            source.blob_fetcher.insert_blob(hash, minimal_blob);
        }

        // Should succeed and skip Mantle RLP decode attempt entirely
        source.load_blobs(&BlockInfo::default(), signer).await.unwrap();
        assert!(source.open);
        assert!(source.mantle_format_failed, "Toggle should remain true");
    }

    #[tokio::test]
    async fn test_mantle_format_failed_reset_on_pipeline_reset() {
        let chain_provider = TestChainProvider::default();
        let blob_fetcher = TestBlobProvider::default();
        let signer = Address::default();

        let mut source = MantleBlobSource::new(chain_provider, blob_fetcher, signer);

        // Simulate toggle being fired
        source.mantle_format_failed = true;
        assert!(source.mantle_format_failed);

        // clear() should NOT reset the toggle
        source.clear();
        assert!(source.mantle_format_failed, "clear() must not reset toggle");

        // reset() SHOULD reset the toggle (pipeline reset / L1 reorg)
        source.reset();
        assert!(!source.mantle_format_failed, "reset() must clear toggle for pipeline reset");
    }
}
