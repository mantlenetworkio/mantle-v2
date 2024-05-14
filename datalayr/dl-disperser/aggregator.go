package disperser

import (
	"context"
	"encoding/hex"
	"log"

	"math/big"
	"time"

	"github.com/Layr-Labs/datalayr/common/contracts"
	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/common/middleware/correlation"
	"github.com/Layr-Labs/datalayr/common/middleware/logger"

	"github.com/Layr-Labs/datalayr/common/graphView"

	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DisperseStatus struct {
	Sigs map[int][]byte
}

type SendResult struct {
	Err        error
	Sig        []byte
	FrameIndex uint32
	Address    common.Address
}

type Aggregator struct {
	ChainClient *contracts.DataLayrChainClient
	Timeout     time.Duration
	Logger      *logging.Logger
}

func NewAggregator(
	chainClient *contracts.DataLayrChainClient,
	timeout time.Duration,
	logger *logging.Logger,
) *Aggregator {
	return &Aggregator{
		ChainClient: chainClient,
		Timeout:     timeout,
		Logger:      logger,
	}
}

type AggregateSigs struct {
	StoredAggPubkeyG1 *bn254.G1Affine // the agg pubkey stored on chain, a G1 point
	UsedAggPubkeyG2   *bn254.G2Affine // the G2 agg pubkey to be used to verify against
	NonSignerPubkeys  []bn254.G1Affine
	AggSig            *bn254.G1Affine
}

// TODO deal with case node/symbol is greater than limit
func (a *Aggregator) Aggregate(ctx context.Context, store *Store, stateView *graphView.StateView) (*AggregateSigs, error) {
	log := a.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering Aggregate function...")
	defer log.Trace().Msg("Exiting Aggregate function...")

	update := make(chan SendResult, len(stateView.Registrants))

	// Disperse
	err := a.DisperseStore(ctx, store, stateView, update)
	if err != nil {
		return nil, err
	}

	// Aggregate
	nonPubKeys, aggSig, aggPub, err := a.AggregateSignatures(ctx, store, update, stateView)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Str("MSGHASH", hex.EncodeToString(store.MsgHash[:])).
		Msg("Signatures Aggregated")
	// Verify Aggregate Signature
	ok := bls.VerifyBlsSig(aggSig, aggPub, store.MsgHash[:])
	if !ok {
		// TODO: Return Error
		log.Error().Err(ErrInvalidAggSig).Msg("Cannot verify Agg sig")
	}

	// Todo: Check that db aggPubKey is correct
	aggPubkey := stateView.TotalOperator.AggPubKey
	if err != nil {
		log.Error().Err(err).Msg("Cannot find aggPubKey")
		return nil, err
	}

	aggPubkeyVerCopy := new(bn254.G1Affine).Set(aggPubkey)
	for _, np := range nonPubKeys {
		aggPubkeyVerCopy.Sub(aggPubkeyVerCopy, &np)
	}

	pubkeysEqual, err := bls.CheckG1AndG2DiscreteLogEquality(aggPubkeyVerCopy, aggPub)
	if err != nil {
		log.Error().Err(err).Msg("Err checking pubkey equality")
		return nil, err
	}

	if !pubkeysEqual {
		log.Error().Err(ErrInconsistentAggPub).Msg("Received Bls Pub diffs from substracted")
		return nil, ErrInconsistentAggPub
	}

	log.Debug().
		Uint32("storeId", store.StoreId).
		Int("numNonPubKeys", len(nonPubKeys)).
		Msg("Signatures Aggregated!")

	return &AggregateSigs{
		StoredAggPubkeyG1: aggPubkey,
		UsedAggPubkeyG2:   aggPub,
		NonSignerPubkeys:  nonPubKeys,
		AggSig:            aggSig,
	}, nil
}

func (a *Aggregator) DisperseStore(ctx context.Context, store *Store, stateView *graphView.StateView, update chan SendResult) error {
	log := a.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering DisperseStore function...")
	defer log.Trace().Msg("Exiting DisperseStore function...")

	log.Trace().Msgf("storeId %v", store.StoreId)
	for _, registrant := range stateView.Registrants {
		go func(registrant graphView.RegistrantView) {

			assignment := store.Assignments[registrant.Index]

			frameBytes := make([][]byte, 0)
			for _, ind := range assignment.GetIndices() {

				frameByte, err := store.Chunks[ind].Encode()
				if err != nil {
					update <- SendResult{
						Err:        err,
						Sig:        nil,
						FrameIndex: 0,
						Address:    registrant.Address,
					}
				}

				frameBytes = append(frameBytes, frameByte)
			}

			log.Trace().Msgf("frameIndex %v, registrant index %v\n", assignment.ChunkIndex, registrant.Index)

			sig, err := a.send(
				ctx,
				frameBytes,
				store.MsgHash,
				registrant.Socket,
			)

			update <- SendResult{
				Err:        err,
				Sig:        sig,
				FrameIndex: uint32(assignment.ChunkIndex),
				Address:    registrant.Address,
			}

		}(*registrant)
	}
	return nil
}

func (a *Aggregator) AggregateSignatures(
	ctx context.Context,
	store *Store,
	update chan SendResult,
	stateView *graphView.StateView,
) ([]bn254.G1Affine, *bn254.G1Affine, *bn254.G2Affine, error) {
	log := a.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering AggregateSignatures function...")
	defer log.Trace().Msg("Exiting AggregateSignatures function...")

	stakeSigned := big.NewInt(0)
	mantleStakeSigned := big.NewInt(0)
	numSuccess := 0
	aggPub := new(bn254.G2Affine)
	aggSig := new(bn254.G1Affine)
	signerAddressMap := make(map[common.Address]bool)

	// Aggregate Signatures
	numRegistrant := len(stateView.Registrants)
	for numReply := 0; numReply < numRegistrant; numReply++ {
		select {
		case r := <-update:

			log.Debug().Msgf("Frame index %v", r.FrameIndex)

			func(r SendResult) {
				if r.Err != nil {
					log.Error().Err(r.Err).Msg("Send Failed")

					registrant, found := stateView.RegistrantMap[r.Address]

					if !found {
					} else {
						log.Error().Msgf("Send Failed with %v", registrant.Socket)
					}
					return
				}

				registrant, ok := stateView.RegistrantMap[r.Address]
				if !ok {
					log.Error().Msg("Failed to retrieve pubKey")
					return
				}
				pubkeyG2 := registrant.PubkeyG2

				log.Trace().Msgf("reg %v. recv pubkey", hexutil.Encode(r.Address[:]))

				// Get stake amounts for operator
				quorumIndex := 0
				mantleQuorumIndex := 1
				stake := registrant.QuorumStakes[quorumIndex]
				mantleStake := registrant.QuorumStakes[mantleQuorumIndex]

				// Create signature from bytes
				var sig bn254.G1Affine
				_, err := sig.SetBytes(r.Sig)
				if err != nil {
					log.Error().Err(err).Msg("Cannot set sig bytes")
					return
				}
				// Verify Signature
				ok = bls.VerifyBlsSig(&sig, pubkeyG2, store.MsgHash[:])
				if !ok {
					log.Error().Msgf("Cannot verify bls sig. for %v. HeaderHash %v. Sig %v", pubkeyG2, hexutil.Encode(store.HeaderHash[:]), sig)
					return
				}
				log.Trace().
					Str("MSGHASH", hex.EncodeToString(store.MsgHash[:])).
					Msg("Valid signature received")

				signerAddressMap[r.Address] = true
				stakeSigned.Add(stakeSigned, stake)
				mantleStakeSigned.Add(mantleStakeSigned, mantleStake) //  add for mantle Quorum
				//log.Trace().Str("pubKey", fmt.Sprint(pubkey)).Msg("Verified signature")
				if numSuccess == 0 {
					aggSig = &sig
					aggPub = pubkeyG2
				} else {
					aggSig.Add(aggSig, &sig)
					aggPub.Add(aggPub, pubkeyG2)
				}

				numSuccess += 1

			}(r)
		}
	}

	log.Debug().Int("numSuccess", numSuccess).Msg("Summary of success response")

	//if not met eth threshold amount of stake
	//if stakeSigned.Cmp(store.QuorumThreshold) < 0 {
	//	log.Warn().Int("numSuccess", numSuccess).Str("stakeSigned", stakeSigned.String()).Str("threshold", store.QuorumThreshold.String()).Str("Total stake", stateView.TotalStake.QuorumStakes[0].String()).Msg("ErrInsufficientEthSigs")
	//	return nil, nil, nil, ErrInsufficientEthSigs
	//}

	// if not met mantle threshold amount of stake
	//if mantleStakeSigned.Cmp(store.MantleQuorumThreshold) < 0 {
	//	log.Warn().Int("numSuccess", numSuccess).Str("mantleStakeSigned", mantleStakeSigned.String()).Str("mantleThreshold", store.MantleQuorumThreshold.String()).Str("Mantle total stake", stateView.TotalStake.QuorumStakes[1].String()).Msg("ErrInsufficientEigenSigs")
	//	return nil, nil, nil, ErrInsufficientEigenSigs
	//}

	// Aggregrate Non signer Pubkey Id
	nonPubKeys := make([]bn254.G1Affine, 0)

	for _, registrant := range stateView.Registrants {
		_, found := signerAddressMap[registrant.Address]
		if !found {
			nonPubKeys = append(nonPubKeys, *registrant.PubkeyG1)
			//log.Trace().Msgf("Id %v . non Pub key %v\n", registrant.Id, pubKey)
		}
	}

	return nonPubKeys, aggSig, aggPub, nil
}

func (a *Aggregator) send(ctx context.Context, frames [][]byte, msgHash [32]byte, socket string) ([]byte, error) {
	log := a.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering send function...")
	defer log.Trace().Msg("Exiting send function...")

	// TODO Add secure Grpc
	conn, err := grpc.Dial(
		socket,
		grpc.WithChainUnaryInterceptor(
			correlation.UnaryClientInterceptor(),
			logger.UnaryClientInterceptor(*a.Logger.Logger),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error().Err(err).Msgf("Disperser cannot connect to %v\n", socket)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(ctx, a.Timeout)
	defer cancel()

	request := &pb.StoreFramesRequest{
		MsgHash: msgHash[:],
		Frame:   frames,
	}

	opt := grpc.MaxCallSendMsgSize(1024 * 1024 * 300)
	reply, dlnErr := c.StoreFrames(ctx, request, opt)

	if dlnErr == nil {
		sig := reply.GetSignature()
		return sig, nil
	} else {
		return nil, dlnErr
	}
}

// Returns which byte array has the first larger byte than the other
// a and b must have equal length.
func compareByteArrays(a []byte, b []byte) int {
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			return 1
		} else if a[i] < b[i] {
			return -1
		}
	}
	return 0
}

func bigIntToBytes(n *big.Int, packTo int) []byte {
	bigIntBytes := n.Bytes()
	bigIntLen := len(bigIntBytes)
	intBytes := make([]byte, packTo)

	if bigIntLen > packTo {
		// TODO: Remove fatal
		log.Fatal("Cannot pad bytes: Desired length is less than existing length")
	}

	for i := 0; i < bigIntLen; i++ {
		intBytes[packTo-1-i] = bigIntBytes[bigIntLen-1-i]
	}

	return intBytes
}

func concatCopyPreAllocate(slices [][]byte) []byte {
	out := make([]byte, 0)
	for _, s := range slices {
		out = append(out, s...)
	}

	return out
}

func quickSort(hashed_pks [][]byte, pks [][]byte, left, right int) {
	if left > right {
		return
	}

	pivot := partition(hashed_pks, pks, left, right)
	quickSort(hashed_pks, pks, left, pivot-1)
	quickSort(hashed_pks, pks, pivot+1, right)
}

func partition(hashed_pks [][]byte, pks [][]byte, left, right int) int {
	pivot := hashed_pks[right]
	for i := left; i < right; i++ {
		if compareByteArrays(hashed_pks[i], pivot) < 0 {
			hashed_pks[i], hashed_pks[left] = hashed_pks[left], hashed_pks[i]
			pks[i], pks[left] = pks[left], pks[i]
			left++
		}
	}

	hashed_pks[left], hashed_pks[right] = hashed_pks[right], hashed_pks[left]
	pks[left], pks[right] = pks[right], pks[left]
	return left
}
