package httpClientWrapper

import "errors"

// ErrNilHttpClient signals that a nil http client has been provided
var ErrNilHttpClient = errors.New("nil http client")

// ErrEmptyData signals that empty data has been provided
var ErrEmptyData = errors.New("empty data")

// ErrHTTPStatusCodeIsNotOK signals that the returned HTTP status code is not OK
var ErrHTTPStatusCodeIsNotOK = errors.New("HTTP status code is not OK")
