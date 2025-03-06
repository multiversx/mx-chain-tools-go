## Description

This tool is used to print accounts data from a trie database.

## Input

The input required by this tool is:
- the database which contains the trie data
- the hexRootHash used to recreate the trie
- the accountsAddresses file that contains the addresses (in hex format) to be printed

## How to use

1. compile the binary by issuing a `go build` command 
2. create a `db` directory and place inside directories `0`, `1` ... that contains the state data
3. start the app with the following parameters:
   `./accountsPrinter --log-level *:DEBUG --log-save --hex-roothash c93be73e9e1d8918ea240523372bc3094aa4bbc7221000300a493a6ae593b348`
