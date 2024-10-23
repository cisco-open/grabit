// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// This Go code defines a command-line interface (CLI) command for downloading resources.
// It uses the Cobra library to create a command called "download" with various flags
// to customize the download process, such as specifying the target directory,
// filtering by tags, setting file permissions, and enabling verbose output.

func addDownload(cmd *cobra.Command) {
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download defined resources",
		Args:  cobra.NoArgs,
		RunE:  runFetch,
	}
	downloadCmd.Flags().String("dir", ".", "Target directory where to store the files")
	downloadCmd.Flags().StringArray("tag", []string{}, "Only download the resources with the given tag")
	downloadCmd.Flags().StringArray("notag", []string{}, "Only download the resources without the given tag")
	downloadCmd.Flags().String("perm", "", "Optional permissions for the downloaded files (e.g. '644')")
	downloadCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.AddCommand(downloadCmd)
}

// Summary: The runFetch function handles the download process by configuring logging levels,
// managing a lock file, and retrieving various command-line arguments such as directory,
// tags, and permissions. It sets up a context for the downloader and initiates the download
// process while logging relevant information at different stages.

func runFetch(cmd *cobra.Command, args []string) error {
	logLevel, _ := cmd.Flags().GetString("log-level")
	level, _ := zerolog.ParseLevel(logLevel)
	zerolog.SetGlobalLevel(level)

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

	return lock.Download(dir, tags, notags, perm)
}
