## Description


# How to use

1. compile the binary by issuing a `go build` command in elrond-tools-go/balancesExporter directory
2. download a Node `db`
3. start the app as follows:

```
./balancesExporter --log-level *:DEBUG --log-save --db-path=my-db-path/1 --shard=0 --epoch=689
```
