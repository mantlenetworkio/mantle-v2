package store

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

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
		CREATE TABLE IF NOT EXISTS users (
			address TEXT PRIMARY KEY,
			registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS claims (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			address TEXT NOT NULL,
			amount_wei TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (address) REFERENCES users(address)
		);

		CREATE INDEX IF NOT EXISTS idx_claims_address_date ON claims(address, created_at);
	`)
	return err
}

// RegisterUser registers a new user by address. Returns false if already registered.
func (s *Store) RegisterUser(addr common.Address) (bool, error) {
	result, err := s.db.Exec(
		`INSERT OR IGNORE INTO users (address) VALUES (?)`,
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
		`SELECT COUNT(*) FROM users WHERE address = ?`,
		addr.Hex(),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user: %w", err)
	}
	return count > 0, nil
}

// DailyClaimedAmount returns the total amount (in wei) claimed by the address today.
func (s *Store) DailyClaimedAmount(addr common.Address) (*big.Int, error) {
	todayStart := todayUTCStart()
	var amountStr sql.NullString
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(CAST(amount_wei AS REAL)), '0') FROM claims WHERE address = ? AND created_at >= ?`,
		addr.Hex(),
		todayStart,
	).Scan(&amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily claimed amount: %w", err)
	}

	amount := new(big.Int)
	if amountStr.Valid && amountStr.String != "" {
		// Parse as float first since COALESCE(SUM(CAST(...))) returns float
		f := new(big.Float)
		f.SetString(amountStr.String)
		f.Int(amount)
	}
	return amount, nil
}

// RecordClaim records a successful claim.
func (s *Store) RecordClaim(addr common.Address, amountWei *big.Int) error {
	_, err := s.db.Exec(
		`INSERT INTO claims (address, amount_wei) VALUES (?, ?)`,
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
