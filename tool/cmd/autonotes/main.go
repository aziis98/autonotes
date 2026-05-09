package main

import (
	"fmt"
	"os"

	"autonotes/tool"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "converter",
	Short: "AutoNotes converter tool",
	Long:  `A tool to check unprocessed images and build HTML notes from .note files.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&autonotes.DebugMode, "debug", "d", false, "Enable debug mode")
	rootCmd.AddCommand(autonotes.StatusCmd)
	rootCmd.AddCommand(autonotes.BuildCmd)
	rootCmd.AddCommand(autonotes.SyncCmd)
	rootCmd.AddCommand(autonotes.CheckCmd)
	rootCmd.AddCommand(autonotes.ServeCmd)
	rootCmd.AddCommand(autonotes.QueryCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
