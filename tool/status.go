package autonotes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check project status (unprocessed images, broken references)",
	Run: func(cmd *cobra.Command, args []string) {
		images, err := getUnprocessedImages("src")
		if err != nil {
			fmt.Println("Error scanning images:", err)
		} else if len(images) > 0 {
			sort.Strings(images)
			fmt.Println("Unprocessed Images:")
			for _, img := range images {
				fmt.Println("-", img)
			}
			fmt.Println()
		} else {
			fmt.Println("No unprocessed images found!")
		}

		errors, warnings, err := validateNotes("src")
		if err != nil {
			fmt.Println("Error validating notes:", err)
		} else {
			if len(warnings) > 0 {
				sort.Strings(warnings)
				fmt.Println("Validation Warnings:")
				for _, msg := range warnings {
					fmt.Println("-", msg)
				}
				fmt.Println()
			}

			if len(errors) > 0 {
				sort.Strings(errors)
				fmt.Println("Validation Errors:")
				for _, msg := range errors {
					fmt.Println("-", msg)
				}
			} else {
				fmt.Println("All .note files are valid!")
			}
		}
	},
}

func validateNotes(srcDir string) ([]string, []string, error) {
	var allErrors []string
	var allWarnings []string

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".note") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			p := NewParser(string(content))
			ast, err := p.Parse()
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("%s: parse error: %v", path, err))
				return nil
			}

			// 1. Collect all declared UIDs
			declaredUIDs := make(map[string]bool)
			collector := &UIDCollector{uids: declaredUIDs}
			ast.Accept(collector)

			// 2. Check all references
			checker := &ReferenceChecker{path: path, declaredUIDs: declaredUIDs}
			ast.Accept(checker)
			allErrors = append(allErrors, checker.errors...)

			// 3. Check image existence
			collectionDir := filepath.Dir(path)
			imgValidator := &ImageValidator{path: path, collectionDir: collectionDir}
			ast.Accept(imgValidator)
			allErrors = append(allErrors, imgValidator.errors...)

			// 4. Check for math delimiters in reword blocks
			mathValidator := &MathValidator{path: path}
			ast.Accept(mathValidator)
			allWarnings = append(allWarnings, mathValidator.warnings...)
		}
		return nil
	})

	return allErrors, allWarnings, err
}

type UIDCollector struct {
	uids map[string]bool
}

func (c *UIDCollector) VisitText(n *TextNode) {}
func (c *UIDCollector) VisitBlock(n *BlockNode) {
	if uid := n.Attr("uid"); uid != "" {
		c.uids[uid] = true
	}
	for _, child := range n.Children {
		child.Accept(c)
	}
}

type ReferenceChecker struct {
	path         string
	declaredUIDs map[string]bool
	errors       []string
}

func (c *ReferenceChecker) VisitText(n *TextNode) {}
func (c *ReferenceChecker) VisitBlock(n *BlockNode) {
	if ref := n.Attr("ref"); ref != "" {
		refs := ExpandRefs(ref)
		for _, r := range refs {
			if !c.declaredUIDs[r] {
				msg := fmt.Sprintf("%s: broken reference '%s' in tag <%s>", c.path, r, n.Name)
				suggestions := FindSimilarUIDs(r, c.declaredUIDs)
				if len(suggestions) > 0 {
					msg += fmt.Sprintf(" (did you mean: %s?)", strings.Join(suggestions, ", "))
				}
				c.errors = append(c.errors, msg)
			}
		}
	}
	for _, child := range n.Children {
		child.Accept(c)
	}
}

type ImageValidator struct {
	path          string
	collectionDir string
	errors        []string
}

