package chain

import (
	"errors"
)

var ErrCannotGetECDSAPubKey = errors.New("ErrCannotGetECDSAPubKey")
var ErrTransactionFailed = errors.New("ErrTransactionFailed")
