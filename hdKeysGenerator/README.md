# Hierarchical deterministic (HD) keys generation

This tool efficiently generates a bulk of _hierarchical deterministic_ (HD) keys from a provided mnemonic phrase, by also applying a set of constraints on the generated public keys, such as:

 - membership to an **actual network shard**
 - membership to a **projected network shard**

**Note:** the *projected shard of an account* is its containing shard, given a network with the maximum number of shards (256). In other words, the projected shard is given by the last byte of the public key.

## How to use

Compile the code as follows:

```
cd elrond-tools-go/hdKeysGenerator
go build .
```

Then, run the command as follows:

```
./hdKeysGenerator --num-keys=5000 --format=plain-json
```

You can configure the public key (address) constraints as follows:

```
# generate keys in the provided actual shard
./hdKeysGenerator --num-keys=5000 --actual-shard=2 --num-shards=3
```

```
# generate keys in the provided projected shard
./hdKeysGenerator --num-keys=5000 --projected-shard=5
```

### Export formats

When running the tool, you can specify the desired export format. The available formats are: 

`plain-text`:

```
AccountIndex AddressIndex	Address	PublicKey	SecretKey
0            5	            erd1...	aaaa        bbbb
0            8	            erd1...	cccc        dddd
```

`plain-json`:

```
[
    {
        "accountIndex": 0,
        "addressIndex": 5,
        "address": "erd1...",
        "publicKey": "aaaa",
        "secretKey": "bbbb"
    },
    ...
]
```

