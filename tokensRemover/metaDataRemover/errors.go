package main

import "errors"

var errInvalidTokenFormat = errors.New("invalid token format")

var errCouldNotConvertNonceToBigInt = errors.New("could not convert nonce to big int")
