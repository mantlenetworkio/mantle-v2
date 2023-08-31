package batcher

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
)

// txData represents the data for a single transaction.
//
// Note: The batcher currently sends exactly one frame per transaction. This
// might change in the future to allow for multiple frames from possibly
// different channels.
type txData struct {
	frame frameData
}

// ID returns the id for this transaction data. It can be used as a map key.
func (td *txData) ID() txID {
	return td.frame.id
}

// Bytes returns the transaction data. It's a version byte (0) followed by the
// concatenated frames for this transaction.
func (td *txData) Bytes() []byte {
	return append([]byte{derive.DerivationVersion0}, td.frame.data...)
}

func (td *txData) Len() int {
	return 1 + len(td.frame.data)
}

// Frame returns the single frame of this tx data.
//
// Note: when the batcher is changed to possibly send multiple frames per tx,
// this should be changed to a func Frames() []frameData.
func (td *txData) Frame() frameData {
	return td.frame
}

// txID is an opaque identifier for a transaction.
// It's internal fields should not be inspected after creation & are subject to change.
// This ID must be trivially comparable & work as a map key.
//
// Note: transactions currently only hold a single frame, so it can be
// identified by the frame. This needs to be changed once the batcher is changed
// to send multiple frames per tx.
type txID = frameID

func (id txID) String() string {
	return fmt.Sprintf("%s:%d", id.chID.String(), id.frameNumber)
}

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (id txID) TerminalString() string {
	return fmt.Sprintf("%s:%d", id.chID.TerminalString(), id.frameNumber)
}

type StoreParams struct {
	ReferenceBlockNumber uint32
	TotalOperatorsIndex  uint32
	OrigDataSize         uint32 // unique nonce for each data store
	NumTotal             uint32 // total number data node active on chain
	Quorum               uint32 // minimal amount of signatures from data node
	NumSys               uint32 // number of data node which contains the systematic chunk
	NumPar               uint32 // number of data node which contains the parity chunk
	Duration             uint32 // duration which data is stored

	// Data and Encoding
	KzgCommit      []byte // elliptic curve kzg commitmetn
	LowDegreeProof []byte
	Degree         uint32   // degree of the polynomial
	TotalSize      uint64   // total size of the data
	Order          []uint32 // mapping for deciding the storer of each coded data chunk

	// Chain
	Fee        *big.Int
	HeaderHash []byte
	Disperser  []byte
}
