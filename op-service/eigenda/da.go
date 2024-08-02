package eigenda

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/ethereum-optimism/optimism/op-service/eigenda/verify"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type IEigenDA interface {
	RetrieveBlob(ctx context.Context, BatchHeaderHash []byte, BlobIndex uint32) ([]byte, error)
	DisperseBlob(ctx context.Context, txData []byte) (*disperser.BlobInfo, []byte, error)
	GetBlobStatus(ctx context.Context, requestID []byte) (*disperser.BlobStatusReply, error)
}

type EigenDA struct {
	Config

	Log      log.Logger
	Metricer Metrics
	signer   BlobRequestSigner
	verifier *verify.Verifier
}

func NewEigenDAClient(config Config, log log.Logger, metricer Metrics) (*EigenDA, error) {
	var signer BlobRequestSigner
	var err error
	if config.EnableHsm {
		signer, err = NewHsmBlobSigner(config.HsmCreden, config.HsmAPIName, config.HsmPubkey)
		if err != nil {
			return nil, fmt.Errorf("error creating HSM signer: %w", err)
		}
	} else if config.PrivateKey != "" {
		signer, err = NewLocalBlobSigner(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error creating local signer: %w", err)
		}
	}

	var verifier *verify.Verifier
	if config.EthRPC != "" && config.SvcManagerAddr != "" {
		vCfg := config.VerificationCfg()
		verifier, err = verify.NewVerifier(vCfg, log)
		if err != nil {
			return nil, fmt.Errorf("error creating verifier: %w", err)
		}
	}

	return &EigenDA{
		Config:   config,
		Log:      log,
		signer:   signer,
		verifier: verifier,
		Metricer: metricer,
	}, nil
}

func (m *EigenDA) GetBlobStatus(ctx context.Context, requestID []byte) (*disperser.BlobStatusReply, error) {
	m.Log.Info("Attempting to get blob status from EigenDA")
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credential)}
	conn, err := grpc.Dial(m.RPC, dialOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	daClient := disperser.NewDisperserClient(conn)

	statusRes, err := daClient.GetBlobStatus(ctx, &disperser.BlobStatusRequest{
		RequestId: requestID,
	})
	if err != nil {
		return nil, err
	}

	return statusRes, nil
}

func (m *EigenDA) RetrieveBlob(ctx context.Context, BatchHeaderHash []byte, BlobIndex uint32) ([]byte, error) {
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credential)}
	conn, err := grpc.Dial(m.RPC, dialOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	daClient := disperser.NewDisperserClient(conn)

	ctxTimeout, cancel := context.WithTimeout(ctx, m.RPCTimeout)
	defer cancel()
	reply, err := daClient.RetrieveBlob(ctxTimeout, &disperser.RetrieveBlobRequest{
		BatchHeaderHash: BatchHeaderHash,
		BlobIndex:       BlobIndex,
	})
	if err != nil {
		return nil, err
	}

	// decode modulo bn254
	decodedData := RemoveEmptyByteFromPaddedBytes(reply.Data)

	return decodedData, nil
}

func (m *EigenDA) DisperseBlob(ctx context.Context, txData []byte) (*disperser.BlobInfo, []byte, error) {
	dispersalStart := time.Now()
	var blobInfo *disperser.BlobInfo
	var blobData []byte
	var err error
	done := m.recordInterval("WholeDisperseLifecycle")
	if m.signer == nil {
		blobInfo, blobData, err = m.DisperseBlobWithoutAuth(ctx, txData)
	} else {
		blobInfo, blobData, err = m.DisperseBlobAuthenticated(ctx, txData)
	}
	done(err)
	if err != nil {
		m.Log.Error("error disperse blob", "err", err)
		return nil, nil, err
	}

	if m.verifier == nil {
		m.Log.Info("DisperseBlob without verification", "time", time.Since(dispersalStart))
		return blobInfo, blobData, nil
	}

	encodedTxData := ConvertByPaddingEmptyByte(txData)
	err = m.verifier.VerifyCommitment(blobInfo.BlobHeader.Commitment, encodedTxData)
	if err != nil {
		m.Log.Error("error verify commitment", "err", err)
		return nil, nil, err
	}

	m.Log.Info("Dispersal took", "time", time.Since(dispersalStart))

	dispersalDuration := time.Since(dispersalStart)
	remainingTimeout := m.StatusQueryTimeout - dispersalDuration

	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), remainingTimeout)
	defer cancel()

	blobJson, _ := json.Marshal(blobInfo)
	m.Log.Info("Verifying blob", "info", string(blobJson))

	cert := (*verify.Certificate)(blobInfo)
	finished := false
	for !finished {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-ticker.C:
			err = m.verifier.VerifyCert(cert)
			if err == nil {
				m.Log.Info("Blob verified successfully", "info", string(blobJson))
				finished = true
			} else if !errors.Is(err, verify.ErrBatchMetadataHashNotFound) {
				m.Log.Error("Blob verified with error", "info", string(blobJson))
				return nil, nil, err
			} else {
				m.Log.Info("Blob confirmed, waiting for sufficient confirmation depth...", "targetDepth", m.EthConfirmationDepth)
			}
		}
	}

	return blobInfo, blobData, nil
}

