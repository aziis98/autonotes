package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	selectFilter  string
	grepFilter    string
	extractFilter string
)

var queryCmd = &cobra.Command{
	Use:   "query [path]",
	Short: "Query information from the notes",
	Run: func(cmd *cobra.Command, args []string) {
		var files []string
		if len(args) > 0 {
			files = append(files, args[0])
		} else {
			filepath.WalkDir("src", func(path string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() && strings.HasSuffix(path, ".note") {
					files = append(files, path)
				}
				return nil
			})
		}

		var selectedTypes []string
		if selectFilter != "" && selectFilter != "all" {
			selectedTypes = strings.Split(selectFilter, ",")
		}

		var extractTypes []string
		if extractFilter != "" {
			extractTypes = strings.Split(extractFilter, ",")
		}

		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			p := NewParser(string(content))
			ast, err := p.Parse()
			if err != nil {
				continue
			}

			// Find the lesson node
			var lessonNode *Node
			for _, child := range ast.Children {
				if child.Name == "lesson" {
					lessonNode = child
					break
				}
			}

			if lessonNode == nil {
				continue
			}

			for _, block := range lessonNode.Children {
				if block.Type != "element" || block.Name == "summary" {
					continue
				}

				// Filter by type
				if selectedTypes != nil {
					found := false
					for _, t := range selectedTypes {
						if block.Name == t {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				// Filter by grep
				if grepFilter != "" {
					text := getAllText(block)
					if !strings.Contains(strings.ToLower(text), strings.ToLower(grepFilter)) {
						continue
					}
				}

				// Extract and Print
				fmt.Printf("[%s] %s\n", file, block.Name)
				if extractTypes == nil {
					printNode(block, 1)
				} else {
					for _, child := range block.Children {
						for _, eType := range extractTypes {
							if child.Name == eType {
								printNode(child, 1)
							}
						}
					}
				}
				fmt.Println()
			}
		}
	},
}

func getAllText(n *Node) string {
	if n.Type == "text" {
		return n.Content
	}
	var res strings.Builder
	for _, child := range n.Children {
		res.WriteString(getAllText(child))
		res.WriteString(" ")
	}
	return res.String()
}

func printNode(n *Node, indent int) {
	if n.Type == "text" {
		content := strings.TrimSpace(n.Content)
		if content == "" {
			return
		}
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			fmt.Printf("%s%s\n", strings.Repeat("  ", indent), strings.TrimSpace(line))
		}
		return
	}

	var attrs []string
	for k, v := range n.Attributes {
		attrs = append(attrs, fmt.Sprintf("%s=%q", k, v))
	}
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	if len(n.Children) == 0 {
		fmt.Printf("%s<%s%s />\n", strings.Repeat("  ", indent), n.Name, attrStr)
		return
	}

	fmt.Printf("%s<%s%s>\n", strings.Repeat("  ", indent), n.Name, attrStr)
	for _, child := range n.Children {
		printNode(child, indent+1)
	}
	fmt.Printf("%s</%s>\n", strings.Repeat("  ", indent), n.Name)
}

var querySummaryCmd = &cobra.Command{
	Use:   "summary <path>",
	Short: "Extract summary from a .note file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		notePath := args[0]
		// We don't really care about the outPath here for debugging extraction
		summary, err := processNoteFile(notePath, "/tmp/debug.html")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Summary for %s:\n", notePath)
		fmt.Printf("------------------\n")
		fmt.Println(summary)
		fmt.Printf("------------------\n")
	},
}

func init() {
	queryCmd.Flags().StringVarP(&selectFilter, "select", "s", "all", "Filter blocks by type")
	queryCmd.Flags().StringVarP(&grepFilter, "grep", "g", "", "Search text in blocks")
	queryCmd.Flags().StringVarP(&extractFilter, "extract", "e", "", "Extract specific child blocks")

	queryCmd.AddCommand(querySummaryCmd)
}
