package dsl

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum-optimism/optimism/op-service/apis"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
)

const (
	mantleDepositGasLimit    uint32 = 300_000
	mantleWithdrawalGasLimit        = 100_000

	mantleMessagePasserABIJSON = `[
		{"anonymous":false,"inputs":[
			{"indexed":true,"internalType":"uint256","name":"nonce","type":"uint256"},
			{"indexed":true,"internalType":"address","name":"sender","type":"address"},
			{"indexed":true,"internalType":"address","name":"target","type":"address"},
			{"indexed":false,"internalType":"uint256","name":"mntValue","type":"uint256"},
			{"indexed":false,"internalType":"uint256","name":"ethValue","type":"uint256"},
			{"indexed":false,"internalType":"uint256","name":"gasLimit","type":"uint256"},
			{"indexed":false,"internalType":"bytes","name":"data","type":"bytes"},
			{"indexed":false,"internalType":"bytes32","name":"withdrawalHash","type":"bytes32"}
		],"name":"MessagePassed","type":"event"}
	]`

	mantlePortalABIJSON = `[
		{"type":"function","name":"L2_ORACLE","inputs":[],"outputs":[{"name":"","type":"address"}],"stateMutability":"view"},
		{"type":"function","name":"proveWithdrawalTransaction","inputs":[
			{"name":"_tx","type":"tuple","components":[
				{"name":"nonce","type":"uint256"},
				{"name":"sender","type":"address"},
				{"name":"target","type":"address"},
				{"name":"mntValue","type":"uint256"},
				{"name":"ethValue","type":"uint256"},
				{"name":"gasLimit","type":"uint256"},
				{"name":"data","type":"bytes"}
			]},
			{"name":"_l2OutputIndex","type":"uint256"},
			{"name":"_outputRootProof","type":"tuple","components":[
				{"name":"version","type":"bytes32"},
				{"name":"stateRoot","type":"bytes32"},
				{"name":"messagePasserStorageRoot","type":"bytes32"},
				{"name":"latestBlockhash","type":"bytes32"}
			]},
			{"name":"_withdrawalProof","type":"bytes[]"}
		],"outputs":[]},
		{"type":"function","name":"finalizeWithdrawalTransaction","inputs":[
			{"name":"_tx","type":"tuple","components":[
				{"name":"nonce","type":"uint256"},
				{"name":"sender","type":"address"},
				{"name":"target","type":"address"},
				{"name":"mntValue","type":"uint256"},
				{"name":"ethValue","type":"uint256"},
				{"name":"gasLimit","type":"uint256"},
			{"name":"data","type":"bytes"}
			]}
		],"outputs":[]}
	]`

	mantleL2OutputOracleABIJSON = `[
		{"type":"function","name":"getL2Output","inputs":[{"name":"_l2OutputIndex","type":"uint256"}],"outputs":[
			{"name":"outputRoot","type":"bytes32"},
			{"name":"timestamp","type":"uint128"},
			{"name":"l2BlockNumber","type":"uint128"}
		],"stateMutability":"view"},
		{"type":"function","name":"getL2OutputIndexAfter","inputs":[{"name":"_l2BlockNumber","type":"uint256"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
		{"type":"function","name":"FINALIZATION_PERIOD_SECONDS","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"}
	]`
)

type MantleBridge struct {
	commonImpl
	standard             *StandardBridge
	l1PortalAddr         common.Address
	l1StandardBridgeAddr common.Address
	rollupCfg            *rollup.Config
	l1Client             *L1ELNode
	l2Client             apis.EthClient
	portalABI            abi.ABI
	messagePasserABI     abi.ABI
	l2OutputOracleAddr   common.Address
	l2OutputOracleABI    abi.ABI
}

func NewMantleBridge(t devtest.T, l2Network *L2Network, supervisor *Supervisor, l1EL *L1ELNode) *MantleBridge {
	standard := NewStandardBridge(t, l2Network, supervisor, l1EL)
	portalABI := mustParseABI(t, mantlePortalABIJSON)
	messagePasserABI := mustParseABI(t, mantleMessagePasserABIJSON)
	l2OutputOracleABI := mustParseABI(t, mantleL2OutputOracleABIJSON)

	bridge := &MantleBridge{
		commonImpl:           commonFromT(t),
		standard:             standard,
		l1PortalAddr:         standard.l1PortalAddr,
		l1StandardBridgeAddr: l2Network.Escape().Deployment().L1StandardBridgeProxyAddr(),
		rollupCfg:            standard.rollupCfg,
		l1Client:             l1EL,
		l2Client:             l2Network.inner.L2ELNode(match.FirstL2EL).EthClient(),
		portalABI:            portalABI,
		messagePasserABI:     messagePasserABI,
		l2OutputOracleABI:    l2OutputOracleABI,
	}
	bridge.l2OutputOracleAddr = bridge.readL2OutputOracleAddr()
	return bridge
}

