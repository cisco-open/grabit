// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

func addDownload(cmd *cobra.Command) {
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download defined resources",
		Args:  cobra.NoArgs,
		RunE:  runFetch,
	}
	downloadCmd.Flags().BoolP("status", "s", false, "Continuously display bytes/resources downloaded, time elapsed, and progress bar")
	downloadCmd.Flags().Lookup("status").NoOptDefVal = "true"
	downloadCmd.Flags().String("dir", ".", "Target directory where to store the files")
	downloadCmd.Flags().StringArray("tag", []string{}, "Only download the resources with the given tag")
	downloadCmd.Flags().StringArray("notag", []string{}, "Only download the resources without the given tag")
	downloadCmd.Flags().String("perm", "", "Optional permissions for the downloaded files (e.g. '644')")
	cmd.AddCommand(downloadCmd)
}

func runFetch(cmd *cobra.Command, args []string) error {
	lockFile, err := cmd.Flags().GetString("lock-file")
	if err != nil {
		return err
	}
	lock, err := internal.NewLock(lockFile, false)
	if err != nil {
		return err
	}
	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}
	tags, err := cmd.Flags().GetStringArray("tag")
	if err != nil {
		return err
	}
	notags, err := cmd.Flags().GetStringArray("notag")
	if err != nil {
		return err
	}
	perm, err := cmd.Flags().GetString("perm")
	if err != nil {
		return err
	}
	status, err := cmd.Flags().GetBool("status")
	if err != nil {
		return err
	}
	err = lock.Download(dir, tags, notags, perm, status)
	if err != nil {
		return err
	}
	return nil
}
