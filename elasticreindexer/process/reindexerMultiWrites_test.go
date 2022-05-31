package process

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeIntervals(t *testing.T) {
	res, err := computeIntervals(0, 2)
	require.Nil(t, err)
	require.NotNil(t, res)
}
