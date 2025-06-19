package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithCoverage(t *testing.T) {
	err := benchmarkTrieAccess()
	assert.Nil(t, err)
}
