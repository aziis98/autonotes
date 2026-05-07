package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
	"golang.org/x/image/draw"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile .note files to HTML",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		// Create out/ directory if it doesn't exist
		os.MkdirAll("out", 0755)

		builtCount := 0
		copiedCount := 0
		generatedCount := 0
		summaries := make(map[string]template.HTML)

		// Copy static assets from tpl/ to out/
		staticAssets := []string{"styles.css", "app.js"}
		for _, asset := range staticAssets {
			content, err := os.ReadFile(filepath.Join("tpl", asset))
			if err == nil {
				os.WriteFile(filepath.Join("out", asset), content, 0644)
				copiedCount++
			}
		}

		err := filepath.WalkDir("src", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			relPath, _ := filepath.Rel("src", path)
			if d.IsDir() {
				os.MkdirAll(filepath.Join("out", relPath), 0755)
				return nil
			}

			outPath := filepath.Join("out", relPath)
			if strings.HasSuffix(path, ".note") {
				outPath = strings.TrimSuffix(outPath, ".note") + ".html"
				summary, err := processNoteFile(path, outPath)
				if err != nil {
					fmt.Printf("Error processing %s: %v\n", path, err)
				} else {
					builtCount++
					summaries[outPath] = summary
				}
			} else {
				ext := strings.ToLower(filepath.Ext(path))
				isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
				
				if isImage {
					imageOutPath := strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".jpg"
					
					// Skip if out file is more recent than original image
					srcInfo, err := d.Info()
					if err == nil {
						dstInfo, err := os.Stat(imageOutPath)
						if err == nil && dstInfo.ModTime().After(srcInfo.ModTime()) {
							copiedCount++
							return nil
						}
					}

					fmt.Printf("Processing image %s...\n", path)
					if err := processImage(path, imageOutPath); err != nil {
						fmt.Printf("Error processing image %s: %v\n", path, err)
						// Fallback to simple copy with original extension
						content, _ := os.ReadFile(path)
						os.WriteFile(outPath, content, 0644)
					} else {
						copiedCount++
					}
				} else {
					content, err := os.ReadFile(path)
					if err == nil {
						os.WriteFile(outPath, content, 0644)
						copiedCount++
					}
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println("Error:", err)
		}

		// Generate index.html for each directory using the new template
		filepath.WalkDir("out", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				entries, _ := os.ReadDir(path)
				type Entry struct {
					Name    string
					Link    string
					IsDir   bool
					Summary template.HTML
				}
				var dashboardEntries []Entry

				rel, _ := filepath.Rel("out", path)

				for _, e := range entries {
					if e.Name() == "index.html" || e.Name() == "styles.css" || e.Name() == "app.js" {
						continue
					}
					link := e.Name()
					name := e.Name()
					if e.IsDir() {
						link += "/"
					} else {
						name = strings.TrimSuffix(name, ".html")
					}
					key := filepath.Join(path, e.Name())
					dashboardEntries = append(dashboardEntries, Entry{
						Name:    name,
						Link:    link,
						IsDir:   e.IsDir(),
						Summary: summaries[key],
					})
				}

				var parentPath string
				if rel != "." {
					parentPath = "../index.html"
				}

				staticPath := ""
				// Calculate relative path to root for styles.css and app.js
				if rel != "." {
					depth := strings.Count(rel, string(os.PathSeparator)) + 1
					for i := 0; i < depth; i++ {
						staticPath += "../"
					}
				}

				tplContent, err := os.ReadFile("tpl/index.html")
				if err != nil {
					return nil
				}
				tpl, err := template.New("index").Parse(string(tplContent))
				if err != nil {
					return nil
				}

				var buf bytes.Buffer
				tpl.Execute(&buf, map[string]interface{}{
					"Entries":    dashboardEntries,
					"ParentPath": parentPath,
					"StaticPath": staticPath,
				})

				indexPath := filepath.Join(path, "index.html")
				os.WriteFile(indexPath, buf.Bytes(), 0644)
				generatedCount++
			}
			return nil
		})

		fmt.Printf("Build complete: %d files built, %d files copied, %d indices generated in %v\n", builtCount, copiedCount, generatedCount, time.Since(start))
	},
}

