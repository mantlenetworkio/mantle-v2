package eigenda

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ErrNotFound is returned when the server could not find the input.
var ErrNotFound = errors.New("not found")

// ErrInvalidInput is returned when the input is not valid for posting to the DA storage.
var ErrInvalidInput = errors.New("invalid input")

// ErrNetwork is returned when there is a eigenda network error.
var ErrNetwork = errors.New("eigenda network error")

// NewEigenDAClient is an HTTP client to communicate with EigenDA Proxy.
// It creates commitments and retrieves input data + verifies if needed.
type EigenDAClient struct {
	proxyUrl            string
	disperserUrl        string
	log                 log.Logger
	metricer            Metrics
	disperseClient      *http.Client
	retrieveClient      *http.Client
	retrieveBlobTimeout time.Duration
}

const (
	CertV0                byte = 0
	EigenDACommitmentType byte = 0
	GenericCommitmentType byte = 1
)

// EigenDAClient returns a new EigenDA Proxy client.
func NewEigenDAClient(cfg Config, log log.Logger, m Metrics) *EigenDAClient {
	return &EigenDAClient{
		proxyUrl:            cfg.ProxyUrl,
		disperserUrl:        cfg.DisperserUrl,
		disperseClient:      &http.Client{Timeout: cfg.DisperseBlobTimeout},
		retrieveClient:      &http.Client{Timeout: cfg.RetrieveBlobTimeout},
		retrieveBlobTimeout: cfg.RetrieveBlobTimeout,
		log:                 log,
		metricer:            m,
	}
}

// RetrieveBlob qqreturns the input data for the given batch header and blob index.
// This method is used for the op-node compatibility.
// Only RetrieveBlobWithCommitment supports EigenDA S3 fallback
func (c *EigenDAClient) RetrieveBlob(ctx context.Context, BatchHeaderHash []byte, BlobIndex uint32) ([]byte, error) {
	c.log.Info("Attempting to retrieve blob from EigenDA", "BatchHeaderHash", hex.EncodeToString(BatchHeaderHash), "blobIndex", BlobIndex)
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credential), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100 * 1024 * 1024))} // 100MiB receive buffer
	conn, err := grpc.Dial(c.disperserUrl, dialOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	daClient := disperser.NewDisperserClient(conn)

	ctxTimeout, cancel := context.WithTimeout(ctx, c.retrieveBlobTimeout)
	defer cancel()
	done := c.recordInterval("RetrieveBlob")
	reply, err := daClient.RetrieveBlob(ctxTimeout, &disperser.RetrieveBlobRequest{
		BatchHeaderHash: BatchHeaderHash,
		BlobIndex:       BlobIndex,
	})
	done(err)
	if err != nil {
		return nil, err
	}

	// decode modulo bn254
	decodedData := RemoveEmptyByteFromPaddedBytes(reply.Data)

	return decodedData, nil
}

// RetrieveBlob returns the input data for the given encoded commitment bytes.
func (c *EigenDAClient) RetrieveBlobWithCommitment(ctx context.Context, commitment []byte) ([]byte, error) {
	c.log.Info("Attempting to retrieve blob from EigenDA with commitment", "commitment", hex.EncodeToString(commitment))
	blobInfo, err := DecodeCommitment(commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to decode commitment: %w", err)
	}
	c.log.Info("Blob info", "BatchHeaderHash", hex.EncodeToString(blobInfo.BlobVerificationProof.BatchMetadata.BatchHeaderHash), "blobIndex", blobInfo.BlobVerificationProof.BlobIndex)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/get/0x%x", c.proxyUrl, commitment), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	done := c.recordInterval("RetrieveBlobWithCommitment")
	resp, err := c.retrieveClient.Do(req)
	err = func() error {
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get preimage: %v", resp.StatusCode)
		}
		return nil
	}()
	done(err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	input, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return input, nil
}

