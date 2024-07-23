package eigenda

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	bsscore "github.com/ethereum-optimism/optimism/bss-core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/api/option"
)

type BlobRequestSigner interface {
	SignBlobRequest(nonce uint32) ([]byte, error)
	GetAccountID() string
}

type localBlobSigner struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewLocalBlobSigner(privateKeyHex string) (*localBlobSigner, error) {

	privateKeyBytes := common.FromHex(privateKeyHex)
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	return &localBlobSigner{
		PrivateKey: privateKey,
	}, nil
}

func (s *localBlobSigner) SignBlobRequest(nonce uint32) ([]byte, error) {

	// Message you want to sign
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, nonce)
	hash := crypto.Keccak256(buf)

	// Sign the hash using the private key
	sig, err := crypto.Sign(hash, s.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %v", err)
	}

	return sig, nil
}

func (s *localBlobSigner) GetAccountID() string {

	publicKeyBytes := crypto.FromECDSAPub(&s.PrivateKey.PublicKey)
	return hexutil.Encode(publicKeyBytes)

}

type hsmBlobSigner struct {
	mk     *bsscore.ManagedKey
	pubkey *ecdsa.PublicKey
}

func NewHsmBlobSigner(hsmCreden, hsmAPIName, hsmPubkey string) (*hsmBlobSigner, error) {

	proBytes, err := hex.DecodeString(hsmCreden)
	if err != nil {
		return nil, err
	}
	apikey := option.WithCredentialsJSON(proBytes)
	ctx, cancle := context.WithTimeout(context.Background(), time.Minute)
	defer cancle()
	client, err := kms.NewKeyManagementClient(ctx, apikey)
	if err != nil {
		return nil, err
	}

	rawData, err := base64.StdEncoding.DecodeString(hsmPubkey)
	if err != nil {
		return nil, err
	}

	//rawData[23:] skip the header
	pubkey, err := crypto.UnmarshalPubkey(rawData[23:])
	if err != nil {
		return nil, err
	}

	mk := &bsscore.ManagedKey{
		KeyName:      hsmAPIName,
		EthereumAddr: crypto.PubkeyToAddress(*pubkey),
		Gclient:      client,
	}
	return &hsmBlobSigner{mk: mk, pubkey: pubkey}, nil
}

func (s *hsmBlobSigner) SignBlobRequest(nonce uint32) ([]byte, error) {

	//Message you want to sign
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, nonce)
	hash := crypto.Keccak256(buf)

	// hash the transaction (with Keccak-256 probably)
	ctx, cancle := context.WithTimeout(context.Background(), time.Second*10)
	defer cancle()
	sig, err := s.mk.SignHash(ctx, common.BytesToHash(hash))
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (s *hsmBlobSigner) GetAccountID() string {
	publicKeyBytes := crypto.FromECDSAPub(s.pubkey)
	return hexutil.Encode(publicKeyBytes)
}
