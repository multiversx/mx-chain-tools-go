package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTasks(t *testing.T) {
	numTasks := 2
	startIndex := 1

	expectedTasks := []generatorTask{
		{
			firstIndex: startIndex + fixedTaskSize*0,
			lastIndex:  startIndex + fixedTaskSize*1,
		},
		{
			firstIndex: startIndex + fixedTaskSize*1,
			lastIndex:  startIndex + fixedTaskSize*2,
		},
	}

	tasks, newIndex := createTasks(argsCreateTasks{
		numTasks:        numTasks,
		startIndex:      startIndex,
		useAccountIndex: false,
	})
	require.Equal(t, expectedTasks, tasks)
	require.Equal(t, newIndex, startIndex+fixedTaskSize*2)
}
