// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cisco-open/grabit/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns a new Cobra command for the Grabit application.
// It sets up the command's usage, description, and persistent flags for configuration options
// such as lock file path, log level, and verbosity. Additional functionalities like delete,
// download, add, version, update, and verify commands are also registered.

var verbose bool

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "grabit",
		Short:        "Grabit downloads files from remote locations and verifies their integrity",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringP("lock-file", "f", filepath.Join(getPwd(), GRAB_LOCK), "lockfile path (default: $PWD/grabit.lock")
	cmd.PersistentFlags().StringP("log-level", "l", "info", "log level (trace, debug, info, warn, error, fatal)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	addDelete(cmd)
	addDownload(cmd)
	addAdd(cmd)
	addVersion(cmd)
	return cmd
}

// getPwd retrieves the current working directory of the program. If an error occurs while fetching the directory, it logs the error and terminates the program.

func getPwd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal().Msgf("error finding working directory %s", err.Error())
	}
	return path
}

// This Go code initializes a global logging level using the zerolog package based on the provided log level string.
// It supports various levels: trace, debug, info, warn, error, and fatal.
// If an unrecognized level is provided, it defaults to info level.
// Additionally, a variable 'd' is defined to hold a function reference to get a URL to a temporary file.

var GRAB_LOCK = "grabit.lock"

func initLog(ll string) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
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

var d = internal.GetUrltoTempFile

// Execute initializes the command-line application by setting up a persistent pre-run function
// that adds a context value for a downloader, retrieves the log level from flags, initializes logging,
// and executes the root command. It handles errors, specifically checking for unknown flags
// and exiting with a specific code if encountered.

func Execute(rootCmd *cobra.Command) {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ctx := context.WithValue(cmd.Context(), "downloader", d)
		cmd.SetContext(ctx)
	}

	ll, err := rootCmd.PersistentFlags().GetString("log-level")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	initLog(ll)
	if err := rootCmd.Execute(); err != nil {
		if strings.Contains(err.Error(), "unknown flag") {
			// exit code 126: Command invoked cannot execute
			os.Exit(126)
		}
		log.Fatal().Msg(err.Error())
	}
}
