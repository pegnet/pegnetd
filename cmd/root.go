package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pegnet/pegnetd/srv"

	"github.com/pegnet/pegnetd/exit"

	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/node"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().String("log", "info", "Change the logging level. Can choose from 'trace', 'debug', 'info', 'warn', 'error', or 'fatal'")
	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:8088/v2", "The url to the factomd endpoint witout a trailing slash")
	rootCmd.PersistentFlags().StringP("wallet", "w", "http://localhost:8089", "The url to the factomd-wallet endpoint witout a trailing slash")
	rootCmd.PersistentFlags().String("api", "8070", "Change the api listening port for the api")
}

// Execute is cobra's entry point
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:              "pegnetd",
	Short:            "pegnetd is the pegnet daemon to track balances/conversion/transactions",
	PersistentPreRun: always,
	PreRun:           ReadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle ctl+c
		ctx, cancel := context.WithCancel(context.Background())
		exit.GlobalExitHandler.AddCancel(cancel)

		// Get the config
		conf := viper.GetViper()
		node, err := node.NewPegnetd(ctx, conf)
		if err != nil {
			log.WithError(err).Errorf("failed to launch pegnet node")
			os.Exit(1)
		}

		apiserver := srv.NewAPIServer(conf, node)
		go apiserver.Start(ctx.Done())

		// Run
		node.DBlockSync(ctx)
	},
}

// always is run before any command
func always(cmd *cobra.Command, args []string) {
	// Setup config reading
	viper.SetConfigName("pegnetd-conf")
	// Add as many config paths as we want to check
	viper.AddConfigPath("$HOME/.pegnetd")
	viper.AddConfigPath(".")

	// Setup global command line flag overrides
	// This gets run before any command executes. It will init global flags to the config
	_ = viper.BindPFlag(config.LoggingLevel, cmd.Flags().Lookup("log"))
	_ = viper.BindPFlag(config.Server, cmd.Flags().Lookup("server"))
	_ = viper.BindPFlag(config.Wallet, cmd.Flags().Lookup("wallet"))
	_ = viper.BindPFlag(config.APIListen, cmd.Flags().Lookup("api"))

	// Also init some defaults
	viper.SetDefault(config.DBlockSyncRetryPeriod, time.Second*5)
	viper.SetDefault(config.SqliteDBPath, "$HOME/pegnetd/mainnet/sql.db")

	// Catch ctl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Info("Gracefully closing")
		exit.GlobalExitHandler.Close()

		log.Info("closing application")
		// If something is hanging, we have to kill it
		os.Exit(0)
	}()
}

// ReadConfig can be put as a PreRun for a command that uses the config file
func ReadConfig(cmd *cobra.Command, args []string) {
	err := viper.ReadInConfig()
	if err != nil {
		log.WithError(err).Error("failed to load config")
		os.Exit(1)
	}

	initLogger()
}

// TODO implement a dedicated logger
func initLogger() {
	switch strings.ToLower(viper.GetString(config.LoggingLevel)) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	}
}
