package eigenda

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Layr-Labs/eigenda/api/grpc/common"
	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

type mockMetrics struct{}

func (m *mockMetrics) RecordInterval(method string) func(error) {
	return func(error) {}
}

func TestNewEigenDAProxy_RetrieveBlob(t *testing.T) {
	requestId := make([]byte, 189)
	_, _ = base64.StdEncoding.Decode(requestId, []byte("YTU3NWE0ZWY5MGY5NTU5ZWVlYTg2Nzg5NzJkNTAxMzllODBjNDExNjBlMGQ1NmU2ZTc3MWJkNjdmZmMxMzQxMy0zMTM3MzIzNzM2MzAzMzMwMzUzNTM0MzgzMTM5MzEzMDM2MzIzODJmMzEyZjMzMzMyZjMwMmYzMzMzMmZlM2IwYzQ0Mjk4ZmMxYzE0OWFmYmY0Yzg5OTZmYjkyNDI3YWU0MWU0NjQ5YjkzNGNhNDk1OTkxYjc4NTJiODU1"))
	type fields struct {
		proxyUrl     string
		disperserUrl string
		log          log.Logger
	}
	type args struct {
		ctx       context.Context
		requestId []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				proxyUrl:     "http://127.0.0.1:3100",
				disperserUrl: "disperser-holesky.eigenda.xyz:443",
				log:          log.New("test"),
			},
			args: args{
				ctx:       context.Background(),
				requestId: requestId,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &EigenDAClient{
				proxyUrl:     tt.fields.proxyUrl,
				disperserUrl: tt.fields.disperserUrl,
				log:          tt.fields.log,
			}
			got, err := c.RetrieveBlob(tt.args.ctx, nil, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("EigenDAClient.RetrieveBlob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Printf("%v 0x%x\n", err, got)
		})
	}
}

func TestNewEigenDAProxy_RetrieveBlobWithCommitment(t *testing.T) {
	type fields struct {
		proxyUrl     string
		disperserUrl string
		log          log.Logger
	}
	type args struct {
		ctx        context.Context
		commitment []byte
	}
	commitment, err := hex.DecodeString("010000f901d5f850f842a00d6900d97bec407c2ca8b3df3b1edf43c1a0844f5860a512bf172a95e989e328a018baf1b88e34d62e61b7d01e753044b5a07096cf06495de22b5dc09b57a184d904cac480213701c401213701f901808298428193f873eba096f24191fcb8cd1bf9efe6d9750fcdbe6ed4e447bedbcd67f02cb66cfdad8d1f82000182636483251ab5a0e3a7e17a3c2de5ae079a08782dd641f6925dcb9b7d867636c44fe718d01cdf150083251b11a040c4f8aa8c31d0383abcbd8c28a6a03d34427402394354fcb321b5538e203189b9010032fc6bbd3c99d54afdb411593e608dde464b472174695673b6ff9778bacc8509c323c0625a066534d65f81c229d2ab9757638571225a7a6b85574cbc400d1441a8cf662b344a859735e196880dcc460852093d8c9d3258fbac558444e971b9bdbb055c6f5e294cb9dd1622dc5a2bcaffb477fbb3ece07ef57dd2cdcb9e988cc54d1c93bbe897f557f9085716d5a2b0cc44e4fe7a32382a73b2102897793bbda1630c2670465485c9f7fcae61071429c7aeb718fc667d761078d8ad677e8b478c83fd646e6d5243d75d27ff177136f8ba7146d4ae8f95cbff78944936ace28659145378eda479b5383df1952899aa42327c32667e02fb2c13ac644494d6da071a820001")
	fmt.Println(err)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				proxyUrl:     "http://127.0.0.1:3100",
				disperserUrl: "disperser-holesky.eigenda.xyz:443",
				log:          log.New("test"),
			},
			args: args{
				ctx:        context.Background(),
				commitment: commitment,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &EigenDAClient{
				proxyUrl:     tt.fields.proxyUrl,
				disperserUrl: tt.fields.disperserUrl,
				log:          tt.fields.log,
			}
			got, err := c.RetrieveBlobWithCommitment(tt.args.ctx, tt.args.commitment)
			if (err != nil) != tt.wantErr {
				t.Errorf("EigenDAClient.RetrieveBlobWithCommitment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Printf("%v 0x%x\n", err, got)
		})
	}
}

