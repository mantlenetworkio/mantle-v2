package arsia

import (
	"crypto/rand"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestCalldataCost(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)

	dat := make([]byte, 2048)
	_, err := rand.Read(dat)
	require.NoError(err)

	idPrecompile := common.BytesToAddress([]byte{0x4})
	idTxOpts := txplan.Combine(
		alice.Plan(),
		txplan.WithData(dat),
		txplan.WithTo(&idPrecompile),
	)

	idTx := txplan.NewPlannedTx(idTxOpts)
	receipt, err := idTx.Included.Eval(t.Ctx())
	require.NoError(err)
	require.Equal(uint64(1), receipt.Status)

	// ID Precompile:
	//   data_word_size = (data_size + 31) / 32
	//   id_static_gas = 15
	//   id_dynamic_gas = 3 * data_word_size
	// EIP-7623:
	//   total_cost_floor_per_token = 10
	//   standard_token_cost = 4
	//   tokens_in_calldata = zero_bytes_in_calldata + nonzero_bytes_in_calldata * 4
	//   calldata_cost = standard_token_cost * tokens_in_calldata
	//
	// Expected gas usage is:
	// 21_000 (base cost) + max(id_static_gas + id_dynamic_gas + calldata_cost, total_cost_floor_per_token * tokens_in_calldata)
	var zeros, nonZeros int
	for _, b := range dat {
		if b == 0 {
			zeros++
		} else {
			nonZeros++
		}
	}
	tokensInCalldata := zeros + nonZeros*4

	expectedGas := 21_000 + max(15+3*((len(dat)+31)/32)+4*tokensInCalldata, 10*tokensInCalldata)
	require.EqualValues(expectedGas, receipt.GasUsed, "Gas usage does not match expected value")

	t.Log("Calldata cost test completed successfully",
		"gasUsed", receipt.GasUsed,
		"expectedGas", expectedGas,
		"zeros", zeros,
		"nonZeros", nonZeros)
}