func (m *EigenDA) DisperseBlobWithoutAuth(ctx context.Context, txData []byte) (*disperser.BlobInfo, []byte, error) {
	m.Log.Info("Attempting to disperse blob to EigenDA without auth")
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credential)}
	conn, err := grpc.Dial(m.RPC, dialOptions...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = conn.Close() }()
	daClient := disperser.NewDisperserClient(conn)

	ctxTimeout, cancel := context.WithTimeout(ctx, m.RPCTimeout)
	defer cancel()

	// encode modulo bn254
	encodedTxData := ConvertByPaddingEmptyByte(txData)
	disperseReq := &disperser.DisperseBlobRequest{
		Data: encodedTxData,
	}

	done := m.recordInterval("DisperseBlob")
	disperseRes, err := daClient.DisperseBlob(ctxTimeout, disperseReq)
	done(err)
	if err != nil || disperseRes == nil {
		m.Log.Error("Unable to disperse blob to EigenDA, aborting", "err", err)
		return nil, nil, err
	}

	if disperseRes.Result == disperser.BlobStatus_UNKNOWN ||
		disperseRes.Result == disperser.BlobStatus_FAILED {
		m.Log.Error("Unable to disperse blob to EigenDA, aborting", "err", err)
		return nil, nil, fmt.Errorf("reply status is %d", disperseRes.Result)
	}

	base64RequestID := base64.StdEncoding.EncodeToString(disperseRes.RequestId)

	m.Log.Info("Blob disepersed to EigenDA, now waiting for confirmation", "requestID", base64RequestID)

	var statusRes *disperser.BlobStatusReply
	timeoutTime := time.Now().Add(m.StatusQueryTimeout)
	// Wait before first status check
	time.Sleep(m.StatusQueryRetryInterval)
	for time.Now().Before(timeoutTime) {
		if ctx.Err() != nil {
			m.Log.Warn("context error", "err", ctx.Err())
			return nil, nil, ctx.Err()
		}
		done := m.recordInterval("GetBlobStatus")
		statusRes, err = daClient.GetBlobStatus(ctx, &disperser.BlobStatusRequest{
			RequestId: disperseRes.RequestId,
		})
		done(err)
		if err != nil {
			m.Log.Warn("Unable to retrieve blob dispersal status, will retry", "requestID", base64RequestID, "err", err)
		} else if statusRes.Status == disperser.BlobStatus_CONFIRMED || statusRes.Status == disperser.BlobStatus_FINALIZED {
			// TODO(eigenlayer): As long as fault proofs are disabled, we can move on once a blob is confirmed
			// but not yet finalized, without further logic. Once fault proofs are enabled, we will need to update
			// the proposer to wait until the blob associated with an L2 block has been finalized, i.e. the EigenDA
			// contracts on Ethereum have confirmed the full availability of the blob on EigenDA.
			batchHeaderHashHex := fmt.Sprintf("0x%s", hex.EncodeToString(statusRes.Info.BlobVerificationProof.BatchMetadata.BatchHeaderHash))
			m.Log.Info("Successfully dispersed blob to EigenDA", "requestID", base64RequestID, "batchHeaderHash", batchHeaderHashHex)
			return statusRes.Info, disperseRes.RequestId, nil
		} else if statusRes.Status == disperser.BlobStatus_UNKNOWN ||
			statusRes.Status == disperser.BlobStatus_FAILED {
			m.Log.Error("EigenDA blob dispersal failed in processing", "requestID", base64RequestID, "err", err)
			return nil, nil, fmt.Errorf("eigenDA blob dispersal failed in processing with reply status %d", statusRes.Status)
		} else {
			m.Log.Warn("Still waiting for confirmation from EigenDA", "requestID", base64RequestID)
		}

		// Wait before first status check
		time.Sleep(m.StatusQueryRetryInterval)
	}

	return nil, nil, fmt.Errorf("timed out getting EigenDA status for dispersed blob key: %s", base64RequestID)
}

