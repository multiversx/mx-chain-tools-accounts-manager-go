package main

import (
	"os"

	"github.com/ElrondNetwork/elrond-accounts-manager/config"
	"github.com/ElrondNetwork/elrond-accounts-manager/process"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/urfave/cli"
)

var (
	log = logger.GetOrCreate("proxy")

	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name:  "config",
		Usage: "The main configuration file to load",
		Value: "./config/config.toml",
	}
	indicesConfigPath = cli.StringFlag{
		Name:  "indices-path",
		Usage: "The path to the indices folder",
		Value: "./config/indices",
	}

	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
	// logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	typeFlag = cli.StringFlag{
		Name:  "type",
		Usage: "This string flag specifies the approach to use for reindexing: clone or reindex (default: reindex)",
		Value: "reindex",
	}
)

func main() {
	app := cli.NewApp()

	app.Name = "Elrond Accounts Manager"
	app.Version = "v1.0.0"
	app.Usage = "This is the entry point for starting a new Elrond accounts manager"
	app.Flags = []cli.Flag{
		configurationFile,
		logLevel,
		logSaveFile,
		typeFlag,
		indicesConfigPath,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startAccountsManager

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startAccountsManager(ctx *cli.Context) error {
	err := initializeLogger(ctx)
	if err != nil {
		return err
	}

	log.Info("Starting elrond-accounts-manager...")

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	generalConfig, err := loadMainConfig(configurationFileName)
	if err != nil {
		return err
	}

	dataProc, err := process.CreateDataProcessor(generalConfig, ctx.GlobalString(typeFlag.Name), ctx.GlobalString(indicesConfigPath.Name))
	if err != nil {
		return err
	}

	return dataProc.ProcessAccountsData()
}

func loadMainConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func initializeLogger(ctx *cli.Context) error {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	return nil
}
