package main

import "github.com/urfave/cli"

var (
	// HexRootHash0 defines a flag for the shard 0 trie root hash expressed in hex format
	HexRootHash0 = cli.StringFlag{
		Name:  "hex-roothash-0",
		Usage: "This flag specifies the roothash to start the checking from",
		Value: "",
	}
	// HexRootHash1 defines a flag for the shard 1 trie root hash expressed in hex format
	HexRootHash1 = cli.StringFlag{
		Name:  "hex-roothash-1",
		Usage: "This flag specifies the roothash to start the checking from",
		Value: "",
	}
	// HexRootHash2 defines a flag for the shard 2 trie root hash expressed in hex format
	HexRootHash2 = cli.StringFlag{
		Name:  "hex-roothash-2",
		Usage: "This flag specifies the roothash to start the checking from",
		Value: "",
	}
	// HexRootHashMeta defines a flag for the meta shard trie root hash expressed in hex format
	HexRootHashMeta = cli.StringFlag{
		Name:  "hex-roothash-meta",
		Usage: "This flag specifies the roothash to start the checking from",
		Value: "",
	}
)
