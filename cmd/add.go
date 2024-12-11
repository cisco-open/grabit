// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package cmd

import (
	"fmt"
	"os"

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
	addCmd.Flags().String("artifactory-cache-url", "", "Artifactory cache URL")
	cmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	lockFile, err := cmd.Flags().GetString("lock-file")
	if err != nil {
		return err
	}
	// Get cache URL
	ArtifactoryCacheURL, err := cmd.Flags().GetString("artifactory-cache-url")
	if err != nil {
		return err
	}
	if ArtifactoryCacheURL != "" {
		token := os.Getenv(internal.GRABIT_ARTIFACTORY_TOKEN_ENV_VAR)
		if token == "" {
			return fmt.Errorf("%s environment variable is not set", internal.GRABIT_ARTIFACTORY_TOKEN_ENV_VAR)
		}
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
	err = lock.AddResource(args, algo, tags, filename, ArtifactoryCacheURL)
	if err != nil {
		return err
	}
	err = lock.Save()
	if err != nil {
		return err
	}
	return nil
}
