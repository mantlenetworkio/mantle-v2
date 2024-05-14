package disperser

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/header"
	"github.com/consensys/gnark-crypto/ecc/bn254"

	"github.com/Layr-Labs/datalayr/common/graphView"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type StoreRequest struct {
	BlockNumber          uint32
	AdvRatioBasisPoints  *big.Int
	LiveRatioBasisPoints *big.Int
	Duration             uint64
	Data                 []byte
}

type Store struct {
	StoreMetadata

	Header      header.DataStoreHeader
	HeaderBytes []byte

	Chunks    []rs.Frame
	TotalSize uint32 // Note: This is likely temporary. We should have information on hand to easily calculate this from number of chunks and chunk size.

	Assignments []encoding.ChunkAssignment
}

type StoreMetadata struct {
	ReferenceBlockNumber        uint32
	TotalOperatorsIndex         uint32
	HeaderHash                  [32]byte
	Duration                    uint8
	MantleFirstQuorumThreshold  *big.Int
	MantleSecnodQuorumThreshold *big.Int
	Fee                         *big.Int
	StoreId                     uint32   // this field cannot be set until initDataStore has been called
	MsgHash                     [32]byte // this field cannot be set until initDataStore has been called
	BlockNumber                 uint64   // this field cannot be set until initDataStore has been called
}

const STORE_META_SIZE = 150

func (s *Store) UpperBoundBytes() uint32 {
	return STORE_META_SIZE + s.TotalSize + uint32(len(s.Assignments)*8)
}

func (d *Disperser) CreateStore(ctx context.Context, req StoreRequest) (*Store, *graphView.StateView, error) {
	log := d.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering CreateStore function...")
	defer log.Trace().Msg("Exiting CreateStore function...")

	if req.BlockNumber == 0 {
		blockNumber, err := d.ChainClient.GetBlockNumber(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Error fetching blocknumber")
			return nil, nil, err
		}
		req.BlockNumber = uint32(blockNumber)
	}

	stateView, err := d.GraphClient.GetStateView(ctx, req.BlockNumber)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get state view")
		return nil, nil, err
	}

	str, err := json.Marshal(stateView)
	if err != nil {
		log.Error().Err(err).Msg("JSON marshal failed")
		return nil, nil, err
	}

	log.Trace().Str("stateView", string(str)).Uint32("blockNumber", req.BlockNumber).Msg("StateView")

	// Check data length
	num := len(stateView.Registrants)
	if num == 0 {
		return nil, nil, ErrNotEnoughParticipants
	}
	err = d.CheckDataLength(len(req.Data), num)
	if err != nil {
		log.Error().Err(err).Msg("Not enough participants!")
		return nil, nil, err
	}

	mantleFirstQuorumIndex := 0
	mantleQuorumParams, err := encoding.GetQuorumParams(req.LiveRatioBasisPoints, req.AdvRatioBasisPoints, stateView, mantleFirstQuorumIndex)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query total operators")
		return nil, nil, err
	}

	log.Debug().
		Uint64("NumSys", mantleQuorumParams.NumSys).
		Uint64("NumPar", mantleQuorumParams.NumPar).
		Str("StakeThreshold", mantleQuorumParams.StakeThreshold.String()).
		Msg("Got quorum params")

	chunks, dsHeader, err := d.Encode(ctx, req.Data, mantleQuorumParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode")
		return nil, nil, err
	}

	headerBytes, headerHash, err := header.CreateUploadHeader(*dsHeader)
	if err != nil {
		log.Error().Err(err).Msg("error creating header")
		return nil, nil, err
	}

	headerHashHex := hexutil.Encode(headerHash[:])
	log.Trace().Msgf("DataStore Encoded with %v", headerHashHex)

	assignments := encoding.GetOperatorAssignments(mantleQuorumParams, headerHash)

	fee, totalSize, err := d.GetDataStoreFee(ctx, chunks, uint8(req.Duration))
	if err != nil {
		return nil, nil, err
	}

	metadata := StoreMetadata{
		ReferenceBlockNumber:        req.BlockNumber,
		TotalOperatorsIndex:         stateView.TotalOperator.Index,
		HeaderHash:                  headerHash,
		Duration:                    uint8(req.Duration),
		MantleFirstQuorumThreshold:  mantleQuorumParams.StakeThreshold,
		MantleSecnodQuorumThreshold: mantleQuorumParams.StakeThreshold,
		Fee:                         fee,
		MsgHash:                     [32]byte{},
		StoreId:                     0,
	}

	return &Store{
		StoreMetadata: metadata,
		Header:        *dsHeader,
		HeaderBytes:   headerBytes,
		Chunks:        chunks,
		TotalSize:     totalSize,
		Assignments:   assignments,
	}, stateView, nil

}