func (b *MantleBridge) WithdrawalDelay() time.Duration {
	secs, err := b.finalizationPeriodSeconds()
	b.require.NoError(err, "failed to read L2OutputOracle finalization period")
	return time.Duration(secs) * time.Second
}

func (b *MantleBridge) L1GasCost(rcpt *types.Receipt) eth.ETH {
	return b.standard.gasCost(rcpt, b.l1Client.EthClient())
}

func (b *MantleBridge) L2GasCost(rcpt *types.Receipt) eth.ETH {
	var blockTimestamp *uint64
	if hasOperatorFee(rcpt) {
		blockTimestamp = b.standard.receiptTimestamp(rcpt, b.l2Client)
	}
	return gasCost(rcpt, b.rollupCfg, blockTimestamp)
}

type l2OutputProposal struct {
	OutputRoot   [32]byte `abi:"outputRoot"`
	Timestamp    *big.Int `abi:"timestamp"`
	L2BlockNumber *big.Int `abi:"l2BlockNumber"`
}

type mantleL2Output struct {
	Index        *big.Int
	OutputRoot   common.Hash
	L2BlockNumber uint64
}

func (b *MantleBridge) callL1(abi abi.ABI, to common.Address, method string, args ...interface{}) ([]byte, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(b.ctx, DefaultTimeout)
	defer cancel()
	return b.l1Client.EthClient().Call(ctx, ethereum.CallMsg{To: &to, Data: data}, rpc.LatestBlockNumber)
}

func (b *MantleBridge) readL2OutputOracleAddr() common.Address {
	out, err := b.callL1(b.portalABI, b.l1PortalAddr, "L2_ORACLE")
	b.require.NoError(err, "failed to read L2_ORACLE from portal")
	values, err := b.portalABI.Unpack("L2_ORACLE", out)
	b.require.NoError(err, "failed to unpack L2_ORACLE from portal")
	addr, ok := values[0].(common.Address)
	b.require.True(ok, "unexpected L2_ORACLE return type %T", values[0])
	return addr
}

func (b *MantleBridge) finalizationPeriodSeconds() (uint64, error) {
	out, err := b.callL1(b.l2OutputOracleABI, b.l2OutputOracleAddr, "FINALIZATION_PERIOD_SECONDS")
	if err != nil {
		return 0, err
	}
	values, err := b.l2OutputOracleABI.Unpack("FINALIZATION_PERIOD_SECONDS", out)
	if err != nil {
		return 0, err
	}
	secs, ok := values[0].(*big.Int)
	if !ok {
		return 0, fmt.Errorf("unexpected finalization period type %T", values[0])
	}
	if !secs.IsUint64() {
		return 0, fmt.Errorf("finalization period overflows uint64: %v", secs)
	}
	return secs.Uint64(), nil
}

func (b *MantleBridge) l2OutputIndexAfter(l2BlockNumber *big.Int) (*big.Int, error) {
	out, err := b.callL1(b.l2OutputOracleABI, b.l2OutputOracleAddr, "getL2OutputIndexAfter", l2BlockNumber)
	if err != nil {
		return nil, err
	}
	values, err := b.l2OutputOracleABI.Unpack("getL2OutputIndexAfter", out)
	if err != nil {
		return nil, err
	}
	index, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected L2 output index type %T", values[0])
	}
	return index, nil
}

func (b *MantleBridge) l2OutputAt(index *big.Int) (l2OutputProposal, error) {
	out, err := b.callL1(b.l2OutputOracleABI, b.l2OutputOracleAddr, "getL2Output", index)
	if err != nil {
		return l2OutputProposal{}, err
	}
	var proposal l2OutputProposal
	if err := b.l2OutputOracleABI.UnpackIntoInterface(&proposal, "getL2Output", out); err != nil {
		return l2OutputProposal{}, err
	}
	return proposal, nil
}

func (b *MantleBridge) forOutputPublished(l2BlockNumber *big.Int) mantleL2Output {
	var output mantleL2Output
	b.require.Eventually(func() bool {
		index, err := b.l2OutputIndexAfter(l2BlockNumber)
		if err != nil {
			return false
		}
		proposal, err := b.l2OutputAt(index)
		if err != nil || proposal.L2BlockNumber == nil {
			return false
		}
		if !proposal.L2BlockNumber.IsUint64() {
			b.require.Fail("L2 output block number overflows uint64", "value", proposal.L2BlockNumber)
		}
		output = mantleL2Output{
			Index:        index,
			OutputRoot:   common.Hash(proposal.OutputRoot),
			L2BlockNumber: proposal.L2BlockNumber.Uint64(),
		}
		return true
	}, 60*time.Second, time.Second, "L2 output not yet proposed")
	return output
}

