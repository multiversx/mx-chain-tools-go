module github.com/ElrondNetwork/elrond-tools-go/trieTools

go 1.17

require (
	github.com/ElrondNetwork/elrond-go v1.3.35
	github.com/ElrondNetwork/elrond-go-core v1.1.16-0.20220622092812-cc07c736c3b8
	github.com/ElrondNetwork/elrond-go-logger v1.0.7
	github.com/ElrondNetwork/elrond-tools-go/elasticreindexer v0.0.0-20220804132358-3c9fff83423c
	github.com/ElrondNetwork/elrond-vm-common v1.3.12
	github.com/pelletier/go-toml v1.9.3
	github.com/stretchr/testify v1.7.1
	github.com/tidwall/gjson v1.14.0
	github.com/urfave/cli v1.22.9
)

require (
	github.com/ElrondNetwork/concurrent-map v0.1.3 // indirect
	github.com/ElrondNetwork/elrond-go-crypto v1.0.1 // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisbrodbeck/machineid v1.0.1 // indirect
	github.com/elastic/go-elasticsearch/v7 v7.12.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/gogo/protobuf => github.com/ElrondNetwork/protobuf v1.3.2

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.40 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.40

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.40 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.40

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.54-rc3 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.54-rc3
