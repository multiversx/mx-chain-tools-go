package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/elasticreindexer/config"
	"github.com/multiversx/mx-chain-tools-go/elasticreindexer/elastic"
	"github.com/multiversx/mx-chain-tools-go/elasticreindexer/reader"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const (
	indicesFolder  = "indices"
	policiesFolder = "policies"
	configFileName = "cluster.toml"
)

type Cfg struct {
	ClusterConfig struct {
		URL            string   `toml:"url"`
		Username       string   `toml:"username"`
		Password       string   `toml:"password"`
		EnabledIndices []string `toml:"enabled-indices"`
		Policies       struct {
			Enable            bool     `toml:"enable"`
			IndicesWithPolicy []string `toml:"indices-with-policy"`
		} `toml:"policies"`
	} `toml:"config"`
}

var (
	log = logger.GetOrCreate("main")

	// defines the path to the config folder
	configPath = cli.StringFlag{
		Name:  "config-path",
		Usage: "The path to the config folder",
		Value: "./config",
	}
)

const helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Index cr"
	app.Version = "v1.0.0"
	app.Usage = "Elasticsearch indices creator tool"
	app.Flags = []cli.Flag{
		configPath,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	_ = logger.SetLogLevel("*:DEBUG")

	app.Action = createIndexesAndMappings

	err := app.Run(os.Args)
	if err != nil {
		log.Error("cannot run app", "error", err)
		os.Exit(1)
	}

}

func createIndexesAndMappings(ctx *cli.Context) {
	cfgPath := ctx.String(configPath.Name)
	cfg, err := loadConfigFile(cfgPath)
	if err != nil {
		log.Error("cannot load config file", "error", err)
		return
	}

	pathToMappings := path.Join(cfgPath, indicesFolder)

	indexTemplateMap, err := reader.GetElasticTemplates(pathToMappings, cfg.ClusterConfig.EnabledIndices)
	if err != nil {
		log.Error("cannot load templates", "error", err)
		return
	}

	err = createIndies(cfg, indexTemplateMap)
	if err != nil {
		log.Error("cannot create indices", "error", err)
		return
	}

	pathToPolicies := path.Join(pathToMappings, policiesFolder)
	err = createPoliciesIfEnabled(cfg, pathToPolicies)
	if err != nil {
		log.Error("cannot create indices policies", "error", err)
	}

	log.Info("all indices were created")
}

func createIndies(cfg *Cfg, indexTemplateMap map[string]*bytes.Buffer) error {
	databaseClient, err := elastic.NewElasticClient(config.ElasticInstanceConfig{
		URL:      cfg.ClusterConfig.URL,
		Username: cfg.ClusterConfig.Username,
		Password: cfg.ClusterConfig.Password,
	})
	if err != nil {
		return err
	}

	for index, indexData := range indexTemplateMap {
		doesTemplateExists := databaseClient.DoesTemplateExist(index)
		if !doesTemplateExists {
			errCheck := databaseClient.PutIndexTemplate(index, indexData)
			if errCheck != nil {
				return fmt.Errorf("databaseClient.CreateIndexWithMapping index: %s, error: %w", index, errCheck)
			}

			log.Info("databaseClient.PutIndexTemplate", "index", index)
		}

		indexWithSuffix := fmt.Sprintf("%s-%s", index, "000001")
		alreadyExists := databaseClient.DoesIndexExist(index)
		if !alreadyExists {
			errCreate := databaseClient.CreateIndexWithMapping(indexWithSuffix, nil)
			if errCreate != nil {
				return fmt.Errorf("databaseClient.CreateIndexWithMapping index: %s, error: %w", index, errCreate)
			}

			log.Info("databaseClient.CreateIndexWithMapping", "index", index)
		}

		aliasExists := databaseClient.DoesAliasExist(index)
		if !aliasExists {
			errAlias := databaseClient.PutAlias(indexWithSuffix, index)
			if err != nil {
				return fmt.Errorf("databaseClient.PutAlias index: %s, error: %w", index, errAlias)
			}

			log.Info("databaseClient.PutAlias", "index", index)
		}
	}

	return nil
}

func createPoliciesIfEnabled(cfg *Cfg, pathToPolicies string) error {
	if !cfg.ClusterConfig.Policies.Enable {
		return nil
	}

	databaseClient, err := elastic.NewElasticClient(config.ElasticInstanceConfig{
		URL:      cfg.ClusterConfig.URL,
		Username: cfg.ClusterConfig.Username,
		Password: cfg.ClusterConfig.Password,
	})
	if err != nil {
		return err
	}

	indexPolicyMap, err := reader.GetElasticTemplates(pathToPolicies, cfg.ClusterConfig.Policies.IndicesWithPolicy)
	if err != nil {
		return err
	}

	for index, policy := range indexPolicyMap {
		indexWithSuffix := fmt.Sprintf("%s-%s", index, "000001")
		err = databaseClient.SetWriteIndexTrue(index, indexWithSuffix)
		if err != nil {
			return fmt.Errorf("databaseClient.SetWriteIndexTrue index: %s, error: %w", index, err)
		}
		log.Info("databaseClient.SetWriteIndexTrue", "index", index)

		policyName := fmt.Sprintf("%s-%s", index, "policy")
		err = databaseClient.PutPolicy(policyName, policy)
		if err != nil {
			return fmt.Errorf("databaseClient.PutPolicy index: %s, error: %w", index, err)
		}

		log.Info("databaseClient.PutPolicy", "index", index)
	}

	return nil
}

func loadConfigFile(pathStr string) (*Cfg, error) {
	tomlBytes, err := loadBytesFromFile(path.Join(pathStr, configFileName))
	if err != nil {
		return nil, err
	}

	var cfg Cfg
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBytesFromFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}
