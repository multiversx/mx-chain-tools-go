##Smart contract counter

This tool will read data from the Elasticsearch database about smart contract deploys.

## How to use

- update the Elasticsearch configuration inside `config.toml` file
- run `cd cmd/sccounter`
- run `go build`
- run `./sccounter --wasm-file="path_to_wasm_file"`