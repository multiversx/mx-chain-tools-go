package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/examples"
	"github.com/urfave/cli"
	"os"
	"reflect"
	"sync"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "Accounts Storage Exporter CLI app"
	app.Usage = "This is the entry point for the tool that exports the storage of a given account"
	//app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return action(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}

	log.Info("finished exporting the storage")
}

func action(c *cli.Context) error {
	//getFlagsConfig()=

	mainnetArgs := blockchain.ArgsProxy{
		ProxyURL:            examples.MainnetGateway,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	shfArgs := blockchain.ArgsProxy{
		ProxyURL:            examples.MainnetGateway,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	ep, err := blockchain.NewProxy(mainnetArgs)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	shf, err := blockchain.NewProxy(shfArgs)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}
	// --nonce x
	// Same hash
	// Get latest hyper block (metachain) nonce
	// !~ScResults
	// !~Logs
	// !~Events
	// !~GasFee

	nonce, err := ep.GetLatestHyperBlockNonce(context.Background())
	if err != nil {
		return fmt.Errorf("failed to retrieve latest block nonce")
	}
	log.Info("latest hyper block", "nonce", nonce)

	block, errGet := ep.GetHyperBlockByNonce(context.Background(), nonce)
	if errGet != nil {
		return fmt.Errorf("failed to retrieve hyper block: %v", err)
	}

	mu := sync.Mutex{}
	netWaitGroup := sync.WaitGroup{}
	shardWaitGroup := sync.WaitGroup{}

	blockEndpointTemplate := "block/%d/by-hash/%s?withTxs=true"

	mainMap := make(map[string][]byte)
	shadowMap := make(map[string][]byte)

	netWaitGroup.Add(2)
	// Mainnet loop
	go func() {
		for _, shardBlock := range block.ShardBlocks {
			shardWaitGroup.Add(1)
			shardBlock := shardBlock
			go func() {
				i := 10

				for i > 0 {
					var b []byte
					b, _, err = ep.GetHTTP(context.Background(), fmt.Sprintf(blockEndpointTemplate, shardBlock.Shard, shardBlock.Hash))
					if err != nil {
						panic(err)
					}

					blockInfo := struct {
						Data struct {
							Block struct {
								Hash          string `json:"hash"`
								PrevBlockHash string `json:"prevBlockHash"`
								Shard         uint32 `json:"shard"`
							} `json:"block"`
						} `json:"data"`
					}{}

					err = json.Unmarshal(b, &blockInfo)

					if _, ok := mainMap[blockInfo.Data.Block.Hash]; !ok {
						mu.Lock()
						mainMap[blockInfo.Data.Block.Hash] = b
						mu.Unlock()
					}

					shardBlock.Hash = blockInfo.Data.Block.PrevBlockHash
					i--
				}

				shardWaitGroup.Done()
			}()
		}
		shardWaitGroup.Wait()
		netWaitGroup.Done()
	}()

	// Shadow-Proxy loop
	block, errGet = shf.GetHyperBlockByNonce(context.Background(), nonce)
	if errGet != nil {
		return fmt.Errorf("failed to retrieve hyper block: %v", err)
	}
	go func() {
		for _, shardBlock := range block.ShardBlocks {
			shardWaitGroup.Add(1)
			shardBlock := shardBlock
			go func() {
				i := 10

				for i > 0 {
					var b []byte
					b, _, err = shf.GetHTTP(context.Background(), fmt.Sprintf(blockEndpointTemplate, shardBlock.Shard, shardBlock.Hash))
					if err != nil {
						panic(err)
					}

					blockInfo := struct {
						Data struct {
							Block struct {
								Hash          string `json:"hash"`
								PrevBlockHash string `json:"prevBlockHash"`
								Shard         uint32 `json:"shard"`
							} `json:"block"`
						} `json:"data"`
					}{}

					err = json.Unmarshal(b, &blockInfo)

					if _, ok := shadowMap[blockInfo.Data.Block.Hash]; !ok {
						mu.Lock()
						shadowMap[blockInfo.Data.Block.Hash] = b
						mu.Unlock()
					}

					shardBlock.Hash = blockInfo.Data.Block.PrevBlockHash
					i--
				}

				shardWaitGroup.Done()
			}()
		}
		shardWaitGroup.Wait()
		netWaitGroup.Done()
	}()

	netWaitGroup.Wait()

	if reflect.DeepEqual(mainMap, shadowMap) {
		fmt.Println("good")
	}

	return nil
}
