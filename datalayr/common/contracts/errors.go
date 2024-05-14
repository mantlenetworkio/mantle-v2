package contracts

import (
	"errors"
)

var ErrCannotReadFromChain = errors.New("failed to read from chain")

var ErrCannotGetECDSAPubKey = errors.New("failed to calculate ecdsa pubkey")

var ErrKeyAlreadyExists = errors.New("commit already exists as key in db")
var ErrKeyNotFound = errors.New("commit not found in db")

// used when contract view functions don't behave expected
var ErrInvalidContractResponse = errors.New("invalid response recieved from contract")

var ErrHeaderInconsistentLength = errors.New("header has inconsistent length")
