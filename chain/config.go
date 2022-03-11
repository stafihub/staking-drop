package chain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/urfave/cli/v2"

	log "github.com/ChainSafe/log15"
)

const (
	defaultConfigPath   = "./config.json"
	defaultKeystorePath = "./keys"
)

var (
	ConfigFileFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "json configuration file",
		Value: defaultConfigPath,
	}

	VerbosityFlag = &cli.StringFlag{
		Name:  "verbosity",
		Usage: "supports levels crit (silent) to trce (trace)",
		Value: log.LvlInfo.String(),
	}

	KeystorePathFlag = &cli.StringFlag{
		Name:  "keystore",
		Usage: "path to keystore directory",
		Value: defaultKeystorePath,
	}
)

type ConfigOption struct {
	Endpoint       string              `json:"endpoint"` // url for rpc endpoint
	KeystorePath   string              `json:"keystorePath"`
	BlockstorePath string              `json:"blockstorePath"`
	StartBlock     int                 `json:"startBlock"`
	Account        string              `json:"account"`
	GasPrice       string              `json:"gasPrice"`
	DropInfos      map[string]DropInfo `json:"dropInfos"`
}

type DropInfo struct {
	MinBondAmount types.Int `json:"minBondAmount"`
	DropAmount    types.Int `json:"dropAmount"`
}

func GetConfig(ctx *cli.Context) (*ConfigOption, error) {
	var cfg ConfigOption
	path := defaultConfigPath
	if file := ctx.String(ConfigFileFlag.Name); file != "" {
		path = file
	}
	err := loadConfig(path, &cfg)
	if err != nil {
		log.Warn("err loading json file", "err", err.Error())
		return nil, err
	}
	log.Debug("Loaded config", "path", path)
	return &cfg, nil
}

func loadConfig(file string, config *ConfigOption) (err error) {
	ext := filepath.Ext(file)
	fp, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	log.Debug("Loading configuration", "path", filepath.Clean(fp))

	f, err := os.Open(filepath.Clean(fp))
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	if ext != ".json" {
		return fmt.Errorf("unrecognized extention: %s", ext)
	}
	return json.NewDecoder(f).Decode(&config)
}