func TestNewEigenDAProxy_DisperseBlob(t *testing.T) {
	type fields struct {
		client *EigenDAClient
	}
	type args struct {
		ctx context.Context
		img []byte
	}

	logger := log.New()
	metrics := &mockMetrics{}

	client := NewEigenDAClient(Config{
		ProxyUrl:            "http://localhost:3100",
		DisperserUrl:        "disperser-holesky.eigenda.xyz:443",
		DisperseBlobTimeout: 10 * time.Minute,
		RetrieveBlobTimeout: 10 * time.Second,
	}, logger, metrics)

	invalidClient := NewEigenDAClient(Config{
		ProxyUrl:            "http://localhost:3333",
		DisperserUrl:        "disperser-holesky.eigenda.xyz:443",
		DisperseBlobTimeout: 10 * time.Minute,
		RetrieveBlobTimeout: 10 * time.Second,
	}, logger, metrics)

	tests := []struct {
		name           string
		fields         fields
		args           args
		want           *disperser.BlobInfo
		wantErr        bool
		wantNetworkErr bool
	}{
		{
			name: "t1",
			fields: fields{
				client: client,
			},
			args: args{
				ctx: context.Background(),
				img: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
		},
		{
			name: "t2",
			fields: fields{
				client: invalidClient,
			},
			args: args{
				ctx: context.Background(),
				img: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			wantErr:        true,
			wantNetworkErr: true,
		},
		{
			name: "t3",
			fields: fields{
				client: client,
			},
			args: args{
				ctx: context.Background(),
				img: make([]byte, 30*1024*1024), //larger than eigenda throughput limit
			},
			wantErr:        true,
			wantNetworkErr: true,
		},
		{
			name: "t4",
			fields: fields{
				client: client,
			},
			args: args{
				ctx: context.Background(),
				img: []byte{}, //empty data error
			},
			wantErr:        true,
			wantNetworkErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.client.DisperseBlob(tt.args.ctx, tt.args.img)
			if (err != nil) != tt.wantErr {
				t.Errorf("EigenDAClient.DisperseBlob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Error(t, err)
				t.Logf("error: %v", err)
			}
			if tt.wantNetworkErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrNetwork)
				t.Logf("network error: %v", err)
			}
			if !tt.wantErr {
				commitment, err := EncodeCommitment(got)
				require.NoError(t, err)
				fmt.Printf("%v 0x%x\n", err, commitment)
			}
		})
	}
}

func TestNewEigenDAProxy_GetBlobStatus(t *testing.T) {
	type fields struct {
		proxyUrl     string
		disperserUrl string
		log          log.Logger
	}
	type args struct {
		ctx       context.Context
		requestID []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *disperser.BlobStatusReply
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &EigenDAClient{
				proxyUrl:     tt.fields.proxyUrl,
				disperserUrl: tt.fields.disperserUrl,
				log:          tt.fields.log,
			}
			got, err := m.GetBlobStatus(tt.args.ctx, tt.args.requestID)
			if (err != nil) != tt.wantErr {
				t.Errorf("EigenDAClient.GetBlobStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EigenDAClient.GetBlobStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBlobExtraInfo(t *testing.T) {
	cfg := Config{
		ProxyUrl:            "http://localhost:3100",
		DisperserUrl:        "disperser-holesky.eigenda.xyz:443",
		DisperseBlobTimeout: 10 * time.Minute,
		RetrieveBlobTimeout: 10 * time.Second,
	}
	logger := log.New()
	metrics := &mockMetrics{}
	client := NewEigenDAClient(cfg, logger, metrics)

	t.Run("GetBlobExtraInfo with valid commitment", func(t *testing.T) {
		ctx := context.Background()
		// test1: DisperseBlob then GetBlobExtraInfo
		blob := []byte("test data")
		blobInfo, err := client.DisperseBlob(ctx, blob)
		require.NoError(t, err)

		commitment, err := EncodeCommitment(blobInfo)
		require.NoError(t, err)

		extraInfo, err := client.GetBlobExtraInfo(ctx, commitment)
		require.NoError(t, err)
		require.NotEmpty(t, extraInfo)
	})

	t.Run("GetBlobExtraInfo with zero commitment", func(t *testing.T) {
		ctx := context.Background()
		// test1: Encode zero value commitment
		zeroBlobInfo := &disperser.BlobInfo{
			BlobHeader: &disperser.BlobHeader{
				Commitment: &common.G1Commitment{},
			},
			BlobVerificationProof: &disperser.BlobVerificationProof{
				BatchMetadata: &disperser.BatchMetadata{
					BatchHeader:     &disperser.BatchHeader{},
					BatchHeaderHash: make([]byte, 32),
				},
				BlobIndex: 0,
			},
		}
		commitment, err := EncodeCommitment(zeroBlobInfo)
		require.NoError(t, err)
		t.Logf("commitment: %x", commitment)
		extraInfo, err := client.GetBlobExtraInfo(ctx, commitment)
		require.NoError(t, err)
		t.Logf("extraInfo: %v", extraInfo)
		require.Empty(t, extraInfo)
	})
}
