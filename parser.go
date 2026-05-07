package main

import (
	"errors"
	"fmt"
	"unicode"
)

type Node struct {
	Type       string // "element" or "text"
	Name       string
	Attributes map[string]string
	Content    string
	Children   []*Node
}

type Parser struct {
	input string
	pos   int
}

func NewParser(input string) *Parser {
	return &Parser{input: input, pos: 0}
}

func (p *Parser) Parse() (*Node, error) {
	root := &Node{Type: "element", Name: "root", Attributes: make(map[string]string)}
	for p.pos < len(p.input) {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			root.Children = append(root.Children, child)
		}
	}
	return root, nil
}

func (p *Parser) parseNode() (*Node, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, nil
	}

	if p.isTagStart() {
		// Is it a closing tag? We shouldn't see it here.
		if p.pos+1 < len(p.input) && p.input[p.pos+1] == '/' {
			return nil, errors.New("unexpected closing tag")
		}
		return p.parseElement()
	}
	return p.parseText()
}

func (p *Parser) isTagStart() bool {
	if p.pos >= len(p.input) || p.input[p.pos] != '<' {
		return false
	}
	if p.pos+1 >= len(p.input) {
		return false
	}
	next := rune(p.input[p.pos+1])
	return unicode.IsLetter(next) || next == '/'
}

func (p *Parser) parseElement() (*Node, error) {
	// Skip '<'
	p.pos++

	name := p.readUntilWhitespaceOr('>', '/')
	if name == "" {
		return nil, errors.New("empty tag name")
	}

	node := &Node{
		Type:       "element",
		Name:       name,
		Attributes: make(map[string]string),
	}

	// Parse attributes
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, errors.New("unclosed tag")
		}
		if p.input[p.pos] == '>' || p.input[p.pos] == '/' {
			break
		}
		attrName := p.readUntilWhitespaceOr('=', '>', '/')
		if attrName == "" {
			break // Could happen if invalid chars
		}
		p.skipWhitespace()
		var attrVal string
		if p.pos < len(p.input) && p.input[p.pos] == '=' {
			p.pos++ // skip '='
			p.skipWhitespace()
			
			if p.pos < len(p.input) && (p.input[p.pos] == '"' || p.input[p.pos] == '\'') {
				quote := p.input[p.pos]
				p.pos++
				start := p.pos
				for p.pos < len(p.input) && p.input[p.pos] != quote {
					p.pos++
				}
				attrVal = p.input[start:p.pos]
				if p.pos < len(p.input) {
					p.pos++ // skip quote
				}
			} else {
				attrVal = p.readUntilWhitespaceOr('>', '/')
			}
		}
		node.Attributes[attrName] = attrVal
	}

	selfClosing := false
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		selfClosing = true
		p.pos++
	}
	
	if p.pos < len(p.input) && p.input[p.pos] == '>' {
		p.pos++
	}

	if selfClosing {
		return node, nil
	}

	// Parse children
	for p.pos < len(p.input) {
		// p.skipWhitespace() // Wait, skipping whitespace might drop text spaces, but text node parsing handles its own.
		// Let's not skip whitespace before parsing text, only check if it's a tag
		
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unclosed tag <%s>", name)
		}
		
		if p.isTagStart() {
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '/' {
				// closing tag
				p.pos += 2
				closeName := p.readUntilWhitespaceOr('>')
				p.skipWhitespace()
				if p.pos < len(p.input) && p.input[p.pos] == '>' {
					p.pos++
				}
				if closeName != name {
					return nil, fmt.Errorf("mismatched closing tag, expected </%s> got </%s>", name, closeName)
				}
				break
			} else {
				child, err := p.parseElement()
				if err != nil {
					return nil, err
				}
				node.Children = append(node.Children, child)
			}
		} else {
			child, err := p.parseText()
			if err != nil {
				return nil, err
			}
			if child != nil {
				node.Children = append(node.Children, child)
			}
		}
	}

	return node, nil
}

func (p *Parser) parseText() (*Node, error) {
	start := p.pos
	for p.pos < len(p.input) && !p.isTagStart() {
		p.pos++
	}
	content := p.input[start:p.pos]
	return &Node{
		Type:    "text",
		Content: content,
	}, nil
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *Parser) readUntilWhitespaceOr(chars ...byte) string {
	start := p.pos
	for p.pos < len(p.input) {
		r := p.input[p.pos]
		if unicode.IsSpace(rune(r)) {
			break
		}
		match := false
		for _, c := range chars {
			if r == c {
				match = true
				break
			}
		}
		if match {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}
