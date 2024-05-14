package dln

import (
	"errors"
)

var ErrKeyAlreadyExists = errors.New("commit already exists as key in db")
var ErrKeyNotFound = errors.New("commit not found in db")
var ErrKeyExpired = errors.New("commit is expired")
var ErrKeyNotFoundOrExpired = errors.New("data is either expired or not found")
