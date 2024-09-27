// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package main

import (
	"github.com/cisco-open/grabit/cmd"
	"github.com/cisco-open/grabit/downloader"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Log to stdout.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Exit immediately upon reception of an interrupt signal.
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	go listenForInterrupt(stopChan)

	d := downloader.NewDownloader(30 * time.Second)
	rootCmd := cmd.NewRootCmd(d)
	cmd.Execute(rootCmd)
}

func listenForInterrupt(stopScan chan os.Signal) {
	<-stopScan
	log.Fatal().Msg("Interrupt signal received. Exiting.")
}