func (m *EigenDA) DisperseBlobAuthenticated(ctx context.Context, rawData []byte) (*disperser.BlobInfo, []byte, error) {
	m.Log.Info("Attempting to disperse blob to EigenDA with auth")

	// disperse blob
	done := m.recordInterval("DisperseBlobAuthenticated")
	blobStatus, requestID, err := m.disperseBlobAuthenticated(ctx, rawData)
	done(err)
	if err != nil {
		return nil, nil, fmt.Errorf("error calling DisperseBlobAuthenticated() client: %w", err)
	}

	// process response
	if *blobStatus == disperser.BlobStatus_FAILED {
		return nil, nil, fmt.Errorf("reply status is %d", blobStatus)
	}

	base64RequestID := base64.StdEncoding.EncodeToString(requestID)
	m.Log.Info("Blob dispersed to EigenDA, now waiting for confirmation", "requestID", base64RequestID)

	ticker := time.NewTicker(m.Config.StatusQueryRetryInterval)
	defer ticker.Stop()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, m.Config.StatusQueryTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("timed out waiting for EigenDA blob to confirm blob with request id=%s: %w", base64RequestID, ctx.Err())
		case <-ticker.C:
			done := m.recordInterval("GetBlobStatus")
			statusRes, err := m.GetBlobStatus(ctx, requestID)
			done(err)
			if err != nil {
				m.Log.Error("Unable to retrieve blob dispersal status, will retry", "requestID", base64RequestID, "err", err)
				continue
			}

			switch statusRes.Status {
			case disperser.BlobStatus_PROCESSING:
				m.Log.Info("Blob submitted, waiting for dispersal from EigenDA", "requestID", base64RequestID)
			case disperser.BlobStatus_FAILED, disperser.BlobStatus_UNKNOWN:
				m.Log.Error("EigenDA blob dispersal failed in processing", "requestID", base64RequestID, "err", err)
				return nil, nil, fmt.Errorf("EigenDA blob dispersal failed in processing, requestID=%s: %w", base64RequestID, err)
			case disperser.BlobStatus_INSUFFICIENT_SIGNATURES:
				m.Log.Error("EigenDA blob dispersal failed in processing with insufficient signatures", "requestID", base64RequestID, "err", err)
				return nil, nil, fmt.Errorf("EigenDA blob dispersal failed in processing with insufficient signatures, requestID=%s: %w", base64RequestID, err)
			case disperser.BlobStatus_CONFIRMED, disperser.BlobStatus_FINALIZED:
				batchHeaderHashHex := fmt.Sprintf("0x%s", hex.EncodeToString(statusRes.Info.BlobVerificationProof.BatchMetadata.BatchHeaderHash))
				m.Log.Info("Successfully dispersed blob to EigenDA", "requestID", base64RequestID, "batchHeaderHash", batchHeaderHashHex)
				return statusRes.Info, requestID, nil
			default:
				m.Log.Warn("Still waiting for confirmation from EigenDA", "requestID", base64RequestID)
			}
		}
	}
}

func (m *EigenDA) disperseBlobAuthenticated(ctx context.Context, data []byte) (*disperser.BlobStatus, []byte, error) {
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credential)}
	conn, err := grpc.Dial(m.RPC, dialOptions...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = conn.Close() }()
	daClient := disperser.NewDisperserClient(conn)

	ctxTimeout, cancel := context.WithTimeout(ctx, m.RPCTimeout)
	defer cancel()

	stream, err := daClient.DisperseBlobAuthenticated(ctxTimeout)
	if err != nil || stream == nil {
		return nil, nil, fmt.Errorf("error while calling DisperseBlobAuthenticated: %w", err)
	}

	encodedTxData := ConvertByPaddingEmptyByte(data)

	request := &disperser.DisperseBlobRequest{
		Data:      encodedTxData,
		AccountId: m.signer.GetAccountID(),
	}

	// Send the initial request
	err = stream.Send(&disperser.AuthenticatedRequest{Payload: &disperser.AuthenticatedRequest_DisperseRequest{
		DisperseRequest: request,
	}})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Get the Challenge
	reply, err := stream.Recv()
	if err != nil || reply == nil {
		return nil, nil, fmt.Errorf("error while receiving: %w", err)
	}
	authHeaderReply, ok := reply.Payload.(*disperser.AuthenticatedReply_BlobAuthHeader)
	if !ok {
		return nil, nil, errors.New("expected challenge")
	}

	authData, err := m.signer.SignBlobRequest(authHeaderReply.BlobAuthHeader.ChallengeParameter)
	if err != nil {
		return nil, nil, errors.New("error signing blob request")
	}

	// Process challenge and send back challenge_reply
	err = stream.Send(&disperser.AuthenticatedRequest{Payload: &disperser.AuthenticatedRequest_AuthenticationData{
		AuthenticationData: &disperser.AuthenticationData{
			AuthenticationData: authData,
		},
	}})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send challenge reply: %w", err)
	}

	reply, err = stream.Recv()
	if err != nil || reply == nil {
		return nil, nil, fmt.Errorf("error while receiving final reply: %w", err)
	}
	disperseReply, ok := reply.Payload.(*disperser.AuthenticatedReply_DisperseReply) // Process the final disperse_reply
	if !ok {
		return nil, nil, errors.New("expected DisperseReply")
	}

	status := disperseReply.DisperseReply.GetResult()

	return &status, disperseReply.DisperseReply.GetRequestId(), nil
}

func (m *EigenDA) recordInterval(method string) func(error) {
	if m.Metricer == nil {
		return func(err error) {}
	}

	return m.Metricer.RecordInterval(method)
}
