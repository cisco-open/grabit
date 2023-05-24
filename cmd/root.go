// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grabit",
	Short: "Grabit downloads files from remote locations and verifies their integrity",
}

func getPwd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal().Msgf("error finding working directory %s", err.Error())
	}
	return path
}

var GRAB_LOCK = "grabit.lock"

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().StringP("lock-file", "f", filepath.Join(getPwd(), GRAB_LOCK), "lockfile path (default: $PWD/grabit.lock")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (trace, debug, info, warn, error, fatal)")
}

func initLog() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	ll, err := rootCmd.Flags().GetString("log-level")
	FatalIfNotNil(err)
	switch strings.ToLower(ll) {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "err", "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if strings.Contains(err.Error(), "unknown flag") {
			// exit code 126: Command invoked cannot execute
			os.Exit(126)
		}
		log.Fatal().Msg(err.Error())
	}
}

func FatalIfNotNil(err error) {
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}
