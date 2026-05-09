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
			findImagesInAST(ast, collectionDir, usedImages)
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

func findImagesInAST(node *Node, collectionDir string, used map[string]bool) {
	if node == nil {
		return
	}
	if img, ok := node.Attributes["image"]; ok && img != "" {
		// Image paths are specified as just "photo1.jpg" and imply "./images/photo1.jpg"
		// The full path would be collectionDir + "/images/" + img
		// Or maybe the user literally wrote "image=images/photo1.jpg"?
		// "lesson-01.note can write image=photo1.jpg to point to ./images/photo1.jpg and the 'images/' part is always implied and can also be omitted"
		imgName := img
		imgName = strings.TrimPrefix(imgName, "images/")
		fullPath := filepath.Join(collectionDir, "images", imgName)
		used[fullPath] = true
	}
	
	for _, child := range node.Children {
		findImagesInAST(child, collectionDir, used)
	}
}
