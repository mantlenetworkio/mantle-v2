package da

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"google.golang.org/protobuf/proto"

	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
)

func TestRetrieveBlob(t *testing.T) {
	// da := eigenda.NewEigenDAClient(
	cfg := Config{
		Config: eigenda.Config{
			DisperserUrl:        "disperser-holesky.eigenda.xyz:443",
			ProxyUrl:            "http://127.0.0.1:3100",
			DisperseBlobTimeout: 20 * time.Minute,
			RetrieveBlobTimeout: 20 * time.Minute,
		},
	}
	// 	log.New(context.Background()),
	// 	nil,
	// )

	calldata, _ := hex.DecodeString("ed1293040a20313594efb8d9721c7f000ba5dcc3e51a717e33662d318bb88c24896fda6473ff102218978a9a012202000128f2a442aa06de03010000f901d8f854f842a01805f5ef4df0dfc3d4ef78648e59b9888345e90ea6e8cdd5d662856782260517a000e4879b77e0e9542db302a559ef63944dcef02f5d1660bfb8f1793ba4b2c3938288dcccc480213740c6012137820400f9017f82b28d22f873eba0e6c0d26a4c0f30c76df3730f82e2f26d9015c43b90a61f260d0b0a8cb261c39d82000182416483268517a0c6cce1fadc0eaea692ce44586249c41b42befd001eb44363e7a7e130cafbe4bd008326856da0313594efb8d9721c7f000ba5dcc3e51a717e33662d318bb88c24896fda6473ffb901009bc91a9ff8246b9a3fd2efd49a48fa41add3eb78a208988f1497baae5837909faa4a8f58ed9cb9bd5bac32b7d2dd9483764d2af98b7752846f5677d4012864bcbbaa78187c8d4a7be2354515e8e25b4d7d9ee75b12581ca995437d6ca9e959d7767ef5a372b4587a4288aa16242217e1ab6252ac6193a88add75b6f7d9ca6f2f5125d73c4ebcadf3d82ab1f54d8fe2e91d4c193905b91fbbd9dd0fd30bd3fd2f5f2d5d7ff44d4baaabcb0f5e60d4217eceba6e3e4c2d160d9c7e78718d2018f26abb188a83feada8847ba19ad49eb97979a6e147eabf6a729b56903eda3722b574890eca80948cc8eff9eb9bd4037cc38b00f20136e66b756692734f403c337c820001")
	calldataFrame := &op_service.CalldataFrame{}
	err := proto.Unmarshal(calldata[1:], calldataFrame)
	if err != nil {
		t.Errorf("Unmarshal err:%v", err)
		return
	}
	frame := calldataFrame.Value.(*op_service.CalldataFrame_FrameRef)
	da, err := NewEigenDADataStore(context.Background(), log.New("t1"), &cfg, nil)
	if err != nil {
		t.Errorf("NewEigenDADataStore err:%v", err)
		return
	}
	fmt.Printf("%x\n%x\n", frame.FrameRef.BatchHeaderHash, frame.FrameRef.Commitment)
	data, err := da.RetrieveBlob(frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex, nil)
	if err != nil {
		t.Errorf("RetrieveBlob err:%v", err)
		return
	}
	fmt.Printf("RetrieveBlob data:%x\n", data)

	data, err = da.RetrieveBlob(frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex, frame.FrameRef.Commitment)
	if err != nil {
		t.Errorf("RetrieveBlob err:%v", err)
		return
	}

	fmt.Printf("RetrieveBlob data:%x\n", data)

	data = data[:frame.FrameRef.BlobLength]
	outData := []eth.Data{}
	err = rlp.DecodeBytes(data, &outData)
	if err != nil {
		log.Error("Decode retrieval frames in error,skip wrong data", "err", err, "blobInfo", fmt.Sprintf("%x:%d", frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex))
		return
	}

	fmt.Printf("RetrieveBlob %d\n", len(outData))
}

func TestEigenDADataStore_RetrieveFromDaIndexer(t *testing.T) {
	tests := []struct {
		name     string
		daConfig Config
		query    string
		wantErr  bool
	}{
		{
			name: "successful",
			daConfig: Config{
				MantleDaIndexerSocket: "da-index-grpc-sepolia-qa7.s7.gomantle.org:443",
				MantleDAIndexerEnable: true,
			},
			query:   "0xc2336ace05b2b72325e860c8856cdf477a03970e0cdd44c5f5e1abdf02359167",
			wantErr: false,
		},
		{
			name: "invalid endpoint",
			daConfig: Config{
				MantleDaIndexerSocket: "da-index-grpc-sepolia-qa7.s7.gomantle.org:80",
				MantleDAIndexerEnable: true,
			},
			query:   "0xc2336ace05b2b72325e860c8856cdf477a03970e0cdd44c5f5e1abdf02359167",
			wantErr: true,
		},
		{
			name: "invalid query",
			daConfig: Config{
				MantleDaIndexerSocket: "da-index-grpc-sepolia-qa7.s7.gomantle.org:443",
				MantleDAIndexerEnable: true,
			},
			query:   "0x00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eigenDaSyncer, err := NewEigenDADataStore(context.Background(), log.New("t1"), &tt.daConfig, nil)
			if err != nil {
				t.Errorf("NewEigenDADataStore err:%v", err)
				return
			}

			if !eigenDaSyncer.IsDaIndexer() {
				t.Fatal("DA indexer should be enabled")
			}

			data, err := eigenDaSyncer.RetrievalFramesFromDaIndexer(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievalFramesFromDaIndexer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				t.Logf("RetrievalFramesFromDaIndexer() error = %v", err)
				return
			}

			outData := []eth.Data{}
			err = rlp.DecodeBytes(data, &outData)
			if err != nil {
				t.Fatalf("Failed to decode retrieval frames: %v", err)
			}

			t.Logf("RetrievalFramesFromDaIndexer() = %v", len(outData))
		})
	}
}
