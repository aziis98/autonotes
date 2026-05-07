package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run status and then build",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--- Running Status ---")
		statusCmd.Run(cmd, args)
		fmt.Println("\n--- Running Build ---")
		buildCmd.Run(cmd, args)
		fmt.Println("\nSync completed successfully!")
	},
}
