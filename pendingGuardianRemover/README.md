## Pending guardian remover

This tool has two main binaries that can be used separate.

### Transactions creator binary
First binary generates a configurable number of co-signed transactions, and writes them to a json file.

The generated file is formatted as a map with the transaction nonce as key and the actual transaction as value.

It needs two pem files, first one for the sender signature and the second one for the guardian signature.

Its configuration can be found on `txsCreator/config.toml` file.

### Transactions sender binary
Second binary relies on the json file generated above. It parses the json and stores the transactions in an internal map.

With a separate goroutine, it checks once in a while if the user has a pending guardian on chain. If so, it makes a new request to the API in order to receive the current account nonce, then searches the transactions map for the current nonce and sends the provided transaction to API.

If transaction is not found, or any other error, it logs a warning. Otherwise, transaction hash is logged.

Its configuration can be found on `txsSender/config.toml` file.
