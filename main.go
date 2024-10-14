// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package main

import (
	"os"
	"os/signal"

	"github.com/cisco-open/grabit/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Log to stdout.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Exit immediately upon reception of an interrupt signal.
	stop4Chan := make(chan os.Signal, 1)
	signal.Notify(stop4Chan, os.Interrupt)
	go listenForInterrupt(stop4Chan)

	rootCmd := cmd.NewRootCmd()
	cmd.Execute(rootCmd)
}

func listenForInterrupt(stopScan chan os.Signal) {
	<-stopScan
	log.Fatal().Msg("Interrupt signal received. Exiting.")
}