type MantleDeposit struct {
	bridge    *MantleBridge
	l1Receipt *types.Receipt
}

func (d MantleDeposit) GasCost() eth.ETH {
	if d.bridge == nil {
		panic("mantle bridge reference not set on deposit")
	}
	return d.bridge.standard.gasCost(d.l1Receipt, d.bridge.l1Client.EthClient())
}

func (b *MantleBridge) DepositETH(amount eth.ETH, from *EOA) MantleDeposit {
	dep := b.standard.Deposit(amount, from)
	return MantleDeposit{bridge: b, l1Receipt: dep.l1Receipt}
}

func (b *MantleBridge) DepositMNT(amount eth.ETH, from *EOA) MantleDeposit {
	depositFn := w3.MustNewFunc("depositMNT(uint256,uint32,bytes)", "")
	calldata, err := depositFn.EncodeArgs(amount.ToBig(), mantleDepositGasLimit, []byte{})
	b.require.NoError(err, "failed to encode depositMNT calldata")

	tx := from.Transact(
		from.Plan(),
		txplan.WithTo(&b.l1StandardBridgeAddr),
		txplan.WithData(calldata),
	)
	l1Receipt := tx.Included.Value()
	b.require.Equal(types.ReceiptStatusSuccessful, l1Receipt.Status, "MNT deposit failed")

	var l2DepositTx *types.DepositTx
	for _, log := range l1Receipt.Logs {
		if dep, err := derive.UnmarshalDepositLogEvent(log); err == nil {
			l2DepositTx = dep
			break
		}
	}
	b.require.NotNil(l2DepositTx, "Could not find L2 deposit transaction in logs")

	l2DepositTxHash := types.NewTx(l2DepositTx).Hash()
	sequencingWindowDuration := time.Duration(b.rollupCfg.SeqWindowSize) * b.l1Client.EstimateBlockTime()
	var l2DepositReceipt *types.Receipt
	b.require.Eventually(func() bool {
		var err error
		l2DepositReceipt, err = b.l2Client.TransactionReceipt(b.ctx, l2DepositTxHash)
		return err == nil
	}, sequencingWindowDuration, 500*time.Millisecond, "L2 MNT deposit never found")
	b.require.Equal(types.ReceiptStatusSuccessful, l2DepositReceipt.Status, "L2 MNT deposit should succeed")

	return MantleDeposit{bridge: b, l1Receipt: l1Receipt}
}

func (b *MantleBridge) InitiateWithdrawalMNT(amount eth.ETH, from *EOA) *MantleWithdrawal {
	withdrawTx := from.Transfer(predeploys.L2ToL1MessagePasserAddr, amount)
	withdrawReceipt, err := withdrawTx.Included.Eval(b.ctx)
	b.require.NoErrorf(err, "Failed to initiate MNT withdrawal from %v for %v", from, amount)
	b.require.Equal(types.ReceiptStatusSuccessful, withdrawReceipt.Status, "initiating MNT withdrawal failed")
	return &MantleWithdrawal{
		commonImpl:  commonFromT(b.t),
		bridge:      b,
		initReceipt: withdrawReceipt,
	}
}

func (b *MantleBridge) InitiateWithdrawalETH(amount eth.ETH, target common.Address, from *EOA) *MantleWithdrawal {
	withdrawFn := w3.MustNewFunc("initiateWithdrawal(uint256,address,uint256,bytes)", "")
	calldata, err := withdrawFn.EncodeArgs(amount.ToBig(), target, big.NewInt(mantleWithdrawalGasLimit), []byte{})
	b.require.NoError(err, "failed to encode initiateWithdrawal calldata")

	withdrawTx := from.Transact(
		from.Plan(),
		txplan.WithTo(&predeploys.L2ToL1MessagePasserAddr),
		txplan.WithData(calldata),
	)
	withdrawReceipt := withdrawTx.Included.Value()
	b.require.Equal(types.ReceiptStatusSuccessful, withdrawReceipt.Status, "initiating ETH withdrawal failed")
	return &MantleWithdrawal{
		commonImpl:  commonFromT(b.t),
		bridge:      b,
		initReceipt: withdrawReceipt,
	}
}

