package main

import (
	"encoding/json"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	logFilePrefix = "meta-data-remover"
)

func main() {
	app := cli.NewApp()
	app.Name = "Tokens exporter CLI app"
	app.Usage = "This is the entry point for the tool that deletes tokens meta-data"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startProcess(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func startProcess(c *cli.Context) error {
	flagsConfig := getFlagsConfig(c)

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig.ContextFlagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("sanity checks...")

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	log.Info("starting processing", "pid", os.Getpid())

	tokens, err := readInput(flagsConfig.Tokens)
	log.Info("read from input", "file", flagsConfig.Tokens, "num of tokens", len(tokens))

	tokensSorted := sortTokensIDByNonce(tokens)

	_ = groupTokensByIntervals(tokensSorted)
	return nil
}

func readInput(tokensFile string) (map[string]struct{}, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(workingDir, tokensFile)
	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	bytesFromJson, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	tokens := make(map[string]struct{})
	err = json.Unmarshal(bytesFromJson, &tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func sortTokensIDByNonce(tokens map[string]struct{}) map[string][]uint64 {
	ret := make(map[string][]uint64)
	for token := range tokens {
		splits := strings.Split(token, "-")
		tokenID := splits[0] + "-" + splits[1] // ticker-randSequence
		nonceBI := big.NewInt(0)
		nonceBI.SetString(splits[2], 16)

		ret[tokenID] = append(ret[tokenID], nonceBI.Uint64())
	}

	log.Info("found", "num of tokensID", len(ret))
	for tokenID, nonces := range ret {
		log.Info("found", "tokenID", tokenID, "num of nonces", len(nonces))
		sort.SliceStable(nonces, func(i, j int) bool {
			return nonces[i] < nonces[j]
		})
	}

	return ret
}

type interval struct {
	start uint64
	end   uint64
}

func groupTokensByIntervals(tokens map[string][]uint64) map[string][]*interval {
	ret := make(map[string][]*interval)

	for token, nonces := range tokens {
		numNonces := len(nonces)
		for idx := 0; idx < numNonces; idx++ {
			nonce := nonces[idx]
			if idx+1 >= numNonces {
				ret[token] = append(ret[token], &interval{
					start: nonce,
					end:   nonce,
				})
				break
			}

			currInterval := &interval{start: nonce}
			numConsecutiveNonces := uint64(0)
			for idx < numNonces-1 {
				currNonce := nonces[idx]
				nextNonce := nonces[idx+1]
				if nextNonce-currNonce > 1 {
					break
				}

				numConsecutiveNonces++
				idx++
			}

			currInterval.end = currInterval.start + numConsecutiveNonces
			ret[token] = append(ret[token], currInterval)
		}
	}

	return ret
}
