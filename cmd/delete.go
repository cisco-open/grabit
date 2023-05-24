// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(delCmd)
}

var delCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete existing resources",
	Args:  cobra.MinimumNArgs(1),
	Run:   runDel,
}

func runDel(cmd *cobra.Command, args []string) {
	lockFile, err := cmd.Flags().GetString("lock-file")
	FatalIfNotNil(err)
	lock, err := internal.NewLock(lockFile, false)
	FatalIfNotNil(err)
	for _, r := range args {
		lock.DeleteResource(r)
	}
	err = lock.Save()
	FatalIfNotNil(err)
}
