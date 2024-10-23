// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"fmt"

	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

// This code defines a command for a Cobra-based CLI application that displays the version of the application when the "version" command is executed.

func addVersion(cmd *cobra.Command) {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display grabit version",
		Run:   runVersion,
	}
	cmd.AddCommand(versionCmd)
}

// runVersion prints the version information of the Grabit application, including the version number, commit hash, and release date.
func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Grabit %s (commit: %s, date: %s)\n", internal.Version, internal.Commit, internal.Date)
}
