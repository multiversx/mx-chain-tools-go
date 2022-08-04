package main

type argsCreateTasks struct {
	numTasks        int
	startIndex      int
	useAccountIndex bool
	constraints     constraints
}

type argsGenerateKeysInParallel struct {
	numTasks        int
	startIndex      int
	useAccountIndex bool
	numKeys         int
	constraints     constraints
}
