package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-node/p2p"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/driver"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	ssources "github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestUnixTimeStale(t *testing.T) {
	require.True(t, unixTimeStale(1_600_000_000, 1*time.Hour))
	require.False(t, unixTimeStale(uint64(time.Now().Unix()), 1*time.Hour))
}

func TestOpNode_initL1BeaconAPI(t *testing.T) {
	type fields struct {
		log            log.Logger
		appVersion     string
		metrics        *metrics.Metrics
		l1HeadsSub     ethereum.Subscription
		l1SafeSub      ethereum.Subscription
		l1FinalizedSub ethereum.Subscription
		l1Source       *sources.L1Client
		l2Driver       *driver.Driver
		l2Source       *sources.EngineClient
		rpcSync        *sources.SyncClient
		server         *rpcServer
		p2pNode        *p2p.NodeP2P
		p2pSigner      p2p.Signer
		tracer         Tracer
		runCfg         *RuntimeConfig
		resourcesCtx   context.Context
		resourcesClose context.CancelFunc
		beacon         *ssources.L1BeaconClient
	}
	type args struct {
		ctx context.Context
		cfg *Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				ctx: context.Background(),
				cfg: &Config{
					Beacon: &L1BeaconEndpointConfig{
						//BeaconAddr:             "https://ethereum-holesky-beacon-api.publicnode.com",
						//BeaconAddr:             "https://eth-sepolia.g.alchemy.com/v2/XMS1J6f654XZolfd7oaMe-kaNPEpWifX",
						//BeaconAddr: "http://ethereum-mainnet.core.chainstack.com/beacon/87483ac9f236e03d3acdd862a0e97dc7/",
						BeaconAddr:             "http://localhost:5052",
						BeaconHeader:           "",
						BeaconArchiverAddr:     "",
						BeaconCheckIgnore:      false,
						BeaconFetchAllSidecars: false,
					},
				},
			},
			fields: fields{
				log: log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &OpNode{
				log:            tt.fields.log,
				appVersion:     tt.fields.appVersion,
				metrics:        tt.fields.metrics,
				l1HeadsSub:     tt.fields.l1HeadsSub,
				l1SafeSub:      tt.fields.l1SafeSub,
				l1FinalizedSub: tt.fields.l1FinalizedSub,
				l1Source:       tt.fields.l1Source,
				l2Driver:       tt.fields.l2Driver,
				l2Source:       tt.fields.l2Source,
				rpcSync:        tt.fields.rpcSync,
				server:         tt.fields.server,
				p2pNode:        tt.fields.p2pNode,
				p2pSigner:      tt.fields.p2pSigner,
				tracer:         tt.fields.tracer,
				runCfg:         tt.fields.runCfg,
				resourcesCtx:   tt.fields.resourcesCtx,
				resourcesClose: tt.fields.resourcesClose,
				beacon:         tt.fields.beacon,
			}

			// if err := n.initL1(tt.args.ctx, tt.args.cfg); err != nil {
			// 	t.Errorf("OpNode.initL1BeaconAPI() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }

			if err := n.initL1BeaconAPI(tt.args.ctx, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("OpNode.initL1BeaconAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			blobs, err := n.beacon.GetBlobs(tt.args.ctx, eth.L1BlockRef{
				// Hash:       common.HexToHash("0x18d668e37c471b1258431c79fd8f1d0be2ad85b78894ec97ed6d1d8d927a3680"),
				// Number:     1541070,
				// ParentHash: common.HexToHash("0x6c5711e7adf4a3d25dbe4ec089b8391171d1fdfcb8940c7ebbdbe18672e21a0d"),
				// Time:       0x66430a1c,
				Hash:       common.HexToHash("0x12be7af179e0937288f2bd6fe498b96dd72723e2e96fc1379df40e078af55e7f"),
				Number:     6009409,
				ParentHash: common.HexToHash("0x22209247f0237ac590d72eb73cf4167795a7d7c17eeff1f8e9ca495b627fad2a"),
				Time:       0x665912ac,
			}, []eth.IndexedBlobHash{
				{
					Index: 2,
					Hash:  common.HexToHash("0x01fa31ad32ee8f1ee109a4fe4786e3f0342abeeeae3b19ad4506e7131b38606b"),
				},
			})

			if err != nil {
				t.Errorf("beacon.GetBlobs() error = %v", err)
				return
			}

			out := []eth.Data{}
			data := []byte{}
			for _, blob := range blobs {
				blobData, err := blob.ToData()
				if err != nil {
					t.Errorf("ignoring blob due to parse failure %v", err)
					continue
				}
				data = append(data, blobData[1:]...)
			}
			fmt.Println(err)
			fmt.Println(derive.ParseFrames(out[0]))
		})
	}
}
