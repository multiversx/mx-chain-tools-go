package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/urfave/cli"

	"github.com/multiversx/mx-chain-tools-go/netComparator/core/domain"
)

const (
	lowNumberOfTransactions             = 1000
	mediumNumberOfTransactions          = 5000
	maximumNumberOfTransactions         = 10000
	maximumNumberOfTransactionsPerBatch = 50

	// Depending on how many transactions one wishes to retrieve. The number of tries needed will increase.
	// See calculateRetryAttempts()
	minimumNumberOfRetries = 10
	mediumNumberOfRetries  = 15
	maximumNumberOfRetries = 20

	txEndpoint  = "transactions/%s"
	txsEndpoint = "transactions?after=%d&size=%d&order=asc&withScResults=true&withOperations=true&withLogs=true"
)

type wrappedTxHashes struct {
	TxHash string `json:"txHash"`
}

type wrappedDifferences struct {
	TxHash      string           `json:"txHash"`
	Differences map[string][]any `json:"differences"`
	Error       string           `json:"error"`
}

type wrappedProxy interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
}

var (
	//go:embed assets/template.html
	fs embed.FS

	primaryProxy, secondaryProxy wrappedProxy
)

func main() {
	app := cli.NewApp()
	app.Name = "Network Comparator CLI app"
	app.Usage = "This is the entry point for the tool that compares transactions between 2 networks."
	app.Flags = getFlags()
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
}

func action(c *cli.Context) error {
	config := getFlagsConfig(c)

	// The txsEndpoint has a certain limit of transactions it can provide given how complex the query is.
	// 50 is the maximum it can do in a request with the requested complexity. That's why we will batch them
	// in order to retrieve all of them.
	batches := make([]uint, 0)
	calculateBatches(uint(config.Number), &batches)

	retryConfig := []retry.Option{
		retry.OnRetry(func(n uint, err error) {
			log.Info("Retry request", n+1, err)
		}),
		retry.Attempts(10),
		retry.Delay(10 * time.Second),
	}

	// Create a client for the mainnet.
	var err error
	primaryProxy, secondaryProxy, err = newProxies(config.PrimaryURL, config.SecondaryURL)
	if err != nil {
		return fmt.Errorf("failed to create proxies: %v", err)
	}

	// Convert epoch timestamp to actual date.
	timestampTime, err := strconv.ParseInt(config.Timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %v", err)
	}
	tm := time.Unix(timestampTime, 0)
	log.Info(fmt.Sprintf("retrieving %d transactions starting from %v", config.Number, tm))

	primaryTransactions := make([]*domain.Transaction, 0)

	// Retrieve n transactions after a specified timestamp and put them in a slice.
	var transactionsResponse []byte
	for _, size := range batches {
		transactions := make([]*domain.Transaction, 0)
		endpoint := fmt.Sprintf(txsEndpoint, timestampTime, size)

		err = retry.Do(func() error {
			transactionsResponse, _, err = primaryProxy.GetHTTP(context.Background(), endpoint)
			if err != nil {
				return fmt.Errorf("failed to retrieve transactions: %w", err)
			}

			err = json.Unmarshal(transactionsResponse, &transactions)
			if err != nil {
				return fmt.Errorf("failed to marshall transactions from primary-url: %w", err)
			}

			primaryTransactions = append(primaryTransactions, transactions...)
			timestampTime = int64(primaryTransactions[len(primaryTransactions)-1].Timestamp)

			return nil
		}, retryConfig...)

		if err != nil {
			return fmt.Errorf("failed to retrieve %q transactions from primary-url: %w", config.Number, err)
		}
	}

	wrappedDiffs, secondaryTransactions := compareTransactions(primaryTransactions, retryConfig)

	// Generate an HTML report based on the differences found.
	err = generateOutputReport(primaryTransactions, secondaryTransactions, wrappedDiffs, config.OutDirectory)
	if err != nil {
		return fmt.Errorf("failed to generate output directory: %v", err)
	}

	return nil
}

func calculateBatches(number uint, batches *[]uint) {
	if number == 0 {
		return
	}

	if number <= maximumNumberOfTransactionsPerBatch {
		*batches = append(*batches, number)
		return
	}

	*batches = append(*batches, maximumNumberOfTransactionsPerBatch)
	calculateBatches(number-maximumNumberOfTransactionsPerBatch, batches)
}

func newProxies(primaryUrl, secondaryUrl string) (wrappedProxy, wrappedProxy, error) {
	primary := blockchain.ArgsProxy{
		ProxyURL:            primaryUrl,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	secondary := blockchain.ArgsProxy{
		ProxyURL:            secondaryUrl,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	// Create a client for the primary network.
	pp, err := blockchain.NewProxy(primary)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create primary proxy: %v", err)
	}

	// Create a client for the secondary network.
	sp, err := blockchain.NewProxy(secondary)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create secondary proxy: %v", err)
	}

	return pp, sp, nil
}

func calculateRetryAttempts(n int) (retriesNo uint) {
	if n < lowNumberOfTransactions {
		return minimumNumberOfRetries
	}
	if n < mediumNumberOfTransactions {
		return mediumNumberOfRetries
	}

	return maximumNumberOfRetries
}

