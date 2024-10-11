package eigenda

import (
	"context"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
)

type IEigenDA interface {
	RetrieveBlob(ctx context.Context, requestID []byte) ([]byte, error)
	RetrieveBlobWithCommitment(ctx context.Context, commitment []byte) ([]byte, error)
	DisperseBlob(ctx context.Context, txData []byte) (*disperser.BlobInfo, error)
	GetBlobStatus(ctx context.Context, requestID []byte) (*disperser.BlobStatusReply, error)
}
