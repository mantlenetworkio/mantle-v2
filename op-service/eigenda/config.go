package eigenda

import "time"

type Config struct {
	// EigenDA Proxy HTTP URL
	ProxyUrl string
	// EigenDA Disperser RPC URL
	DisperserUrl string
	// The total amount of time that the batcher will spend waiting for EigenDA to disperse a blob
	DisperseBlobTimeout time.Duration
	// The total amount of time that the batcher will spend waiting for EigenDA to retrieve a blob
	RetrieveBlobTimeout time.Duration
}
