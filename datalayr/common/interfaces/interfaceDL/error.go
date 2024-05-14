package interfaceDL

import "errors"

var ErrInsufficientFund = errors.New("Cannot get InitDataStore to Check if payment is made")
var ErrLocalLostGraphState = errors.New("I cannot get state from graph connection")
var ErrInsufficientFrames = errors.New("not enough frames to verify incorrect coding")
var ErrNonIdenticalKzgCommits = errors.New("all kzgCommits should be identical, plan is to merklize them")
var ErrVerifyFail = errors.New("incorrect RS encoding. Or wrong index")
var ErrRegistrantIdNotFound = errors.New("ErrRegistrantIdNotFound")
var ErrLocalPaused = errors.New("I am pausing my service")
var ErrLocalUnregistered = errors.New("I has not registered/ am dereged")
var ErrInconsistantTotalNodes = errors.New("ErrInconsistantTotalNodes")
var ErrLocalNotInState = errors.New("I am not in the state from graph")
var ErrDecodeDataStore = errors.New("Error decoding data store")
var ErrLocalCannotSave = errors.New("I cannot save data into database")
var ErrDecodeFrame = errors.New("Cannot decode kzgFFT frame")
var ErrSavedAlready = errors.New("frames corresponding to the msgHash is saved already")
var ErrInconsistantDegreeAndOrigDataSize = errors.New("dataStore header is invalid, inconsistant degree and OrigDataSize")
var ErrIncorrectNumberOfChunks = errors.New("incorrect number of chunks received")
