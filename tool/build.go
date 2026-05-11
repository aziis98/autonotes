package autonotes

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

	"github.com/spf13/cobra"
	"golang.org/x/image/draw"
)

var BuildCmd = &cobra.Command{
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
							if DebugMode {
								fmt.Printf("Skipping %s (up to date)\n", path)
							}
							copiedCount++
							return nil
						}
					}

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
						fmt.Printf("Copying %s to %s\n", path, outPath)
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

	renderer := &HTMLRenderer{absDir: absDir}
	ast.Accept(renderer)
	htmlContent := renderer.String()

	// Check if <lesson> is present
	root, _ := ast.(*BlockNode)
	lessonNode := root.FindChild("lesson")
	if lessonNode == nil {
		return "", fmt.Errorf("missing <lesson> tag in %s", notePath)
	}

	// Extract summary if present
	summary := template.HTML("")
	if b := lessonNode.FindChild("summary"); b != nil {
		r := &HTMLRenderer{absDir: absDir}
		b.Accept(r)
		summary = template.HTML(r.String())
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
		Debug:         DebugMode,
		Summary:       summary,
	})
	if err != nil {
		return "", err
	}

	return summary, os.WriteFile(outPath, buf.Bytes(), 0644)
}

func expandCompactRefs(input string) string {
	expanded := ExpandRefs(input)
	result := strings.Join(expanded, " ")
	if input != "" && DebugMode {
		fmt.Printf("DEBUG: Expanding ref '%s' -> '%s'\n", input, result)
	}
	return result
}

type HTMLRenderer struct {
	sb                strings.Builder
	absDir            string
	currentImgContext string
	inMath            bool
}

func (r *HTMLRenderer) String() string {
	return r.sb.String()
}

func (r *HTMLRenderer) VisitText(n *TextNode) {
	if r.inMath {
		r.sb.WriteString(n.Content)
	} else {
		r.sb.WriteString(template.HTMLEscapeString(n.Content))
	}
}

