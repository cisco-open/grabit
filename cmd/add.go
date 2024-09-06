// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

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
	cmd.AddCommand(addCmd)
}

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
	err = lock.AddResource(args, algo, tags, filename)
	if err != nil {
		return err
	}
	err = lock.Save()
	if err != nil {
		return err
	}
	return nil
}
