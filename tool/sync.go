package autonotes

import (
	"fmt"

	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run status and then build",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--- Running Status ---")
		StatusCmd.Run(cmd, args)
		fmt.Println("\n--- Running Build ---")
		BuildCmd.Run(cmd, args)
		fmt.Println("\nSync completed successfully!")
	},
}