func (r *HTMLRenderer) VisitBlock(n *BlockNode) {
	// Image context tracking downward
	if imgAttr, ok := n.Attributes["image"]; ok {
		imgName := imgAttr
		if after, ok0 := strings.CutPrefix(imgName, "images/"); ok0 {
			imgName = after
		}
		// Force .jpg extension for the output
		imgName = strings.TrimSuffix(imgName, filepath.Ext(imgName)) + ".jpg"
		r.currentImgContext = r.absDir + "images/" + imgName
	}

	isBox := n.Name == "box"
	isMath := n.Name == "math"
	isTheorem := n.Name == "theorem" || n.Name == "lemma" || n.Name == "definition" || n.Name == "proposition" || n.Name == "corollary" || n.Name == "exercise"
	isImage := n.Name == "image"
	isListElement := n.Name == "itemize" || n.Name == "enumerate" || n.Name == "item"
	isInline := n.Name == "strong" || n.Name == "emph" || n.Name == "a"

	if isImage {
		srcAttr := n.Attributes["src"]
		if srcAttr == "" && r.currentImgContext != "" {
			srcAttr = r.currentImgContext
		} else if strings.HasPrefix(srcAttr, "images/") {
			srcAttr = r.absDir + srcAttr
		} else if srcAttr != "" {
			srcAttr = r.absDir + "images/" + srcAttr
		}

		// Force .jpg extension for the output
		if strings.Contains(srcAttr, "images/") {
			srcAttr = strings.TrimSuffix(srcAttr, filepath.Ext(srcAttr)) + ".jpg"
		}

		fmt.Fprintf(&r.sb, `<div class="inline-image-crop" data-src="%s" data-img="%s" data-top="%s" data-right="%s" data-bottom="%s" data-left="%s"></div>`,
			template.HTMLEscapeString(srcAttr),
			template.HTMLEscapeString(srcAttr),
			template.HTMLEscapeString(n.Attr("top")),
			template.HTMLEscapeString(n.Attr("right")),
			template.HTMLEscapeString(n.Attr("bottom")),
			template.HTMLEscapeString(n.Attr("left")),
		)
	} else if isBox {
		uid := n.Attr("uid")
		top := n.Attr("top")
		right := n.Attr("right")
		bottom := n.Attr("bottom")
		left := n.Attr("left")

		idAttr := ""
		if uid != "" {
			idAttr = fmt.Sprintf(` id="%s"`, template.HTMLEscapeString(uid))
		}

		r.sb.WriteString(fmt.Sprintf(`<div%s class="box-text" data-img="%s" data-top="%s" data-right="%s" data-bottom="%s" data-left="%s">`,
			idAttr,
			template.HTMLEscapeString(r.currentImgContext),
			template.HTMLEscapeString(top),
			template.HTMLEscapeString(right),
			template.HTMLEscapeString(bottom),
			template.HTMLEscapeString(left),
		))
	} else if isMath {
		displayAttr := "false"
		if val, ok := n.Attributes["display"]; ok && val == "true" {
			displayAttr = "true"
		}
		fmt.Fprintf(&r.sb, `<span class="math" data-display="%s">`, displayAttr)

		if displayAttr == "true" {
			// Special handling for display math to trim indentation
			mathRenderer := &HTMLRenderer{absDir: r.absDir, currentImgContext: r.currentImgContext, inMath: true}
			for _, child := range n.Children {
				child.Accept(mathRenderer)
			}
			trimmed := trimCommonIndent(mathRenderer.String())
			r.sb.WriteString(trimmed)

			// Skip normal child rendering since we just did it
			r.sb.WriteString(`</span>`)
			return
		}
	} else if n.Name == "reword" {
		ref := expandCompactRefs(n.Attr("ref"))
		refAttr := ""
		if ref != "" {
			refAttr = fmt.Sprintf(` data-ref="%s"`, template.HTMLEscapeString(ref))
		}
		fmt.Fprintf(&r.sb, `<div class="reword"%s>`, refAttr)
	} else if n.Name == "summary" {
		fmt.Fprintf(&r.sb, `<div class="summary-block hidden">`)
	} else if isTheorem {
		fmt.Fprintf(&r.sb, `<div class="%s">`, n.Name)
	} else if n.Name == "itemize" {
		fmt.Fprintf(&r.sb, `<ul class="list-disc pl-8 my-2">`)
	} else if n.Name == "enumerate" {
		fmt.Fprintf(&r.sb, `<ol class="list-decimal pl-8 my-2">`)
	} else if n.Name == "item" {
		fmt.Fprintf(&r.sb, `<li>`)
	} else if n.Name == "strong" {
		fmt.Fprintf(&r.sb, `<strong>`)
	} else if n.Name == "emph" {
		fmt.Fprintf(&r.sb, `<em>`)
	} else if n.Name == "a" {
		hrefAttr := ""
		if href, ok := n.Attributes["href"]; ok {
			hrefAttr = fmt.Sprintf(` href="%s"`, template.HTMLEscapeString(href))
		}
		fmt.Fprintf(&r.sb, `<a%s>`, hrefAttr)
	} else if n.Name == "section" {
		title := n.Attributes["title"]
		if title != "" {
			fmt.Fprintf(&r.sb, `<h2>%s</h2>`, template.HTMLEscapeString(title))
		}
		fmt.Fprintf(&r.sb, `<div class="section">`)
	} else if n.Name == "root" {
		// Just render children
	} else {
		// Generic fallback
		fmt.Fprintf(&r.sb, `<div class="%s">`, n.Name)
	}

	oldInMath := r.inMath
	if isMath {
		r.inMath = true
	}
	for _, child := range n.Children {
		child.Accept(r)
	}
	r.inMath = oldInMath

	if (isBox || isTheorem || (n.Name != "root" && n.Name != "section" && !isBox && !isMath && !isTheorem && !isImage && n.Name != "reword" && !isListElement && !isInline)) && n.Name != "math" {
		fmt.Fprintf(&r.sb, `</div>`)
	} else if n.Name == "section" {
		fmt.Fprintf(&r.sb, `</div>`)
	} else if n.Name == "math" {
		fmt.Fprintf(&r.sb, `</span>`)
	} else if n.Name == "reword" {
		fmt.Fprintf(&r.sb, `</div>`)
	} else if n.Name == "summary" {
		fmt.Fprintf(&r.sb, `</div>`)
	} else if n.Name == "itemize" {
		fmt.Fprintf(&r.sb, `</ul>`)
	} else if n.Name == "enumerate" {
		fmt.Fprintf(&r.sb, `</ol>`)
	} else if n.Name == "item" {
		fmt.Fprintf(&r.sb, `</li>`)
	} else if n.Name == "strong" {
		fmt.Fprintf(&r.sb, `</strong>`)
	} else if n.Name == "emph" {
		fmt.Fprintf(&r.sb, `</em>`)
	} else if n.Name == "a" {
		fmt.Fprintf(&r.sb, `</a>`)
	}
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

	img, format, err := image.Decode(f)
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
		fmt.Printf("Resizing %s from %dx%d (%s) to %dx%d (jpg)\n", srcPath, width, height, format, newWidth, newHeight)
	} else {
		fmt.Printf("Re-encoding %s from %s to jpg (%dx%d)\n", srcPath, format, width, height)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
}
