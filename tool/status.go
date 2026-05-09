package autonotes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "List unprocessed images",
	Run: func(cmd *cobra.Command, args []string) {
		images, err := getUnprocessedImages("src")
		if err != nil {
			fmt.Println("Error scanning src:", err)
			return
		}
		if len(images) == 0 {
			fmt.Println("No unprocessed images found!")
			return
		}
		fmt.Println("Unprocessed Images:")
		for _, img := range images {
			fmt.Println("-", img)
		}
	},
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