type mantleMessagePassed struct {
	Nonce          *big.Int
	Sender         common.Address
	Target         common.Address
	MNTValue       *big.Int
	ETHValue       *big.Int
	GasLimit       *big.Int
	Data           []byte
	WithdrawalHash common.Hash
}

type mantleWithdrawalTx struct {
	Nonce    *big.Int
	Sender   common.Address
	Target   common.Address
	MNTValue *big.Int
	ETHValue *big.Int
	GasLimit *big.Int
	Data     []byte
}

type mantleProvenWithdrawalParameters struct {
	Tx                mantleWithdrawalTx
	L2OutputIndex     *big.Int
	OutputRootProof   bindings.OutputRootProof
	WithdrawalProof   [][]byte
	L2OutputBlockHash common.Hash
}

type MantleWithdrawal struct {
	commonImpl
	bridge          *MantleBridge
	initReceipt     *types.Receipt
	proveParams     mantleProvenWithdrawalParameters
	proveReceipt    *types.Receipt
	finalizeReceipt *types.Receipt
}

func (w *MantleWithdrawal) InitiateGasCost() eth.ETH {
	return w.bridge.L2GasCost(w.initReceipt)
}

func (w *MantleWithdrawal) ProveGasCost() eth.ETH {
	w.require.NotNil(w.proveReceipt, "Must have proven withdrawal before calculating gas cost")
	return w.bridge.L1GasCost(w.proveReceipt)
}

func (w *MantleWithdrawal) FinalizeGasCost() eth.ETH {
	w.require.NotNil(w.finalizeReceipt, "Must have finalized withdrawal before calculating gas cost")
	return w.bridge.L1GasCost(w.finalizeReceipt)
}

func (w *MantleWithdrawal) InitiateBlockHash() common.Hash {
	return w.initReceipt.BlockHash
}

func (w *MantleWithdrawal) Prove(user *EOA) {
	w.t.Log("proveWithdrawal: proving withdrawal...")
	params := w.proveWithdrawalParameters()

	data, err := w.bridge.portalABI.Pack(
		"proveWithdrawalTransaction",
		params.Tx,
		params.L2OutputIndex,
		params.OutputRootProof,
		params.WithdrawalProof,
	)
	w.require.NoError(err, "failed to pack proveWithdrawalTransaction calldata")

	w.require.Eventually(func() bool {
		receipt, err := w.bridge.sendPortalTx(user, data)
		if err != nil {
			w.log.Error("Failed to send prove transaction", "err", err)
			return false
		}
		w.proveParams = params
		w.proveReceipt = receipt
		return true
	}, 30*time.Second, time.Second, "Sending prove transaction")
}

func (w *MantleWithdrawal) Finalize(user *EOA) {
	data, err := w.bridge.portalABI.Pack("finalizeWithdrawalTransaction", w.proveParams.Tx)
	w.require.NoError(err, "failed to pack finalizeWithdrawalTransaction calldata")

	w.log.Info("FinalizeWithdrawal: finalizing withdrawal...")
	w.require.Eventually(func() bool {
		receipt, err := w.bridge.sendPortalTx(user, data)
		if err != nil {
			return false
		}
		w.finalizeReceipt = receipt
		return receipt.Status == types.ReceiptStatusSuccessful
	}, 60*time.Second, 100*time.Millisecond, "finalize withdrawal failed")
}

func (w *MantleWithdrawal) proveWithdrawalParameters() mantleProvenWithdrawalParameters {
	output := w.bridge.forOutputPublished(w.initReceipt.BlockNumber)

	l2Header, err := w.bridge.l2Client.InfoByNumber(w.ctx, output.L2BlockNumber)
	w.require.NoErrorf(err, "failed to fetch block header %v", output.L2BlockNumber)

	ev, err := w.bridge.parseMessagePassed(w.initReceipt)
	w.require.NoError(err, "failed to parse message passed receipt")

	return w.proveWithdrawalParametersForEvent(ev, l2Header, output)
}

