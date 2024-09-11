// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	//Progress bar will not show unless "--progress-bar" or "-p" flag is used.
	downloadCmd.Flags().BoolP("progress-bar", "p", false, "Display progress bar during download")
	downloadCmd.Flags().Lookup("progress-bar").NoOptDefVal = "true"

	downloadCmd.Flags().String("dir", ".", "Target directory where to store the files")
	downloadCmd.Flags().StringArray("tag", []string{}, "Only download the resources with the given tag")
	downloadCmd.Flags().StringArray("notag", []string{}, "Only download the resources without the given tag")
	downloadCmd.Flags().String("perm", "", "Optional permissions for the downloaded files (e.g. '644')")
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
	lock, err := internal.NewLock(lockFile, false)
	FatalIfNotNil(err)
	dir, err := cmd.Flags().GetString("dir")
	FatalIfNotNil(err)
	tags, err := cmd.Flags().GetStringArray("tag")
	FatalIfNotNil(err)
	notags, err := cmd.Flags().GetStringArray("notag")
	FatalIfNotNil(err)
	perm, err := cmd.Flags().GetString("perm")
	FatalIfNotNil(err)
	bar, err := cmd.Flags().GetBool("progress-bar")
	FatalIfNotNil(err)
	err = lock.Download(dir, tags, notags, perm, bar)
	FatalIfNotNil(err)
}
