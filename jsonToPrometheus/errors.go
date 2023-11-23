package jsonToPrometheus

import "errors"

// ErrNilHTTPClientWrapper signals that a nil HTTP client wrapper was provided
var ErrNilHTTPClientWrapper = errors.New("nil HTTP client wrapper")

// ErrNoKeyProvided signals that no key was provided
var ErrNoKeyProvided = errors.New("no key provided")
