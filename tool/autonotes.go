package autonotes

import (
	"html/template"
)

var DebugMode bool

type Node interface {
	Accept(Visitor)
	Type() string
}

type TextNode struct {
	Content string
}

func (n *TextNode) Accept(v Visitor) { v.VisitText(n) }
func (n *TextNode) Type() string     { return "text" }

type BlockNode struct {
	Name       string
	Attributes map[string]string
	Children   []Node
}

func (n *BlockNode) Accept(v Visitor) { v.VisitBlock(n) }
func (n *BlockNode) Type() string     { return "block" }

func (n *BlockNode) Attr(name string) string {
	return n.Attributes[name]
}

func (n *BlockNode) FindChild(name string) *BlockNode {
	for _, child := range n.Children {
		if b, ok := child.(*BlockNode); ok && b.Name == name {
			return b
		}
	}
	return nil
}

type Visitor interface {
	VisitText(*TextNode)
	VisitBlock(*BlockNode)
}

type Parser interface {
	Parse() (Node, error)
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
