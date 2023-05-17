package pendingGuardianRemover

import "errors"

// ErrNilHttpClient signals that a nil http client has been provided
var ErrNilHttpClient = errors.New("nil http client")

// ErrEmptyData signals that empty data has been provided
var ErrEmptyData = errors.New("empty data")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilTXBuilder signals that a nil transaction builder has been provided
var ErrNilTXBuilder = errors.New("nil tx builder")

// ErrNilCryptoHolder signals that a nil crypto holder has been provided
var ErrNilCryptoHolder = errors.New("nil crypto holder")
