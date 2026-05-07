package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "converter",
	Short: "AutoNotes converter tool",
	Long:  `A tool to check unprocessed images and build HTML notes from .note files.`,
}

var debugMode bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug mode")
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(queryCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
