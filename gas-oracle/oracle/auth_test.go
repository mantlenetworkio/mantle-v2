package oracle

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestAuth(t *testing.T) {
	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		l2ChainID:  big.NewInt(1),
		privateKey: privateKey,
		EnableHsm:  false,
		HsmAddress: "",
	}
	client, err := ethclient.Dial("https://rpc.mantle.xyz")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := NewAuth(cfg, client)
	if err != nil {
		t.Fatal(err)
	}
	opts := auth.Opts()
	if opts.From != crypto.PubkeyToAddress(privateKey.PublicKey) {
		t.Errorf("incorrect From address: got %v, want %v", opts.From, crypto.PubkeyToAddress(privateKey.PublicKey))
	}
	if opts.GasPrice.Cmp(big.NewInt(0)) == 0 {
		t.Error("gas price is not set")
	}
}