func compareTransaction(wrappedDiffs []wrappedDifferences, txM *domain.Transaction, secondaryTransaction []*domain.Transaction, i int, wg *sync.WaitGroup, retryConfig []retry.Option) {
	defer wg.Done()

	// Get transactions from both networks and then compares all the fields contained within the struct in a retry loop.
	err := retry.Do(
		func() error {
			txS, wd, err := getTransaction(txM.TxHash, "secondary", secondaryProxy)
			if err != nil {
				return err
			}

			if wd != nil {
				wrappedDiffs[i] = *wd
				return nil
			}

			secondaryTransaction[i] = txS

			wrappedDiffs[i] = getDifference(txM.TxHash, *txM, *txS)
			return nil

		}, retryConfig...,
	)

	if err != nil {
		log.Error(err.Error())
	}
}

func compareTransactions(
	allPrimaryTransactions []*domain.Transaction,
	retryConfig []retry.Option,
) ([]wrappedDifferences, []*domain.Transaction) {
	// Iterate over the transactions hashes and fetch all the information contained in both networks
	// and compare each field respectively.
	wg := sync.WaitGroup{}
	wrappedDiffs := make([]wrappedDifferences, len(allPrimaryTransactions))
	secondaryTransactions := make([]*domain.Transaction, len(allPrimaryTransactions))

	for i, t := range allPrimaryTransactions {
		wg.Add(1)
		compareTransaction(wrappedDiffs, t, secondaryTransactions, i, &wg, retryConfig)
	}

	wg.Wait()
	return wrappedDiffs, secondaryTransactions
}

func getDifference(txHash string, t1, t2 domain.Transaction) wrappedDifferences {
	diff := wrappedDifferences{TxHash: txHash}
	diffMap := make(map[string][]any)

	structType := reflect.TypeOf(t1)

	structVal1 := reflect.ValueOf(t1)
	structVal2 := reflect.ValueOf(t2)
	fieldNum := structVal1.NumField()

	// Iterate over all the fields.
	for i := 0; i < fieldNum; i++ {
		fieldName := structType.Field(i).Name
		value1 := structVal1.Field(i).Interface()
		value2 := structVal2.Field(i).Interface()

		// If the structure are equal skip.
		if reflect.DeepEqual(value1, value2) {
			continue
		}

		// Store in the differences map, both values and their field.
		slice := []any{value1, value2}
		diffMap[fieldName] = slice
	}

	if len(diffMap) > 0 {
		diff.Differences = diffMap
	}

	return diff
}

func getTransaction(txHash, networkName string, p wrappedProxy) (*domain.Transaction, *wrappedDifferences, error) {
	//Retrieve transaction from network.
	resp, code, err := p.GetHTTP(context.Background(), fmt.Sprintf(txEndpoint, txHash))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tx %q from %q: %v", txHash, networkName, err)
	}

	wrappedErr := struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error"`
		Code  string      `json:"code"`
	}{}

	// If the status code is different from 200, further investigate.
	if code != http.StatusOK {
		switch code {

		// If the status is 404, we don't want to retry looking for it. It is the only case where we
		// also return a wrappedDifferences struct with the not found error.
		case http.StatusNotFound:
			err = json.Unmarshal(resp, &wrappedErr)
			if err != nil {
				wd := &wrappedDifferences{TxHash: txHash, Error: err.Error()}
				return nil, wd, nil
			}

		// If the status is 429, that means we are sending way too many requests at the moment. We return an error
		// in order to further retry looking for the transaction.
		case http.StatusTooManyRequests:
			tooManyReqErr := errors.New(fmt.Sprintf("too many requests: %q", txHash))
			return nil, nil, tooManyReqErr

		// If the code is something else, we keep looking for the transaction.
		default:
			return nil, nil, errors.New(fmt.Sprintf("got %d while trying to retrieve transaction %q", code, txHash))
		}
	}

	var tx domain.Transaction
	err = json.Unmarshal(resp, &tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshall tx %q from %q: %v", txHash, networkName, err)
	}

	return &tx, nil, nil
}

func generateOutputReport(
	primaryTransactions []*domain.Transaction,
	secondaryTransactions []*domain.Transaction,
	wrappedDiffs []wrappedDifferences,
	outDirectory string,
) error {
	var outPath = "./"

	// If no output dir has been provided, write the files in the current directory.
	if outDirectory != "" {
		err := os.MkdirAll(outDirectory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		outPath = outPath + outDirectory + "/"
	}

	err := generateHTMLReport(wrappedDiffs, outPath+"index.html")
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	err = generateJSONReport(primaryTransactions, outPath+"primaryTransactions.json")
	if err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	err = generateJSONReport(secondaryTransactions, outPath+"secondaryTransactions.json")
	if err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	return nil
}

func generateHTMLReport(wrappedDiffs []wrappedDifferences, outPath string) error {
	// Retrieve the template.
	tmpl, err := template.ParseFS(fs, "assets/template.html")
	if err != nil {
		panic(err)
	}

	// Create the output file.
	file, err := os.Create(outPath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}

	// Execute the template.
	err = tmpl.Execute(file, wrappedDiffs)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	return nil
}

func generateJSONReport(transactions []*domain.Transaction, outPath string) error {
	rankingsJson, _ := json.Marshal(transactions)
	err := os.WriteFile(outPath, rankingsJson, 0644)
	if err != nil {
		return fmt.Errorf("failed to create json report %q: %w", outPath, err)
	}

	return nil
}
