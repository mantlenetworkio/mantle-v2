package store

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

// New opens a PostgreSQL connection.
// dsn example: "postgres://user:password@host:5432/dbname?sslmode=require"
func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS faucet_users (
			address TEXT PRIMARY KEY,
			registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS faucet_claims (
			id BIGSERIAL PRIMARY KEY,
			address TEXT NOT NULL,
			amount_wei NUMERIC(78,0) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_faucet_claims_address_date ON faucet_claims(address, created_at);
	`)
	return err
}

// RegisterUser registers a new user by address. Returns true if newly registered, false if already exists.
func (s *Store) RegisterUser(addr common.Address) (bool, error) {
	result, err := s.db.Exec(
		`INSERT INTO faucet_users (address) VALUES ($1) ON CONFLICT (address) DO NOTHING`,
		addr.Hex(),
	)
	if err != nil {
		return false, fmt.Errorf("failed to register user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// IsRegistered checks if a user is registered.
func (s *Store) IsRegistered(addr common.Address) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM faucet_users WHERE address = $1`,
		addr.Hex(),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user: %w", err)
	}
	return count > 0, nil
}

// DailyClaimedAmount returns the total amount (in wei) claimed by the address today (UTC).
func (s *Store) DailyClaimedAmount(addr common.Address) (*big.Int, error) {
	todayStart := todayUTCStart()
	var amountStr sql.NullString
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(amount_wei)::text, '0') FROM faucet_claims WHERE address = $1 AND created_at >= $2`,
		addr.Hex(),
		todayStart,
	).Scan(&amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily claimed amount: %w", err)
	}

	amount := new(big.Int)
	if amountStr.Valid && amountStr.String != "" {
		amount.SetString(amountStr.String, 10)
	}
	return amount, nil
}

// RecordClaim records a successful claim.
func (s *Store) RecordClaim(addr common.Address, amountWei *big.Int) error {
	_, err := s.db.Exec(
		`INSERT INTO faucet_claims (address, amount_wei) VALUES ($1, $2)`,
		addr.Hex(),
		amountWei.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to record claim: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func todayUTCStart() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
