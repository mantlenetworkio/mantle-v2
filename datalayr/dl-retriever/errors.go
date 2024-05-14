package retriever

import (
	"errors"
)

var ErrRetrieveReply_CANT_GET_INITDATASTORE_EVENT = errors.New("ErrRetrieveReply_CANT_GET_INITDATASTORE_EVENT")
var ErrRetrieveReply_NONEXIST = errors.New("ErrRetrieveReply_NONEXIST")
var ErrRetrieveReply_EXPIRED = errors.New("ErrRetrieveReply_EXPIRED")
var ErrRetrieveReply_INTERNALERR = errors.New("ErrRetrieveReply_INTERNALERR")
var ErrRetrieveReply_DECODEERR = errors.New("ErrRetrieveReply_DECODEERR")
