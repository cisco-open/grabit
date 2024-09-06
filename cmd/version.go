// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"fmt"

	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func addVersion(cmd *cobra.Command) {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display grabit version",
		Run:   runVersion,
	}
	cmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Grabit %s (commit: %s, date: %s)\n", internal.Version, internal.Commit, internal.Date)
}
