package autonotes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check .note files for syntax errors",
	Run: func(cmd *cobra.Command, args []string) {
		errorsFound := 0
		filesChecked := 0

		err := filepath.WalkDir("src", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(path, ".note") {
				filesChecked++
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("❌ %s: Could not read file: %v\n", path, err)
					errorsFound++
					return nil
				}

				p := NewParser(string(content))
				_, err = p.Parse()
				if err != nil {
					fmt.Printf("❌ %s: %v\n", path, err)
					errorsFound++
				} else {
					fmt.Printf("✅ %s: OK\n", path)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking src directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSummary: Checked %d files, found %d errors.\n", filesChecked, errorsFound)
		if errorsFound > 0 {
			os.Exit(1)
		}
	},
}
