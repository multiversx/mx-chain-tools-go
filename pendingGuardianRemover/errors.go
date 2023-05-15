package pendingGuardianRemover

import "errors"

// ErrNilHttpClient signals that a nil http client has been provided
var ErrNilHttpClient = errors.New("nil http client")

// ErrEmptyData signals that empty data was received
var ErrEmptyData = errors.New("empty data")
