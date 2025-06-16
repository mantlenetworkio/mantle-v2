package db

import (
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethdb"
)

func Open(path string, cache int, handles int, readOnly bool) (ethdb.Database, error) {
	chaindataPath := filepath.Join(path, "geth", "chaindata")
	ancientPath := filepath.Join(chaindataPath, "ancient")
	ldb, err := openDatabase(openOptions{
		Type:              "leveldb",
		Directory:         chaindataPath,
		AncientsDirectory: ancientPath,
		Namespace:         "",
		Cache:             cache,
		Handles:           handles,
		ReadOnly:          readOnly,
	})
	if err != nil {
		return nil, err
	}
	return ldb, nil
}
