//! Contains the [MantleEthereumDataSource], which is a concrete implementation of the
//! [DataAvailabilityProvider] trait for the Ethereum protocol with Mantle blob support.
//!
//! This data source handles blob decoding with fallback logic:
//! - First tries MantleBlobSource (Mantle blob decoding)
//! - Falls back to BlobSource (standard blob decoding) if MantleBlobSource fails

use super::MantleBlobSource;
use crate::{
    BlobProvider, CalldataSource, ChainProvider, DataAvailabilityProvider, PipelineResult,
};
use alloc::{boxed::Box, fmt::Debug};
use alloy_primitives::{Address, Bytes};
use async_trait::async_trait;
use kona_genesis::RollupConfig;
use kona_protocol::BlockInfo;

/// A factory for creating an Ethereum data source provider with Mantle Arsia hardfork support.
#[derive(Debug, Clone)]
pub struct MantleEthereumDataSource<C, B>
where
    C: ChainProvider + Send + Clone,
    B: BlobProvider + Send + Clone,
{
    /// The ecotone timestamp.
    pub ecotone_timestamp: Option<u64>,
    /// The Mantle Arsia timestamp.
    pub mantle_arsia_timestamp: Option<u64>,
    /// The Mantle blob source
    pub mantle_blob_source: MantleBlobSource<C, B>,
    /// The calldata source.
    pub calldata_source: CalldataSource<C>,
}

impl<C, B> MantleEthereumDataSource<C, B>
where
    C: ChainProvider + Send + Clone + Debug,
    B: BlobProvider + Send + Clone + Debug,
{
    /// Instantiates a new [`MantleEthereumDataSource`].
    pub const fn new(
        mantle_blob_source: MantleBlobSource<C, B>,
        calldata_source: CalldataSource<C>,
        cfg: &RollupConfig,
    ) -> Self {
        Self {
            ecotone_timestamp: cfg.hardforks.ecotone_time,
            mantle_arsia_timestamp: cfg.mantle_hardforks.mantle_arsia_time,
            mantle_blob_source,
            calldata_source,
        }
    }

    /// Instantiates a new [`MantleEthereumDataSource`] from parts.
    pub fn new_from_parts(provider: C, blobs: B, cfg: &RollupConfig) -> Self {
        Self {
            ecotone_timestamp: cfg.hardforks.ecotone_time,
            mantle_arsia_timestamp: cfg.mantle_hardforks.mantle_arsia_time,
            mantle_blob_source: MantleBlobSource::new(
                provider.clone(),
                blobs.clone(),
                cfg.batch_inbox_address,
            ),
            calldata_source: CalldataSource::new(provider, cfg.batch_inbox_address),
        }
    }
}

