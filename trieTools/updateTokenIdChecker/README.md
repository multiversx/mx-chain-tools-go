## Description

This tool is used to check that the esdt type has been sent to all shards
# How to use

1. compile the binary by issuing a `go build` command in mx-chain-tools-go/trieTools/updateTokenIdChecker directory
2. create a `db` directory and place inside directories `meta`, `shard0`,`shard  ... that contains the state data
3. start the app with the following parameters: `./updateTokenIdChecker -log-level *:DEBUG -log-save -hex-roothash-0 a5c5c79601ce3d15908902081e60cc0a376937605d6271e4061f24fbbecbfb58 -hex-roothash-1 fdde12ef7ac8514149c827536aad32a09007333a342a7dc412d0c1f3c61a9cdc -hex-roothash-2 914e5f45f32a5dce3be1e7944f325d7f0ed2cacc78320478c05e4f36303651a6 -hex-roothash-meta f00091b3bd2a83ae0463cdaa1e670f35265b7a48a47c9f40cfa547c6d36a3305`
