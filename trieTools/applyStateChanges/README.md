## Description

This tool is used to check that the collected state changes are correct
# How to use

1. compile the binary by issuing a `go build` command in mx-chain-tools-go/trieTools/applyStateChanges directory
2. create a `db` directory and place inside directories that contains the starting trie data
3. copy the `stateChanges` directory from the collector
4. copy the `headerHashes` file from the collector
5. start the app with the following parameters: `./applyStateChanges -log-level *:DEBUG -log-save -hex-roothash bd87855838187d359398243f03dfd7d896d7cca7542a559c518339ce22027b85`, where -hex-rootHash is the trie root hash from where the state changes will start to be applied
