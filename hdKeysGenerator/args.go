package main

type argsCreateTasks struct {
	numTasks        int
	startIndex      int
	useAccountIndex bool
}

type argsGenerateKeysInParallel struct {
	numTasks        int
	startIndex      int
	useAccountIndex bool
	numKeys         int
}