type Breadcrumb struct {
	Name string
	Link string
}

type LessonData struct {
	Title         string
	Content       template.HTML
	StaticPath    string
	DashboardPath string
	Breadcrumbs   []Breadcrumb
	Debug         bool
	Summary       template.HTML
}

func processNoteFile(notePath, outPath string) (template.HTML, error) {
	content, err := os.ReadFile(notePath)
	if err != nil {
		return "", err
	}

	p := NewParser(string(content))
	ast, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parsing error: %w", err)
	}

	outDir := filepath.Dir(outPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	relToOut, _ := filepath.Rel("out", outPath)
	absDir := "/"
	if filepath.Dir(relToOut) != "." {
		absDir = "/" + filepath.ToSlash(filepath.Dir(relToOut)) + "/"
	}

	htmlContent := renderNode(ast, absDir, "", false)

	// Check if <lesson> is present
	var lessonNode *Node
	for _, child := range ast.Children {
		if child.Name == "lesson" {
			lessonNode = child
			break
		}
	}
	if lessonNode == nil {
		return "", fmt.Errorf("missing <lesson> tag in %s", notePath)
	}

	// Extract summary if present
	summary := template.HTML("")
	for _, child := range lessonNode.Children {
		if child.Name == "summary" {
			summary = template.HTML(renderNode(child, absDir, "", false))
			break
		}
	}

	title := filepath.Base(notePath)
	title = strings.TrimSuffix(title, ".note")

	// Calculate relative path to root for static assets
	rel, _ := filepath.Rel("out", outDir)
	staticPath := ""
	if rel != "." {
		depth := strings.Count(rel, string(os.PathSeparator)) + 1
		for i := 0; i < depth; i++ {
			staticPath += "../"
		}
	}
	dashboardPath := staticPath

	var breadcrumbs []Breadcrumb
	breadcrumbs = append(breadcrumbs, Breadcrumb{Name: "Dashboard", Link: dashboardPath + "index.html"})

	if rel != "." {
		parts := strings.Split(rel, string(os.PathSeparator))
		for i, part := range parts {
			link := ""
			for j := 0; j < len(parts)-i-1; j++ {
				link += "../"
			}
			if link == "" {
				link = "./"
			}
			breadcrumbs = append(breadcrumbs, Breadcrumb{
				Name: strings.Title(part),
				Link: link + "index.html",
			})
		}
	}
	breadcrumbs = append(breadcrumbs, Breadcrumb{Name: title, Link: ""})

	tplContent, err := os.ReadFile("tpl/lesson.html")
	if err != nil {
		return "", err
	}

	tpl, err := template.New("lesson").Parse(string(tplContent))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, LessonData{
		Title:         title,
		Content:       template.HTML(htmlContent),
		StaticPath:    staticPath,
		DashboardPath: dashboardPath,
		Breadcrumbs:   breadcrumbs,
		Debug:         debugMode,
		Summary:       summary,
	})
	if err != nil {
		return "", err
	}

	return summary, os.WriteFile(outPath, buf.Bytes(), 0644)
}

func expandCompactRefs(input string) string {
	// Custom split that respects brackets
	var parts []string
	var current strings.Builder
	depth := 0
	for i := 0; i < len(input); i++ {
		if input[i] == '[' {
			depth++
			current.WriteByte(input[i])
		} else if input[i] == ']' {
			depth--
			current.WriteByte(input[i])
		} else if unicode.IsSpace(rune(input[i])) && depth == 0 {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(input[i])
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	var expanded []string
	for _, part := range parts {
		expanded = append(expanded, expand(part)...)
	}
	result := strings.Join(expanded, " ")
	if input != "" && debugMode {
		fmt.Printf("DEBUG: Expanding ref '%s' -> '%s'\n", input, result)
	}
	return result
}

func expand(s string) []string {
	start := strings.Index(s, "[")
	if start == -1 {
		return []string{s}
	}

	// Find matching ]
	end := -1
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '[' {
			depth++
		} else if s[i] == ']' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return []string{s} // Invalid syntax, return as is
	}

	prefix := s[:start]
	suffix := s[end+1:]
	content := s[start+1 : end]

	// Split content by commas, but only at top level
	var items []string
	currStart := 0
	currDepth := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '[' {
			currDepth++
		} else if content[i] == ']' {
			currDepth--
		} else if content[i] == ',' && currDepth == 0 {
			items = append(items, content[currStart:i])
			currStart = i + 1
		}
	}
	items = append(items, content[currStart:])

	var results []string
	for _, item := range items {
		subExp := expand(prefix + strings.TrimSpace(item) + suffix)
		results = append(results, subExp...)
	}
	return results
}

