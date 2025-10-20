package main

import (
	"errors"
	"log/slog"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l1"
	"github.com/compose-network/local-testnet/internal/l2"
	"github.com/compose-network/local-testnet/internal/logger"
	"github.com/compose-network/local-testnet/internal/observability"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const appName = "localnet-control-plane"

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "CLI for managing local L1 and L2 network",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger.Initialize(slog.LevelDebug)

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")

		if err := viper.ReadInConfig(); err != nil {
			const errMsg = "error reading config file"
			slog.With("err", err.Error()).Error(errMsg)
			return errors.Join(err, errors.New(errMsg))
		}
		if err := viper.Unmarshal(&configs.Values); err != nil {
			const errMsg = "unable to decode application config"
			slog.With("err", err.Error()).Error(errMsg)
			return errors.Join(err, errors.New(errMsg))
		}

		slog.
			With("config_file", viper.ConfigFileUsed()).
			With("config", configs.Values).
			Debug("configurations loaded")

		return nil
	},
}

func main() {
	rootCmd.Short = appName

	rootCmd.AddCommand(l1.CMD)
	rootCmd.AddCommand(l2.CMD)
	rootCmd.AddCommand(observability.CMD)

	if err := rootCmd.Execute(); err != nil {
		slog.With("err", err.Error()).Error("failed to execute root command")
		panic(err.Error())
	}
}
