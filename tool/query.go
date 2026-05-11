package autonotes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	selectFilter  string
	grepFilter    string
	extractFilter string
	histogramMode bool
	verboseMode   bool
)

var QueryCmd = &cobra.Command{
	Use:   "query [path]",
	Short: "Query information from the notes",
	Example: `  # Search for all theorems
  autonotes query --select theorem

  # Search for a specific term
  autonotes query --grep "Frobenius"

  # Extract only the reworded text for all definitions
  autonotes query --select definition --extract reword

  # Show counts for all found tag types
  autonotes query --histogram

  # Verbose output for debugging
  autonotes query -v --grep "topologia"`,
	Run: func(cmd *cobra.Command, args []string) {
		var files []string
		if len(args) > 0 {
			files = append(files, args[0])
		} else {
			if verboseMode {
				fmt.Fprintln(os.Stderr, "Globbing src/ for .note files...")
			}
			filepath.WalkDir("src", func(path string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() && strings.HasSuffix(path, ".note") {
					files = append(files, path)
				}
				return nil
			})
		}

		if verboseMode {
			fmt.Fprintf(os.Stderr, "Found %d files.\n", len(files))
		}

		histogram := make(map[string]int)

		var selectedTypes []string
		if selectFilter != "" && selectFilter != "all" {
			selectedTypes = strings.Split(selectFilter, ",")
		}

		var extractTypes []string
		if extractFilter != "" {
			extractTypes = strings.Split(extractFilter, ",")
		}

		foundAny := false
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			if verboseMode {
				fmt.Fprintf(os.Stderr, "Parsing %s...\n", file)
			}
			p := NewParser(string(content))
			ast, err := p.Parse()
			if err != nil {
				if verboseMode {
					fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file, err)
				}
				continue
			}

			// Find the lesson node
			root, ok := ast.(*BlockNode)
			if !ok {
				continue
			}

			lessonNode := root.FindChild("lesson")

			if lessonNode == nil {
				continue
			}

			for _, child := range lessonNode.Children {
				block, ok := child.(*BlockNode)
				if !ok || block.Name == "summary" {
					continue
				}

				// Filter by type
				if selectedTypes != nil {
					if !slices.Contains(selectedTypes, block.Name) {
						continue
					}
				}

				// Filter by grep
				if grepFilter != "" {
					extractor := &TextExtractor{}
					block.Accept(extractor)
					text := extractor.String()
					if !strings.Contains(strings.ToLower(text), strings.ToLower(grepFilter)) {
						continue
					}
				}

				if histogramMode {
					histogram[block.Name]++
					foundAny = true
					continue
				}

				foundAny = true
				// Extract and Print
				fmt.Printf("[%s] %s\n", file, block.Name)
				printer := NewPrinter(os.Stdout)
				if extractTypes == nil {
					printer.Print(block)
				} else {
					for _, child := range block.Children {
						if b, ok := child.(*BlockNode); ok {
							if slices.Contains(extractTypes, b.Name) {
								printer.Print(b)
							}
						}
					}
				}
				fmt.Println()
			}
		}

		if histogramMode {
			type entry struct {
				tag   string
				count int
			}
			var entries []entry
			for tag, count := range histogram {
				entries = append(entries, entry{tag, count})
			}
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].count == entries[j].count {
					return entries[i].tag < entries[j].tag
				}
				return entries[i].count > entries[j].count
			})

			fmt.Println("Tag Histogram:")
			for _, e := range entries {
				fmt.Printf("  %s: %d\n", e.tag, e.count)
			}
		}

		if !foundAny {
			os.Exit(1)
		}
	},
}

type TextExtractor struct {
	sb strings.Builder
}

func (e *TextExtractor) VisitText(n *TextNode) {
	e.sb.WriteString(n.Content)
	e.sb.WriteString(" ")
}

func (e *TextExtractor) VisitBlock(n *BlockNode) {
	for _, child := range n.Children {
		child.Accept(e)
	}
}

func (e *TextExtractor) String() string {
	return e.sb.String()
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
	QueryCmd.Flags().StringVarP(&selectFilter, "select", "s", "all", "Filter blocks by type")
	QueryCmd.Flags().StringVarP(&grepFilter, "grep", "g", "", "Search text in blocks")
	QueryCmd.Flags().StringVarP(&extractFilter, "extract", "e", "", "Extract specific child blocks")
	QueryCmd.Flags().BoolVar(&histogramMode, "histogram", false, "Print counts for all found tag types")
	QueryCmd.Flags().BoolVarP(&verboseMode, "verbose", "v", false, "Print globbing and parsing details to stderr")

	QueryCmd.AddCommand(querySummaryCmd)
}
