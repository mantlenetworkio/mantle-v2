package sysext

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/devnet-sdk/descriptors"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
)

const (
	ProtocolVersionsAddressName = "ProtocolVersionsProxy"
	SuperchainConfigAddressName = "SuperchainConfigProxy"

	SystemConfigAddressName   = "systemConfigProxy"
	DisputeGameFactoryName    = "disputeGameFactoryProxy"
	L1StandardBridgeProxyName = "l1StandardBridgeProxy"
	L1MNTAddressName          = "l1MantleTokenProxy"
)

type l1AddressBook struct {
	protocolVersions common.Address
	superchainConfig common.Address
}

func newL1AddressBook(t devtest.T, addresses descriptors.AddressMap) *l1AddressBook {
	// TODO(#15817) op-devstack: sysext: fix address book
	return &l1AddressBook{}
}

func (a *l1AddressBook) ProtocolVersionsAddr() common.Address {
	return a.protocolVersions
}

func (a *l1AddressBook) SuperchainConfigAddr() common.Address {
	return a.superchainConfig
}

var _ stack.SuperchainDeployment = (*l1AddressBook)(nil)

type l2AddressBook struct {
	systemConfig       common.Address
	disputeGameFactory common.Address
	l1StandardBridge   common.Address
	l1MNT              common.Address
}

func newL2AddressBook(l1Addresses descriptors.AddressMap, l1ELRPC string, logger log.Logger) *l2AddressBook {
	book := &l2AddressBook{
		systemConfig:       l1Addresses[SystemConfigAddressName],
		disputeGameFactory: l1Addresses[DisputeGameFactoryName],
		l1StandardBridge:   l1Addresses[L1StandardBridgeProxyName],
		l1MNT:              l1Addresses[L1MNTAddressName],
	}

	// Auto-resolve l1MNT from OptimismPortal.L1_MNT_ADDRESS() if not in descriptor
	if book.l1MNT == (common.Address{}) {
		portalAddr := l1Addresses["optimismPortalProxy"]
		if portalAddr != (common.Address{}) && l1ELRPC != "" {
			if resolved := queryL1MNTAddress(portalAddr, l1ELRPC, logger); resolved != (common.Address{}) {
				book.l1MNT = resolved
			}
		}
	}

	return book
}

// queryL1MNTAddress calls OptimismPortal.L1_MNT_ADDRESS() (selector 0xac6986c5)
// to resolve the MNT token address from L1.
func queryL1MNTAddress(portalAddr common.Address, l1ELRPC string, logger log.Logger) common.Address {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cl, err := ethclient.DialContext(ctx, l1ELRPC)
	if err != nil {
		logger.Warn("Failed to dial L1 for MNT address resolution", "err", err)
		return common.Address{}
	}
	defer cl.Close()

	// L1_MNT_ADDRESS() selector = 0xac6986c5
	selector := common.Hex2Bytes("ac6986c5")
	result, err := cl.CallContract(ctx, ethereum.CallMsg{
		To:   &portalAddr,
		Data: selector,
	}, nil)
	if err != nil {
		logger.Warn("Failed to query L1_MNT_ADDRESS from OptimismPortal", "portal", portalAddr, "err", err)
		return common.Address{}
	}

	if len(result) < 32 {
		logger.Warn("Unexpected response length from L1_MNT_ADDRESS", "len", len(result))
		return common.Address{}
	}

	addr := common.BytesToAddress(result[12:32])
	logger.Info("Resolved l1MNT address from OptimismPortal", "l1MNT", addr, "portal", portalAddr)
	return addr
}

func (a *l2AddressBook) SystemConfigProxyAddr() common.Address {
	return a.systemConfig
}

func (a *l2AddressBook) DisputeGameFactoryProxyAddr() common.Address {
	return a.disputeGameFactory
}

func (a *l2AddressBook) L1StandardBridgeProxyAddr() common.Address {
	return a.l1StandardBridge
}

func (a *l2AddressBook) L1MNTAddr() common.Address {
	return a.l1MNT
}

var _ stack.L2Deployment = (*l2AddressBook)(nil)
