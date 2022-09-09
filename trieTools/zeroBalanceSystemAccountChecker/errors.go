package main

import "errors"

var errCouldNotGetBalance = errors.New("could not get balance")

var errNilHttpResponse = errors.New("received nil http response")

var errNilHttpResponseBody = errors.New("received nil http response body")