// DisperseBlob sets the input data and returns the respective commitment.
func (c *EigenDAClient) DisperseBlob(ctx context.Context, img []byte) (*disperser.BlobInfo, error) {
	c.log.Info("Attempting to disperse blob to EigenDA")
	if len(img) == 0 {
		return nil, ErrInvalidInput
	}

	body := bytes.NewReader(img)
	url := fmt.Sprintf("%s/put/", c.proxyUrl)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	done := c.recordInterval("DisperseBlob")
	resp, err := c.disperseClient.Do(req)
	done(err)
	if err != nil {
		return nil, fmt.Errorf("failed to post store data orirgin error: %w error: %w", err, ErrNetwork)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to store data status code: %v error: %w", resp.StatusCode, ErrNetwork)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	comm, err := DecodeCommitment(b)
	if err != nil {
		return nil, err
	}

	blobProof := comm.BlobVerificationProof
	c.log.Info("Dispersed blob to EigenDA successfully", "BatchHeaderHash", hex.EncodeToString(blobProof.BatchMetadata.BatchHeaderHash), "BlobIndex", blobProof.BlobIndex)

	return comm, nil
}

func (c *EigenDAClient) GetBlobStatus(ctx context.Context, requestID []byte) (*disperser.BlobStatusReply, error) {
	c.log.Info("Attempting to get blob status from EigenDA", "RequestID", hex.EncodeToString(requestID))
	config := &tls.Config{}
	credential := credentials.NewTLS(config)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(credential)}
	conn, err := grpc.Dial(c.disperserUrl, dialOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	daClient := disperser.NewDisperserClient(conn)

	ctxTimeout, cancel := context.WithTimeout(ctx, c.retrieveBlobTimeout)
	defer cancel()
	done := c.recordInterval("GetBlobStatus")
	statusRes, err := daClient.GetBlobStatus(ctxTimeout, &disperser.BlobStatusRequest{
		RequestId: requestID,
	})
	done(err)
	if err != nil {
		return nil, err
	}

	return statusRes, nil
}

// GetBlobExtraInfo returns the extra data for the given encoded commitment bytes.
// Currently, it only returns request_id.
func (c *EigenDAClient) GetBlobExtraInfo(ctx context.Context, commitment []byte) (map[string]interface{}, error) {
	c.log.Info("Attempting to retrieve blob extra info with commitment", "commitment", hex.EncodeToString(commitment))
	blobInfo, err := DecodeCommitment(commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to decode commitment: %w", err)
	}
	c.log.Info("Blob info", "BatchHeaderHash", hex.EncodeToString(blobInfo.BlobVerificationProof.BatchMetadata.BatchHeaderHash), "blobIndex", blobInfo.BlobVerificationProof.BlobIndex)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/get_extra/0x%x", c.proxyUrl, commitment), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	done := c.recordInterval("GetBlobExtraInfo")
	resp, err := c.retrieveClient.Do(req)
	err = func() error {
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get extra info: %v", resp.StatusCode)
		}
		return nil
	}()
	done(err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	output := make(map[string]interface{})
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (m *EigenDAClient) recordInterval(method string) func(error) {
	if m.metricer == nil {
		return func(err error) {}
	}

	return m.metricer.RecordInterval(method)
}

func DecodeCommitment(commitment []byte) (*disperser.BlobInfo, error) {
	if len(commitment) < 3 {
		return nil, fmt.Errorf("commitment is too short")
	}

	opType, daProvider, certVersion := commitment[0], commitment[1], commitment[2]
	if opType != GenericCommitmentType || daProvider != EigenDACommitmentType || certVersion != CertV0 {
		return nil, fmt.Errorf("invalid commitment type")
	}

	data := commitment[3:]
	blobInfo := &disperser.BlobInfo{}
	err := rlp.DecodeBytes(data, blobInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to decode commitment")
	}

	return blobInfo, nil
}

func EncodeCommitment(val *disperser.BlobInfo) ([]byte, error) {
	bytes, err := rlp.EncodeToBytes(val)
	if err != nil {
		return nil, fmt.Errorf("failed to encode DA cert to RLP format: %w", err)
	}

	return append([]byte{GenericCommitmentType, EigenDACommitmentType, CertV0}, bytes...), nil
}
