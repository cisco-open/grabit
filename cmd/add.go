// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().String("algo", internal.RecommendedAlgo, "Integrity algorithm")
	addCmd.Flags().String("filename", "", "Target file name to use when downloading the resource")
	addCmd.Flags().StringArray("tag", []string{}, "Resource tags")
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new resource",
	Args:  cobra.MinimumNArgs(1),
	Run:   runAdd,
}

func runAdd(cmd *cobra.Command, args []string) {
	lockFile, err := cmd.Flags().GetString("lock-file")
	FatalIfNotNil(err)
	lock, err := internal.NewLock(lockFile, true)
	FatalIfNotNil(err)
	algo, err := cmd.Flags().GetString("algo")
	FatalIfNotNil(err)
	tags, err := cmd.Flags().GetStringArray("tag")
	FatalIfNotNil(err)
	filename, err := cmd.Flags().GetString("filename")
	FatalIfNotNil(err)
	err = lock.AddResource(args, algo, tags, filename)
	FatalIfNotNil(err)
	err = lock.Save()
	FatalIfNotNil(err)
}
