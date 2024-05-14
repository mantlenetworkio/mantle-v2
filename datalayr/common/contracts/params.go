package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type FeeParams struct {
	FeePerBytePerTime   *big.Int
	DlnStake            *big.Int
	CommitSigCollateral *big.Int
}

func (f *FeeParams) GetPrecommitFee(duration uint8, numByte uint32, scale uint64) *big.Int {
	durationBig := new(big.Int).SetUint64(uint64(duration))
	scaleBig := new(big.Int).SetUint64(scale)
	scaledDurationBig := new(big.Int).Mul(durationBig, scaleBig)
	feePerByte := new(big.Int).Mul(scaledDurationBig, f.FeePerBytePerTime)
	numByteBig := big.NewInt(int64(numByte))
	fee := new(big.Int).Mul(feePerByte, numByteBig)
	return fee
}

func (dlcc *DataLayrChainClient) GetQuorumThresholdBasisPoints() (*big.Int, error) {
	basisPoints, err := dlcc.Bindings.DlSm.QuorumThresholdBasisPoints(&bind.CallOpts{})
	if err != nil {
		dlcc.Logger.Info().Msgf("QUO err %v", err)
		return nil, ErrCannotReadFromChain
	}
	basisPointsBitInt := new(big.Int).SetUint64(uint64(basisPoints))
	return basisPointsBitInt, nil
}

func (dlcc *DataLayrChainClient) GetAdversaryThresholdBasisPoints() (*big.Int, error) {
	basisPoints, err := dlcc.Bindings.DlSm.AdversaryThresholdBasisPoints(&bind.CallOpts{})
	if err != nil {
		dlcc.Logger.Info().Msgf("ADV err %v", err)
		return nil, ErrCannotReadFromChain
	}
	basisPointsBitInt := new(big.Int).SetUint64(uint64(basisPoints))
	return basisPointsBitInt, nil
}

func (dlcc *DataLayrChainClient) GetFeeParams() (FeeParams, error) {
	feePerBytePerTime, err := dlcc.Bindings.DlSm.FeePerBytePerTime(&bind.CallOpts{})
	if err != nil {
		dlcc.Logger.Info().Msgf("FEE err %v", err)
		return FeeParams{}, ErrCannotReadFromChain
	}
	return FeeParams{
		FeePerBytePerTime: feePerBytePerTime,
	}, nil
}

func (dlcc *DataLayrChainClient) GetDurationScale() (uint32, error) {
	scale, err := dlcc.Bindings.DlSm.DURATIONSCALE(&bind.CallOpts{})
	if err != nil {
		dlcc.Logger.Err(nil).Msg("[ChainIO] Could not fetch duration scale from chain")
		return 0, err
	}
	return uint32(scale.Uint64()), nil
}
