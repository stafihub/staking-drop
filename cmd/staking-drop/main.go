package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/ChainSafe/log15"
	"github.com/stafihub/staking-drop/chain"
	"github.com/urfave/cli/v2"
)

var app = cli.NewApp()

var mainFlags = []cli.Flag{
	chain.ConfigFileFlag,
	chain.VerbosityFlag,
}

// init initializes CLI
func init() {
	app.Action = run
	app.Copyright = "Copyright 2022 Stafi Protocol Authors"
	app.Name = "staking-dropd"
	app.Usage = "staking-dropd"
	app.Authors = []*cli.Author{{Name: "Stafi Protocol 2022"}}
	app.Version = "0.0.1"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{}

	app.Flags = append(app.Flags, mainFlags...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startLogger(ctx *cli.Context) error {
	logger := log.Root()
	var lvl log.Lvl
	if lvlToInt, err := strconv.Atoi(ctx.String(chain.VerbosityFlag.Name)); err == nil {
		lvl = log.Lvl(lvlToInt)
	} else if lvl, err = log.LvlFromString(ctx.String(chain.VerbosityFlag.Name)); err != nil {
		return err
	}

	logger.SetHandler(log.MultiHandler(
		log.LvlFilterHandler(
			lvl,
			log.StreamHandler(os.Stdout, log.LogfmtFormat())),
		log.LvlFilterHandler(
			lvl,
			log.Must.FileHandler("log.json", log.JsonFormat()),
		),
		log.LvlFilterHandler(
			log.LvlError,
			log.Must.FileHandler("log_errors.json", log.JsonFormat()))))

	return nil
}

func run(ctx *cli.Context) error {
	err := startLogger(ctx)
	if err != nil {
		return err
	}

	cfg, err := chain.GetConfig(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("config: %+v\n", cfg)

	// Used to signal core shutdown due to fatal error
	sysErr := make(chan error)

	stafiHubChain := chain.NewChain()
	logger := log.Root()
	err = stafiHubChain.Initialize(cfg, logger, sysErr)
	if err != nil {
		return err
	}

	// =============== start
	err = stafiHubChain.Start()
	if err != nil {
		return err
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	// Block here and wait for a signal
	select {
	case err := <-sysErr:
		logger.Error("FATAL ERROR. Shutting down.", "err", err)
	case <-sigc:
		logger.Warn("Interrupt received, shutting down now.")
	}
	stafiHubChain.Stop()
	return nil
}