func (c *ImageValidator) VisitText(n *TextNode) {}
func (c *ImageValidator) VisitBlock(n *BlockNode) {
	checkImage := func(imgAttr string) {
		if imgAttr == "" {
			return
		}
		imgName := strings.TrimPrefix(imgAttr, "images/")
		fullPath := filepath.Join(c.collectionDir, "images", imgName)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.errors = append(c.errors, fmt.Sprintf("%s: image file '%s' not found", c.path, imgAttr))
		}
	}

	checkImage(n.Attr("image"))
	if n.Name == "image" {
		checkImage(n.Attr("src"))
	}

	for _, child := range n.Children {
		child.Accept(c)
	}
}

type MathValidator struct {
	path     string
	warnings []string
}

func (v *MathValidator) VisitText(n *TextNode) {}
func (v *MathValidator) VisitBlock(n *BlockNode) {
	if n.Name == "reword" {
		var sb strings.Builder
		v.collectTextIgnoringMath(n, &sb)
		text := sb.String()

		if idx := strings.Index(text, "$"); idx != -1 {
			v.warnings = append(v.warnings, fmt.Sprintf("%s: found '$' in <reword> block (%s), do not use this prefer <math>", v.path, v.extractSnippet(text, idx)))
		}

		latexDelims := []string{`\(`, `\)`, `\[`, `\]`}
		for _, d := range latexDelims {
			if idx := strings.Index(text, d); idx != -1 {
				v.warnings = append(v.warnings, fmt.Sprintf("%s: found LaTeX-style delimiter '%s' in <reword> block (%s), do not use this prefer <math>", v.path, d, v.extractSnippet(text, idx)))
				break
			}
		}
	}
	for _, child := range n.Children {
		child.Accept(v)
	}
}

func (v *MathValidator) collectTextIgnoringMath(n Node, sb *strings.Builder) {
	if t, ok := n.(*TextNode); ok {
		sb.WriteString(t.Content)
	} else if b, ok := n.(*BlockNode); ok {
		if b.Name == "math" {
			return
		}
		for _, child := range b.Children {
			v.collectTextIgnoringMath(child, sb)
		}
	}
}

func (v *MathValidator) extractSnippet(text string, index int) string {
	start := max(index-20, 0)
	end := index + 20
	if end > len(text) {
		end = len(text)
	}
	snippet := text[start:end]
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	return "..." + strings.TrimSpace(snippet) + "..."
}

func getUnprocessedImages(srcDir string) ([]string, error) {
	var unprocessed []string

	// Find all images
	allImages := make(map[string]bool)
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				// We keep track by relative path from the project root or the collection root?
				// "src/ist-geom/images/photo1.jpg"
				// Let's store the full relative path
				allImages[path] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Find all .note and extract linked images
	usedImages := make(map[string]bool)
	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".note") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			p := NewParser(string(content))
			ast, err := p.Parse()
			if err != nil {
				fmt.Printf("Warning: couldn't parse %s: %v\n", path, err)
				return nil
			}

			// We need to resolve the image attribute relative to the collection.
			// Path is like: src/ist-geom/lesson-01.note
			// collection dir is src/ist-geom
			collectionDir := filepath.Dir(path)
			finder := &ImageFinder{collectionDir: collectionDir, used: usedImages}
			ast.Accept(finder)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for img := range allImages {
		if !usedImages[img] {
			unprocessed = append(unprocessed, img)
		}
	}

	return unprocessed, nil
}

type ImageFinder struct {
	collectionDir string
	used          map[string]bool
}

func (f *ImageFinder) VisitText(n *TextNode) {}

func (f *ImageFinder) VisitBlock(n *BlockNode) {
	if img := n.Attr("image"); img != "" {
		imgName := strings.TrimPrefix(img, "images/")
		fullPath := filepath.Join(f.collectionDir, "images", imgName)
		f.used[fullPath] = true
	}
	// Also check "src" for "image" tag
	if n.Name == "image" {
		if src := n.Attr("src"); src != "" {
			imgName := strings.TrimPrefix(src, "images/")
			fullPath := filepath.Join(f.collectionDir, "images", imgName)
			f.used[fullPath] = true
		}
	}

	for _, child := range n.Children {
		child.Accept(f)
	}
}
