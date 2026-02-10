package dsl

import (
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum/go-ethereum/common"
)

// MNTFunder funds accounts with L1 MNT using an existing holder account.
type MNTFunder struct {
	commonImpl
	holder    *EOA
	tokenAddr common.Address
}

func NewMNTFunder(t devtest.T, tokenAddr common.Address, holder *EOA) *MNTFunder {
	t.Require().NotNil(holder, "MNT holder must not be nil")
	return &MNTFunder{
		commonImpl: commonFromT(t),
		holder:     holder,
		tokenAddr:  tokenAddr,
	}
}

func (f *MNTFunder) FundAtLeast(to *EOA, amount eth.ETH) eth.ETH {
	currentBalance := to.GetTokenBalance(f.tokenAddr)
	if currentBalance.Gt(amount) || currentBalance.ToBig().Cmp(amount.ToBig()) == 0 {
		return currentBalance
	}
	missing := amount.Sub(currentBalance)
	finalBalance := currentBalance.Add(missing)
	f.Fund(to, missing)
	to.WaitForTokenBalance(f.tokenAddr, finalBalance)
	return finalBalance
}

func (f *MNTFunder) Fund(to *EOA, amount eth.ETH) {
	f.require.NotNil(to, "recipient must not be nil")
	f.require.Equal(f.holder.el.ChainID(), to.el.ChainID(), "holder and recipient must be on same chain")

	token := bindings.NewBindings[bindings.OptimismMintableERC20](
		bindings.WithTest(f.t),
		bindings.WithClient(f.holder.el.stackEL().EthClient()),
		bindings.WithTo(f.tokenAddr),
	)

	holderBalance, err := contractio.Read(token.BalanceOf(f.holder.Address()), f.t.Ctx())
	f.require.NoError(err, "failed to read MNT holder balance")
	f.require.True(holderBalance.Gt(amount) || holderBalance.ToBig().Cmp(amount.ToBig()) == 0, "MNT holder has insufficient balance")
	_, err = contractio.Write(token.Transfer(to.Address(), amount), f.t.Ctx(), f.holder.Plan())
	f.require.NoError(err, "failed to transfer MNT")
}
