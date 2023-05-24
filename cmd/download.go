// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().String("dir", ".", "Target directory where to store the files")
	downloadCmd.Flags().StringArray("tag", []string{}, "Only download the resources with the given tag")
	downloadCmd.Flags().StringArray("notag", []string{}, "Only download the resources without the given tag")
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download defined resources",
	Args:  cobra.NoArgs,
	Run:   runFetch,
}

func runFetch(cmd *cobra.Command, args []string) {
	lockFile, err := cmd.Flags().GetString("lock-file")
	FatalIfNotNil(err)
	lock, err := internal.NewLock(lockFile, true)
	FatalIfNotNil(err)
	dir, err := cmd.Flags().GetString("dir")
	FatalIfNotNil(err)
	tags, err := cmd.Flags().GetStringArray("tag")
	FatalIfNotNil(err)
	notags, err := cmd.Flags().GetStringArray("notag")
	FatalIfNotNil(err)
	err = lock.Download(dir, tags, notags)
	FatalIfNotNil(err)
}
