package main

import "github.com/urfave/cli"

var (
	// StateChangesDBPath defines a flag for the path to the state changes database
	StateChangesDBPath = cli.StringFlag{
		Name:  "state-changes-db-path",
		Usage: "This flag specifies the path to the state changes database",
		Value: "stateChanges",
	}
)