#[async_trait]
impl<C, B> DataAvailabilityProvider for MantleEthereumDataSource<C, B>
where
    C: ChainProvider + Send + Sync + Clone + Debug,
    B: BlobProvider + Send + Sync + Clone + Debug,
{
    type Item = Bytes;

    async fn next(
        &mut self,
        block_ref: &BlockInfo,
        batcher_address: Address,
    ) -> PipelineResult<Self::Item> {
        self.mantle_blob_source.next(block_ref, batcher_address).await
    }

    fn clear(&mut self) {
        self.mantle_blob_source.clear();
        self.calldata_source.clear();
    }

    fn reset(&mut self) {
        self.mantle_blob_source.reset();
        self.calldata_source.clear();
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{
        sources::blob_data::BLOB_ENCODING_VERSION,
        test_utils::{TestBlobProvider, TestChainProvider},
    };
    use alloc::vec;
    use alloy_consensus::TxEnvelope;
    use alloy_eips::{eip2718::Decodable2718, eip4844::BYTES_PER_BLOB};
    use alloy_primitives::{Address, Bytes, address};
    use kona_genesis::{RollupConfig, SystemConfig};
    use kona_protocol::BlockInfo;

    /// Minimal standard OP blob encoding that decodes to a single 0x00 byte.
    /// Standard encoding: version byte at [0], length (1 byte) at [2..5], rest zeros.
    /// Decode output is 0x00 because the blob payload is all zeros (see blob_data::decode).
    fn minimal_standard_blob_bytes() -> Bytes {
        let mut data = vec![0u8; BYTES_PER_BLOB];
        data[0] = BLOB_ENCODING_VERSION;
        data[2] = 0;
        data[3] = 0;
        data[4] = 1;
        Bytes::from(data)
    }

    #[tokio::test]
    async fn test_clear_mantle_ethereum_data_source() {
        let chain = TestChainProvider::default();
        let blob = TestBlobProvider::default();
        let cfg = RollupConfig::default();
        let mut calldata = CalldataSource::new(chain.clone(), Address::ZERO);
        calldata.calldata.insert(0, Default::default());
        calldata.open = true;
        let mut mantle_blob = MantleBlobSource::new(chain.clone(), blob, Address::ZERO);
        mantle_blob.data = vec![Default::default()];
        mantle_blob.open = true;
        let mut data_source = MantleEthereumDataSource::new(mantle_blob, calldata, &cfg);

        data_source.clear();
        assert!(data_source.mantle_blob_source.data.is_empty());
        assert!(!data_source.mantle_blob_source.open);
        assert!(data_source.calldata_source.calldata.is_empty());
        assert!(!data_source.calldata_source.open);
    }

    /// OP-style blob tx + standard-encoded blobs. Shared setup for mantle-fails / fallback-succeeds
    /// tests.
    fn op_style_blob_test_setup() -> (BlockInfo, Address, TestChainProvider, TestBlobProvider) {
        use alloy_consensus::{Blob, transaction::SignerRecoverable};

        use crate::sources::blobs::tests::valid_blob_txs;

        let txs = valid_blob_txs();
        let batcher_address = txs[0].recover_signer().unwrap_or_default();
        let block_ref = BlockInfo::default();
        let mut chain = TestChainProvider::default();
        chain.insert_block_with_transactions(1, block_ref, txs);

        let blob_hashes = [
            alloy_primitives::b256!(
                "012ec3d6f66766bedb002a190126b3549fce0047de0d4c25cffce0dc1c57921a"
            ),
            alloy_primitives::b256!(
                "0152d8e24762ff22b1cfd9f8c0683786a7ca63ba49973818b3d1e9512cd2cec4"
            ),
            alloy_primitives::b256!(
                "013b98c6c83e066d5b14af2b85199e3d4fc7d1e778dd53130d180f5077e2d1c7"
            ),
            alloy_primitives::b256!(
                "01148b495d6e859114e670ca54fb6e2657f0cbae5b08063605093a4b3dc9f8f1"
            ),
            alloy_primitives::b256!(
                "011ac212f13c5dff2b2c6b600a79635103d6f580a4221079951181b25c7e6549"
            ),
        ];
        let standard_blob = Blob::try_from(minimal_standard_blob_bytes().as_ref()).unwrap();
        let mut blob_provider = TestBlobProvider::default();
        for hash in blob_hashes {
            blob_provider.insert_blob(hash, standard_blob);
        }
        (block_ref, batcher_address, chain, blob_provider)
    }

    /// Integration test: block with standard OP-format blob tx.
    /// mantle_blob_source fails (decoded frame bytes are not RLP list), blob_source decodes and
    /// returns the first frame (one byte 0x00 from minimal_standard_blob_bytes).
    /// MantleEthereumDataSource returns that frame via fallback.
    #[tokio::test]
    async fn test_op_style_blob_mantle_fails_blob_source_succeeds() {
        let (block_ref, batcher_address, chain, blob_provider) = op_style_blob_test_setup();
        let cfg = RollupConfig {
            batch_inbox_address: address!("11E9CA82A3a762b4B5bd264d4173a242e7a77064"),
            ..Default::default()
        };

        let mut data_source = MantleEthereumDataSource::new_from_parts(chain, blob_provider, &cfg);
        // mantle_blob_source tries RLP decode on concatenated decoded frames (5 bytes 0x00) ->
        // fails blob_source decodes first blob -> returns 1 byte 0x00
        // (minimal_standard_blob_bytes decodes to 0x00)
        let data = data_source.next(&block_ref, batcher_address).await.unwrap();
        assert_eq!(data, Bytes::from([0]));
    }

    #[tokio::test]
    async fn test_open_calldata_source_pre_ecotone() {
        let mut chain = TestChainProvider::default();
        let blob = TestBlobProvider::default();
        let batcher_address = address!("6887246668a3b87F54DeB3b94Ba47a6f63F32985");
        let batch_inbox = address!("FF00000000000000000000000000000000000010");
        let block_ref = BlockInfo { number: 10, ..Default::default() };

        let mut cfg = RollupConfig::default();
        cfg.genesis.system_config = Some(SystemConfig { batcher_address, ..Default::default() });
        cfg.batch_inbox_address = batch_inbox;

        // load a test batcher transaction
        let raw_batcher_tx = include_bytes!("../../testdata/raw_batcher_tx.hex");
        let tx = TxEnvelope::decode_2718(&mut raw_batcher_tx.as_ref()).unwrap();
        chain.insert_block_with_transactions(10, block_ref, vec![tx]);

        // Should successfully retrieve a calldata batch from the block (before ecotone)
        let mut data_source = MantleEthereumDataSource::new_from_parts(chain, blob, &cfg);
        let calldata_batch = data_source.next(&block_ref, batcher_address).await.unwrap();
        assert_eq!(calldata_batch.len(), 119823);
    }
}
