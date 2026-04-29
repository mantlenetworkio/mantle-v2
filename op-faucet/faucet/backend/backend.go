package backend

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/config"
	"github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/store"
	ftypes "github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/types"
	"github.com/ethereum-optimism/optimism/op-faucet/faucet/frontend"
	"github.com/ethereum-optimism/optimism/op-faucet/metrics"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/locks"
)

type APIRouter interface {
	AddRPC(route string) error
	AddAPI(api rpc.API) error
	AddAPIToRPC(route string, api rpc.API) error
}

type Backend struct {
	log        log.Logger
	m          metrics.Metricer
	faucets    locks.RWMap[ftypes.FaucetID, *Faucet]
	defaults   locks.RWMap[eth.ChainID, ftypes.FaucetID]
	store      *store.Store
	dailyLimit *big.Int
}

func FromConfig(log log.Logger, m metrics.Metricer, cfg *config.Config, router APIRouter) (*Backend, error) {
	b := &Backend{
		log: log,
		m:   m,
	}

	// Initialize store for user registration and rate limiting.
	// If db_dsn is empty, rate limiting is disabled.
	if cfg.DBDSN != "" {
		s, err := store.New(cfg.DBDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open store: %w", err)
		}
		b.store = s
		log.Info("Store initialized (PostgreSQL)")
	} else {
		log.Warn("db_dsn not configured, rate limiting is disabled")
	}

	if cfg.DailyLimitWei != "" {
		limit, ok := new(big.Int).SetString(cfg.DailyLimitWei, 10)
		if !ok {
			return nil, fmt.Errorf("invalid daily_limit_wei: %s", cfg.DailyLimitWei)
		}
		b.dailyLimit = limit
		log.Info("Daily claim limit configured", "limit_wei", cfg.DailyLimitWei)
	}

	var faucetIDs []ftypes.FaucetID
	for fID, fCfg := range cfg.Faucets {
		f, err := FaucetFromConfig(log, m, fID, fCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup faucet %q: %w", fID, err)
		}
		f.store = b.store
		f.dailyLimit = b.dailyLimit
		b.faucets.Set(fID, f)
		faucetIDs = append(faucetIDs, fID)
	}

	for chID, fID := range cfg.Defaults {
		if !b.faucets.Has(fID) {
			return nil, fmt.Errorf("unknown faucet, cannot set as default for chain %s: %q", chID, fID)
		}
		b.defaults.Set(chID, fID)
	}

	// Infer defaults for chains that were not explicitly mentioned.
	// Always use the lowest faucet ID, so map-iteration doesn't affect defaults.
	sort.Slice(faucetIDs, func(i, j int) bool {
		return faucetIDs[i] < faucetIDs[j]
	})
	for _, fID := range faucetIDs {
		f, _ := b.faucets.Get(fID)
		b.defaults.SetIfMissing(f.chainID, fID)
	}

	// Register shared RPCs (faucet_register, faucet_eligibility) on the root path.
	// Use the first faucet as the shared backend — register/eligibility use the shared store,
	// so any faucet instance works (they all share the same store and dailyLimit).
	var sharedFaucet *Faucet
	b.faucets.Range(func(id ftypes.FaucetID, f *Faucet) bool {
		sharedFaucet = f
		return false // stop after first
	})
	if sharedFaucet != nil {
		if err := router.AddAPI(rpc.API{
			Namespace: "faucet",
			Service:   frontend.NewSharedFrontend(sharedFaucet),
		}); err != nil {
			return nil, fmt.Errorf("failed to setup shared faucet RPC on root: %w", err)
		}
	}

	// Register per-chain RPCs (faucet_balance, faucet_requestMNT) on /chain/{chainId}.
	var chainSetupErr error
	b.defaults.Range(func(chain eth.ChainID, id ftypes.FaucetID) bool {
		f, _ := b.faucets.Get(id)
		if err := router.AddRPC("/chain/" + chain.String()); err != nil {
			chainSetupErr = errors.Join(chainSetupErr,
				fmt.Errorf("failed to setup chain route for %q: %w", chain, err))
			return true
		}
		if err := router.AddAPIToRPC("/chain/"+chain.String(), rpc.API{
			Namespace: "faucet",
			Service:   frontend.NewChainFrontend(f),
		}); err != nil {
			chainSetupErr = errors.Join(chainSetupErr,
				fmt.Errorf("failed to setup chain RPC for %q: %w", chain, err))
		}
		return true
	})
	if chainSetupErr != nil {
		return nil, fmt.Errorf("failed to set up chain route(s): %w", chainSetupErr)
	}

	return b, nil
}

// FaucetByChain gets the default faucet for the given chain.
// This returns nil if there is no such faucet configured (unknown chain or missing faucet).
func (b *Backend) FaucetByChain(id eth.ChainID) *Faucet {
	fID, ok := b.defaults.Get(id)
	if !ok {
		return nil
	}
	out, _ := b.faucets.Get(fID)
	return out
}

// FaucetByID gets the faucet by its identifier.
// This returns nil if there is no such faucet configured.
func (b *Backend) FaucetByID(id ftypes.FaucetID) *Faucet {
	out, _ := b.faucets.Get(id)
	return out
}

func (b *Backend) EnableFaucet(id ftypes.FaucetID) {
	f, ok := b.faucets.Get(id)
	if ok {
		f.Enable()
	}
}

func (b *Backend) DisableFaucet(id ftypes.FaucetID) {
	f, ok := b.faucets.Get(id)
	if ok {
		f.Disable()
	}
}

func (b *Backend) Stop(ctx context.Context) error {
	b.faucets.Range(func(key ftypes.FaucetID, value *Faucet) bool {
		value.Close()
		return true
	})
	if b.store != nil {
		if err := b.store.Close(); err != nil {
			return fmt.Errorf("failed to close store: %w", err)
		}
	}
	return nil
}


func (b *Backend) Faucets() (out map[ftypes.FaucetID]eth.ChainID) {
	out = make(map[ftypes.FaucetID]eth.ChainID)
	b.faucets.Range(func(key ftypes.FaucetID, value *Faucet) bool {
		out[key] = value.chainID
		return true
	})
	return out
}

func (b *Backend) Defaults() (out map[eth.ChainID]ftypes.FaucetID) {
	out = make(map[eth.ChainID]ftypes.FaucetID)
	b.defaults.Range(func(key eth.ChainID, value ftypes.FaucetID) bool {
		out[key] = value
		return true
	})
	return out
}
