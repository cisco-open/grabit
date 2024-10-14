package cmd

import (
	"github.com/cisco-open/grabit/internal"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the integrity of downloaded resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		lockFile, _ := cmd.Flags().GetString("lock-file")
		lock, err := internal.NewLock(lockFile, false)
		if err != nil {
			return err
		}
		return lock.VerifyIntegrity()
	},
}

func AddVerify(cmd *cobra.Command) {
	cmd.AddCommand(verifyCmd)
}
