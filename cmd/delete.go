// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func addDelete(cmd *cobra.Command) {
	var delCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete existing resources",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runDel,
	}
	cmd.AddCommand(delCmd)
}

func runDel(cmd *cobra.Command, args []string) error {
	lockFile, err := cmd.Flags().GetString("lock-file")
	if err != nil {
		return err
	}
	lock, err := internal.NewLock(lockFile, false)
	if err != nil {
		return err
	}
	for _, r := range args {
		err = lock.DeleteResource(r)
		if err != nil {
			return err
		}
	}
	err = lock.Save()
	if err != nil {
		return err
	}
	return nil
}
