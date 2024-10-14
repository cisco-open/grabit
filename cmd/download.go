// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"github.com/cisco-open/grabit/downloader"
	"github.com/cisco-open/grabit/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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

func runFetch(cmd *cobra.Command, args []string) error {
	logLevel, _ := cmd.Flags().GetString("log-level")
	level, _ := zerolog.ParseLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	if level <= zerolog.DebugLevel {
		log.Debug().Msg("Starting download")
		// Add more debug logs as needed
	}

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

	d := cmd.Context().Value("downloader").(*downloader.Downloader)

	if verbose {
		log.Debug().Str("lockFile", lockFile).Str("dir", dir).Strs("tags", tags).Strs("notags", notags).Str("perm", perm).Msg("Starting download")
	}

	err = lock.Download(dir, tags, notags, perm, d)
	if err != nil {
		return err
	}

	if verbose {
		log.Debug().Msg("Download completed successfully")
	}

	return nil
}
