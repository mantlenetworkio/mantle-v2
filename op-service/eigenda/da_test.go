package eigenda

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestRequestID(t *testing.T) {
	a := "0x2a713e7295abd3dd05d02826c49f10b0eefc4a7d51560ef8ce473ba6e4d59d83" //  hex return value from the log
	b, _ := hexutil.Decode(a)                                                 // import "github.com/ethereum/go-ethereum/common/hexutil"
	data := base64.StdEncoding.EncodeToString(b)                              // import "encoding/base64"
	fmt.Println(data)
}

func TestBase64(t *testing.T) {
	data, _ := base64.StdEncoding.DecodeString("E0QCZqLeithsMYw3IoyrgaSZL0MBGeyoLIUpuRnWnPE=")
	fmt.Println(big.NewInt(0).SetBytes(data))
	fmt.Println(data)
	fmt.Println(hexutil.Encode(data))
	a := "0x2a713e7295abd3dd05d02826c49f10b0eefc4a7d51560ef8ce473ba6e4d59d83"
	b, _ := hexutil.Decode(a)
	s := base64.StdEncoding.EncodeToString(b)
	fmt.Println(s)
}
