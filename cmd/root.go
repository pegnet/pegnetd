package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/exit"
	"github.com/pegnet/pegnetd/node"
	"github.com/pegnet/pegnetd/srv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().String("log", "info", "Change the logging level. Can choose from 'trace', 'debug', 'info', 'warn', 'error', or 'fatal'")
	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:8088/v2", "The url to the factomd endpoint without a trailing slash")
	rootCmd.PersistentFlags().StringP("wallet", "w", "http://localhost:8089/v2", "The url to the factomd-wallet endpoint without a trailing slash")
	rootCmd.PersistentFlags().String("walletuser", "", "The username for Wallet RPC")
	rootCmd.PersistentFlags().String("walletpassword", "", "The password for Wallet RPC")
	rootCmd.PersistentFlags().StringP("pegnetd", "p", "http://localhost:8070", "The url to the pegnetd endpoint without a trailing slash")
	rootCmd.PersistentFlags().String("api", "8070", "Change the api listening port for the api")
	rootCmd.PersistentFlags().String("config", "", "Optional file location of the config file")

	rootCmd.Flags().String("dbmode", "", "Turn on custom sqlite modes")
	rootCmd.Flags().Bool("wal", false, "Turn on WAL mode for sqlite")

	// This is for testing purposes
	rootCmd.PersistentFlags().Bool("testing", false, "If this flag is set, all activations heights are set to 0.")
	rootCmd.PersistentFlags().Int("act", -1, "Able to manually set the activation heights")
}

// Execute is cobra's entry point
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//fmt.Println(err)
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
	// See if we are in testing mode
	if ok, _ := cmd.Flags().GetBool("testing"); ok {
		log.Infof("in testing mode, activation heights are 0")
		act, _ := cmd.Flags().GetInt("act")
		if act <= 0 {
			act = 0
		}

		// Set all activations for testing
		node.SetAllActivations(uint32(act))
	}

	// Setup config reading
	if cFilePath, _ := cmd.Flags().GetString("config"); cFilePath != "" {
		base := filepath.Base(cFilePath)
		dir := filepath.Dir(cFilePath)
		viper.SetConfigFile(base)
		viper.AddConfigPath(dir)
	} else {
		viper.SetConfigName("pegnetd-conf")
		// Add as many config paths as we want to check
		viper.AddConfigPath("$HOME/.pegnetd")
		viper.AddConfigPath(".")
	}

	// Setup global command line flag overrides
	// This gets run before any command executes. It will init global flags to the config
	_ = viper.BindPFlag(config.LoggingLevel, cmd.Flags().Lookup("log"))
	_ = viper.BindPFlag(config.Server, cmd.Flags().Lookup("server"))
	_ = viper.BindPFlag(config.Wallet, cmd.Flags().Lookup("wallet"))
	_ = viper.BindPFlag(config.WalletUser, cmd.Flags().Lookup("walletuser"))
	_ = viper.BindPFlag(config.WalletPass, cmd.Flags().Lookup("walletpassword"))
	_ = viper.BindPFlag(config.Pegnetd, cmd.Flags().Lookup("pegnetd"))
	_ = viper.BindPFlag(config.APIListen, cmd.Flags().Lookup("api"))
	_ = viper.BindPFlag(config.SQLDBWalMode, cmd.Flags().Lookup("wal"))
	_ = viper.BindPFlag(config.CustomSQLDBMode, cmd.Flags().Lookup("dbmode"))

	// Also init some defaults
	viper.SetDefault(config.DBlockSyncRetryPeriod, time.Second*5)
	viper.SetDefault(config.SqliteDBPath, "$HOME/.pegnetd/mainnet/sql.db")

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

	// If no config is found, we will attempt to make one
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// No config found? We will write the default config for the user
		// If the custom config path is set, then we should not write a new config.
		if custom, _ := cmd.Flags().GetString("config"); custom == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				log.WithError(err).Fatal("failed to create config path")
			}

			// Create the pegnetd directory if it is not already
			err = os.MkdirAll(filepath.Join(home, ".pegnetd"), 0777)
			if err != nil {
				log.WithError(err).Fatal("failed to create config path")
			}

			configpath := filepath.Join(home, ".pegnetd", "pegnetd-conf.toml")
			_, err = os.Stat(configpath)
			if os.IsExist(err) { // Double check a file does not already exist. Don't overwrite a config
				log.WithField("path", configpath).Fatal("config exists, but unable to read")
			}

			// Attempt to write a new config file
			err = viper.WriteConfigAs(configpath)
			if err != nil {
				log.WithField("path", configpath).WithError(err).Fatal("failed to create config")
			}
			// Inform the user we made a config
			log.WithField("path", configpath).Infof("no config file, one was created")

			// Try to read it again
			err = viper.ReadInConfig()
			if err != nil {
				log.WithError(err).Fatal("failed to load config")
			}
		}
	} else if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	// Indicate which config was used
	log.Infof("Using config from %s", viper.ConfigFileUsed())

	initLogger()
}

// SoftReadConfig will not fail. It can be used for a command that needs the config,
// but is happy with the defaults
func SoftReadConfig(cmd *cobra.Command, args []string) {
	err := viper.ReadInConfig()
	if err != nil {
		log.WithError(err).Debugf("failed to load config")
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
