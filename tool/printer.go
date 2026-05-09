package autonotes

import (
	"fmt"
	"io"
	"strings"
)

type Printer struct {
	w      io.Writer
	indent int
}

func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w, indent: 0}
}

func (p *Printer) Print(n Node) {
	n.Accept(p)
}

func (p *Printer) VisitText(n *TextNode) {
	content := strings.TrimSpace(n.Content)
	if content == "" {
		return
	}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fmt.Fprintf(p.w, "%s%s\n", strings.Repeat("  ", p.indent), strings.TrimSpace(line))
	}
}

func (p *Printer) VisitBlock(n *BlockNode) {
	var attrs []string
	for k, v := range n.Attributes {
		attrs = append(attrs, fmt.Sprintf("%s=%q", k, v))
	}
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	if len(n.Children) == 0 {
		fmt.Fprintf(p.w, "%s<%s%s />\n", strings.Repeat("  ", p.indent), n.Name, attrStr)
		return
	}

	fmt.Fprintf(p.w, "%s<%s%s>\n", strings.Repeat("  ", p.indent), n.Name, attrStr)
	p.indent++
	for _, child := range n.Children {
		child.Accept(p)
	}
	p.indent--
	fmt.Fprintf(p.w, "%s</%s>\n", strings.Repeat("  ", p.indent), n.Name)
}
