// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

// This code defines a command for deleting existing resources using the Cobra library in Go.
// The 'delete' command requires at least one argument and executes the 'runDel' function when invoked.

func addDelete(cmd *cobra.Command) {
	var delCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete existing resources",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runDel,
	}
	cmd.AddCommand(delCmd)
}

// runDel function handles the deletion of resources specified in the command arguments.
// It retrieves a lock file path from the command flags, creates a lock instance,
// deletes each resource listed in the arguments, and saves the updated lock state.

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
		lock.DeleteResource(r)
	}
	err = lock.Save()
	if err != nil {
		return err
	}
	return nil
}
