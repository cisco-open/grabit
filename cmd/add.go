// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

// This code defines a command for adding a new resource using the Cobra library.
// It sets up the "add" command with a short description, requires at least one argument,
// and specifies flags for the integrity algorithm, target filename, and resource tags.

func addAdd(cmd *cobra.Command) {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add new resource",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runAdd,
	}
	addCmd.Flags().String("algo", internal.RecommendedAlgo, "Integrity algorithm")
	addCmd.Flags().String("filename", "", "Target file name to use when downloading the resource")
	addCmd.Flags().StringArray("tag", []string{}, "Resource tags")
	addCmd.Flags().Bool("dynamic", false, "Mark resource as dynamic") // Add this line
	cmd.AddCommand(addCmd)
}

// This function, runAdd, handles the addition of a resource by first retrieving various command line flags such as lock file, algorithm, tags, filename, and dynamic status. It then creates a new lock using the specified lock file, adds the resource using the provided arguments, and finally saves the lock. Error handling is implemented to ensure that any issues during these operations are returned appropriately.

func runAdd(cmd *cobra.Command, args []string) error {
	lockFile, err := cmd.Flags().GetString("lock-file")
	if err != nil {
		return err
	}
	lock, err := internal.NewLock(lockFile, true)
	if err != nil {
		return err
	}
	algo, err := cmd.Flags().GetString("algo")
	if err != nil {
		return err
	}
	tags, err := cmd.Flags().GetStringArray("tag")
	if err != nil {
		return err
	}
	filename, err := cmd.Flags().GetString("filename")
	if err != nil {
		return err
	}
	dynamic, err := cmd.Flags().GetBool("dynamic")
	if err != nil {
		return err
	}

	err = lock.AddResource(args, algo, tags, filename, dynamic)
	if err != nil {
		return err
	}

	return lock.Save()
}
