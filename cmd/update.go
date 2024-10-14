package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [URL]",
	Short: "Update a resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lockFile, _ := cmd.Flags().GetString("lock-file")
		lock, err := internal.NewLock(lockFile, false)
		if err != nil {
			return err
		}
		return lock.UpdateResource(args[0])
	},
}

func AddUpdate(cmd *cobra.Command) {
	cmd.AddCommand(updateCmd)
}