func (d *Disperser) GetDataStoreFee(ctx context.Context, chunks []rs.Frame, duration uint8) (*big.Int, uint32, error) {
	log := d.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering GetDataStoreFee function...")
	defer log.Trace().Msg("Exiting GetDataStoreFee function...")

	// Determine Fee
	feeParams, err := d.ChainClient.GetFeeParams()
	if err != nil {
		log.Error().Err(err).Msg("Error fetching feeParams")
		return nil, 0, err
	}

	totalSize := uint32(0)
	for i := 0; i < len(chunks); i++ {
		chunkBytes, err := chunks[i].Encode()
		if err != nil {
			return nil, 0, err
		}
		totalSize += uint32(len(chunkBytes))
	}

	fee := feeParams.GetPrecommitFee(duration, totalSize, 60*60)

	return fee, totalSize, nil
}

func (e *Disperser) Encode(ctx context.Context, data []byte, params encoding.QuorumParams) ([]rs.Frame, *header.DataStoreHeader, error) {
	log := e.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering Encode function...")
	defer log.Trace().Msg("Exiting Encode function...")

	var disperserAddrByte [20]byte
	copy(disperserAddrByte[:], common.HexToAddress(e.Config.Address).Bytes())

	origDataSize := len(data)
	// ToDo, maybe check size less than uint32, if not checked before

	var kzgCommit, lowDegreeProof []byte
	var chunks []rs.Frame
	var err error

	digest := crypto.Keccak256(data)

	log.Trace().Str("input digest", string(digest)).Str("cache digest", string(e.Cache.digest[:])).Bool("use cache", e.Config.UseCache).Msg("cache info")

	if e.Config.UseCache && e.Cache.CheckHit(digest) {
		log.Trace().Msg("using a cache")
		kzgCommit, lowDegreeProof, chunks = e.Cache.Get()
	} else {
		log.Trace().Msg("skipping cache")

		kzgCommit, lowDegreeProof, chunks, err = e.encodeWrapper(ctx, data, params.NumSys, params.NumPar)

		log.Trace().Msg("finish encoding")

		if err != nil {
			return nil, nil, err
		}

		if e.Config.UseCache {
			e.Cache.Put(digest, kzgCommit, lowDegreeProof, chunks)
		}
	}

	// Check number of frames in each frame matches NumSys+NumPar
	if len(chunks) != int(params.NumSys+params.NumPar) {
		log.Error().
			Uint64("numSym+NumPar", params.NumSys+params.NumPar).
			Int("lenFrames", len(chunks)).
			Msg(" NumSys+NumPar does not match len(frames)")
		return nil, nil, ErrInconsistentChainStateFrame
	}

	var kzgCommit64 [64]byte
	copy(kzgCommit64[:], kzgCommit[:])

	var lowDegreeProof64 [64]byte
	copy(lowDegreeProof64[:], lowDegreeProof[:])

	// Todo: Get this directly from encoding params
	polySize := uint32(len(chunks[len(chunks)-1].Coeffs))

	dataStoreHeader := &header.DataStoreHeader{
		NumSys:         uint32(params.NumSys),
		NumPar:         uint32(params.NumPar),
		KzgCommit:      kzgCommit64,
		LowDegreeProof: lowDegreeProof64,
		Degree:         polySize,
		OrigDataSize:   uint32(origDataSize),
		Disperser:      disperserAddrByte,
	}

	log.Debug().Msgf("kzgCommitArray %v", hexutil.Encode(dataStoreHeader.KzgCommit[:]))
	log.Debug().Msgf("Degree %v", dataStoreHeader.Degree)
	log.Debug().Msgf("NumSys %v", dataStoreHeader.NumSys)
	log.Debug().Msgf("NumPar %v", dataStoreHeader.NumPar)
	log.Debug().Msgf("OrigDataSize %v", dataStoreHeader.OrigDataSize)
	log.Debug().Msgf("Disperser %v", hexutil.Encode(dataStoreHeader.Disperser[:]))
	log.Debug().Msgf("LowDegreeProof %v", hexutil.Encode(dataStoreHeader.LowDegreeProof[:]))

	return chunks, dataStoreHeader, nil
}

func (e *Disperser) encodeWrapper(ctx context.Context, data []byte, numSys, numPar uint64) ([]byte, []byte, []rs.Frame, error) {
	log := e.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering encodeWrapper function...")
	defer log.Trace().Msg("Exiting encodeWrapper function...")

	encoder, err := e.KzgEncoderGroup.NewKzgEncoder(numSys, numPar, uint64(len(data)))
	if err != nil {
		return nil, nil, nil, err
	}

	commit, lowDegreeProof, frames, _, err := encoder.EncodeBytes(ctx, data)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debug().Int("length", len(frames)).Msg("frames")

	castedG1 := *(*bn254.G1Affine)(commit)
	kzgCommit := bls.SerializeG1(&castedG1)
	castedG1LowDegreeProof := *(*bn254.G1Affine)(lowDegreeProof)
	lowDegreeProofBytes := bls.SerializeG1(&castedG1LowDegreeProof)

	return kzgCommit, lowDegreeProofBytes, frames, nil
}
