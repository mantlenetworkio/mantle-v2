module github.com/Layr-Labs/datalayr/lib/merkzg

go 1.18

replace github.com/Layr-Labs/datalayr/common => ../../common

require (
	github.com/Layr-Labs/datalayr/common v0.0.0-00010101000000-000000000000
	github.com/ethereum/go-ethereum v1.10.25
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	golang.org/x/crypto v0.0.0-20220817201139-bc19a97f63c8 // indirect
	golang.org/x/sys v0.0.0-20220727055044-e65921a090b8 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