func (w *MantleWithdrawal) proveWithdrawalParametersForEvent(ev *mantleMessagePassed, l2Header eth.BlockInfo, output mantleL2Output) mantleProvenWithdrawalParameters {
	withdrawalHash, err := mantleWithdrawalHash(ev)
	w.require.NoErrorf(err, "failed to calculate hash for withdrawal %v", ev)
	w.require.Equal(withdrawalHash, ev.WithdrawalHash, "computed withdrawal hash incorrectly")

	slot := withdrawals.StorageSlotOfWithdrawalHash(withdrawalHash)
	proof, err := w.bridge.l2Client.GetProof(w.ctx, predeploys.L2ToL1MessagePasserAddr, []common.Hash{slot}, hexutil.Uint64(l2Header.NumberU64()).String())
	w.require.NoErrorf(err, "failed to fetch proof for withdrawal %v", ev)
	w.require.Len(proof.StorageProof, 1, "invalid amount of storage proofs")

	err = verifyProof(l2Header.Root(), proof)
	w.require.NoErrorf(err, "failed to verify proof for withdrawal")

	trieNodes := make([][]byte, len(proof.StorageProof[0].Proof))
	for i, node := range proof.StorageProof[0].Proof {
		trieNodes[i] = node
	}

	withdrawalsRoot := l2Header.WithdrawalsRoot()
	w.require.NotNil(withdrawalsRoot, "missing withdrawals root")

	return mantleProvenWithdrawalParameters{
		Tx: mantleWithdrawalTx{
			Nonce:    ev.Nonce,
			Sender:   ev.Sender,
			Target:   ev.Target,
			MNTValue: ev.MNTValue,
			ETHValue: ev.ETHValue,
			GasLimit: ev.GasLimit,
			Data:     ev.Data,
		},
		L2OutputIndex: output.Index,
		OutputRootProof: bindings.OutputRootProof{
			Version:                  [32]byte{},
			StateRoot:                l2Header.Root(),
			MessagePasserStorageRoot: *withdrawalsRoot,
			LatestBlockhash:          l2Header.Hash(),
		},
		WithdrawalProof:   trieNodes,
		L2OutputBlockHash: l2Header.Hash(),
	}
}

func (b *MantleBridge) parseMessagePassed(receipt *types.Receipt) (*mantleMessagePassed, error) {
	event, ok := b.messagePasserABI.Events["MessagePassed"]
	if !ok {
		return nil, fmt.Errorf("MessagePassed event not in ABI")
	}

	for _, log := range receipt.Logs {
		if log.Address != predeploys.L2ToL1MessagePasserAddr {
			continue
		}
		if len(log.Topics) == 0 || log.Topics[0] != event.ID {
			continue
		}
		if len(log.Topics) < 4 {
			return nil, fmt.Errorf("MessagePassed log missing topics")
		}

		out := &mantleMessagePassed{
			Nonce:  new(big.Int).SetBytes(log.Topics[1].Bytes()),
			Sender: common.BytesToAddress(log.Topics[2].Bytes()[12:]),
			Target: common.BytesToAddress(log.Topics[3].Bytes()[12:]),
		}

		var decoded struct {
			MNTValue       *big.Int
			ETHValue       *big.Int
			GasLimit       *big.Int
			Data           []byte
			WithdrawalHash common.Hash
		}
		if err := b.messagePasserABI.UnpackIntoInterface(&decoded, "MessagePassed", log.Data); err != nil {
			return nil, fmt.Errorf("failed to unpack MessagePassed log: %w", err)
		}

		out.MNTValue = decoded.MNTValue
		out.ETHValue = decoded.ETHValue
		out.GasLimit = decoded.GasLimit
		out.Data = decoded.Data
		out.WithdrawalHash = decoded.WithdrawalHash
		return out, nil
	}

	return nil, fmt.Errorf("MessagePassed event not found")
}

func mantleWithdrawalHash(ev *mantleMessagePassed) (common.Hash, error) {
	args := abi.Arguments{
		{Name: "nonce", Type: withdrawals.Uint256Type},
		{Name: "sender", Type: withdrawals.AddressType},
		{Name: "target", Type: withdrawals.AddressType},
		{Name: "mntValue", Type: withdrawals.Uint256Type},
		{Name: "ethValue", Type: withdrawals.Uint256Type},
		{Name: "gasLimit", Type: withdrawals.Uint256Type},
		{Name: "data", Type: withdrawals.BytesType},
	}
	enc, err := args.Pack(ev.Nonce, ev.Sender, ev.Target, ev.MNTValue, ev.ETHValue, ev.GasLimit, ev.Data)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack for withdrawal hash: %w", err)
	}
	return crypto.Keccak256Hash(enc), nil
}

func (b *MantleBridge) sendPortalTx(user *EOA, data []byte) (*types.Receipt, error) {
	tx := txplan.NewPlannedTx(
		user.Plan(),
		txplan.WithTo(&b.l1PortalAddr),
		txplan.WithData(data),
	)
	receipt, err := tx.Included.Eval(b.ctx)
	if err != nil {
		return nil, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("tx failed with status %d", receipt.Status)
	}
	return receipt, nil
}

func mustParseABI(t devtest.T, jsonABI string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(jsonABI))
	t.Require().NoError(err)
	return parsed
}