func renderNode(node *Node, absDir string, currentImgContext string, inMath bool) string {
	if node == nil {
		return ""
	}
	if node.Type == "text" {
		if inMath {
			return node.Content
		}
		return template.HTMLEscapeString(node.Content)
	}

	var sb strings.Builder

	// Image context tracking downward
	if imgAttr, ok := node.Attributes["image"]; ok {
		imgName := imgAttr
		if after, ok0 := strings.CutPrefix(imgName, "images/"); ok0 {
			imgName = after
		}
		// Force .jpg extension for the output
		imgName = strings.TrimSuffix(imgName, filepath.Ext(imgName)) + ".jpg"
		currentImgContext = absDir + "images/" + imgName
	}

	isBox := node.Name == "box"
	isMath := node.Name == "math"
	isTheorem := node.Name == "theorem" || node.Name == "lemma" || node.Name == "definition" || node.Name == "proposition" || node.Name == "corollary" || node.Name == "exercise"
	isImage := node.Name == "image"
	isListElement := node.Name == "itemize" || node.Name == "enumerate" || node.Name == "item"
	isInline := node.Name == "strong" || node.Name == "emph" || node.Name == "a"

	if isImage {
		srcAttr := node.Attributes["src"]
		if srcAttr == "" && currentImgContext != "" {
			srcAttr = currentImgContext
		} else if strings.HasPrefix(srcAttr, "images/") {
			srcAttr = absDir + srcAttr
		} else if srcAttr != "" {
			srcAttr = absDir + "images/" + srcAttr
		}
		
		// Force .jpg extension for the output
		if strings.Contains(srcAttr, "images/") {
			srcAttr = strings.TrimSuffix(srcAttr, filepath.Ext(srcAttr)) + ".jpg"
		}

		sb.WriteString(fmt.Sprintf(`<div class="inline-image-crop" data-src="%s" data-img="%s" data-top="%s" data-right="%s" data-bottom="%s" data-left="%s"></div>`,
			template.HTMLEscapeString(srcAttr),
			template.HTMLEscapeString(srcAttr),
			template.HTMLEscapeString(node.Attributes["top"]),
			template.HTMLEscapeString(node.Attributes["right"]),
			template.HTMLEscapeString(node.Attributes["bottom"]),
			template.HTMLEscapeString(node.Attributes["left"]),
		))
	} else if isBox {
		uid := node.Attributes["uid"]
		top := node.Attributes["top"]
		right := node.Attributes["right"]
		bottom := node.Attributes["bottom"]
		left := node.Attributes["left"]

		idAttr := ""
		if uid != "" {
			idAttr = fmt.Sprintf(` id="%s"`, template.HTMLEscapeString(uid))
		}

		sb.WriteString(fmt.Sprintf(`<div%s class="box-text" data-img="%s" data-top="%s" data-right="%s" data-bottom="%s" data-left="%s">`,
			idAttr,
			template.HTMLEscapeString(currentImgContext),
			template.HTMLEscapeString(top),
			template.HTMLEscapeString(right),
			template.HTMLEscapeString(bottom),
			template.HTMLEscapeString(left),
		))
	} else if isMath {
		displayAttr := "false"
		if val, ok := node.Attributes["display"]; ok && val == "true" {
			displayAttr = "true"
		}
		sb.WriteString(fmt.Sprintf(`<span class="math" data-display="%s">`, displayAttr))

		if displayAttr == "true" {
			// Special handling for display math to trim indentation
			var mathContent strings.Builder
			for _, child := range node.Children {
				mathContent.WriteString(renderNode(child, absDir, currentImgContext, true))
			}
			trimmed := trimCommonIndent(mathContent.String())
			sb.WriteString(trimmed)

			// Skip normal child rendering since we just did it
			sb.WriteString(`</span>`)
			return sb.String()
		}
	} else if node.Name == "reword" {
		ref := expandCompactRefs(node.Attributes["ref"])
		refAttr := ""
		if ref != "" {
			refAttr = fmt.Sprintf(` data-ref="%s"`, template.HTMLEscapeString(ref))
		}
		sb.WriteString(fmt.Sprintf(`<div class="reword"%s>`, refAttr))
	} else if node.Name == "summary" {
		sb.WriteString(`<div class="summary-block hidden">`)
	} else if isTheorem {
		sb.WriteString(fmt.Sprintf(`<div class="%s">`, node.Name))
	} else if node.Name == "itemize" {
		sb.WriteString(`<ul class="list-disc pl-8 my-2">`)
	} else if node.Name == "enumerate" {
		sb.WriteString(`<ol class="list-decimal pl-8 my-2">`)
	} else if node.Name == "item" {
		sb.WriteString(`<li>`)
	} else if node.Name == "strong" {
		sb.WriteString(`<strong>`)
	} else if node.Name == "emph" {
		sb.WriteString(`<em>`)
	} else if node.Name == "a" {
		hrefAttr := ""
		if href, ok := node.Attributes["href"]; ok {
			hrefAttr = fmt.Sprintf(` href="%s"`, template.HTMLEscapeString(href))
		}
		sb.WriteString(fmt.Sprintf(`<a%s>`, hrefAttr))
	} else if node.Name == "section" {
		title := node.Attributes["title"]
		if title != "" {
			sb.WriteString(fmt.Sprintf(`<h2>%s</h2>`, template.HTMLEscapeString(title)))
		}
		sb.WriteString(`<div class="section">`)
	} else if node.Name == "root" {
		// Just render children
	} else {
		// Generic fallback
		sb.WriteString(fmt.Sprintf(`<div class="%s">`, node.Name))
	}

	for _, child := range node.Children {
		sb.WriteString(renderNode(child, absDir, currentImgContext, inMath || isMath))
	}

	if (isBox || isTheorem || (node.Name != "root" && node.Name != "section" && !isBox && !isMath && !isTheorem && !isImage && node.Name != "reword" && !isListElement && !isInline)) && node.Name != "math" {
		sb.WriteString(`</div>`)
	} else if node.Name == "section" {
		sb.WriteString(`</div>`)
	} else if node.Name == "math" {
		sb.WriteString(`</span>`)
	} else if node.Name == "reword" {
		sb.WriteString(`</div>`)
	} else if node.Name == "summary" {
		// Summary tag itself is just a container, we don't render a wrapper div for it
		// but we might want to hide it in the main lesson view if it's already in the dashboard
		// Actually, let's wrap it in a hidden div so it doesn't show up twice
		sb.WriteString(`</div>`)
	} else if node.Name == "itemize" {
		sb.WriteString(`</ul>`)
	} else if node.Name == "enumerate" {
		sb.WriteString(`</ol>`)
	} else if node.Name == "item" {
		sb.WriteString(`</li>`)
	} else if node.Name == "strong" {
		sb.WriteString(`</strong>`)
	} else if node.Name == "emph" {
		sb.WriteString(`</em>`)
	} else if node.Name == "a" {
		sb.WriteString(`</a>`)
	}

	return sb.String()
}

func trimCommonIndent(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return s
	}

	// Trim first newline if it's empty
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Trim last line if it's just whitespace
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for _, r := range line {
			if r == ' ' || r == '\t' {
				indent++
			} else {
				break
			}
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.Join(lines, "\n")
	}

	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
}

func processImage(srcPath, dstPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	const maxDim = 1920
	if width > maxDim || height > maxDim {
		var newWidth, newHeight int
		if width > height {
			newWidth = maxDim
			newHeight = (height * maxDim) / width
		} else {
			newHeight = maxDim
			newWidth = (width * maxDim) / height
		}

		newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.BiLinear.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = newImg
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
}
