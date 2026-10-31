package main

import (
	"errors"
	"fmt"
	"strings"
)

type NodeKind int

type ParseResult struct {
	isXML bool
	Node  *Node
	End   int
}

const (
	KindElement NodeKind = iota // ex: <div>...</div>
	KindText                    // ex: "Hello "
	KindExpr                    // ex: {name}
)

type Attr struct {
	Name  string
	Value string // ex: `"box"` ou `{2026}`
}

type Node struct {
	Kind     NodeKind
	Tag      string
	Attrs    []Attr
	Text     string // utilisé pour KindText et KindExpr
	Children []*Node
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isAlphanumeric(b byte) bool {
	return isLetter(b) || ('0' <= b && b <= '9')
}

func isLegalChar(b byte) bool {
	return b == '_' || b == '-' || b == ':' || b == '.'
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

type Parser struct {
	content string
	index   int
}

func NewParser(content string) *Parser {
	return &Parser{content: content, index: 0}
}

func (p *Parser) current() byte {
	return p.content[p.index]
}

func (p *Parser) skipWhitespace() {
	for p.index < len(p.content) && isWhitespace(p.current()) {
		p.index++
	}
}

func (p *Parser) takeWhile(cond func(byte) bool) string {
	start := p.index
	for p.index < len(p.content) && cond(p.content[p.index]) {
		p.index++
	}
	return p.content[start:p.index]
}

func (p *Parser) delimited(open, close byte) (string, error) {
	if p.index >= len(p.content) || p.content[p.index] != open {
		return "", fmt.Errorf("%q attendu à la position %d", open, p.index)
	}
	p.index++
	start := p.index
	for p.index < len(p.content) && p.content[p.index] != close {
		p.index++
	}
	if p.index >= len(p.content) {
		return "", fmt.Errorf("%q fermant manquant", close)
	}
	value := p.content[start:p.index]
	p.index++
	return value, nil
}

func (p *Parser) name() (string, error) {
	n := p.takeWhile(func(b byte) bool { return isAlphanumeric(b) || isLegalChar(b) })
	if n == "" {
		return "", fmt.Errorf("nom attendu à la position %d", p.index)
	}
	return n, nil
}

func (p *Parser) peek(prefix string) bool {
	return strings.HasPrefix(p.content[p.index:], prefix)
}

func (p *Parser) expect(prefix string) error {
	if !p.peek(prefix) {
		return errors.New(fmt.Sprintf("wanted %v at %d", prefix, p.index))
	}
	p.index += len(prefix)
	return nil
}

func (p *Parser) element() (*Node, error) {
	p.expect("<")
	node := &Node{}
	p.skipWhitespace()

	nodeAttrs := make([]Attr, 0)
	nodeTag, err := p.name()
	if err != nil {
		return nil, err
	}

	node.Kind = KindElement
	node.Tag = nodeTag
	node.Attrs = nodeAttrs

	for {
		p.skipWhitespace()

		// fin du node autofermant,
		// on return car aucun enfant.
		if p.peek("/>") {
			p.index += 2
			return node, nil
		}

		// fin du node
		if p.peek(">") {
			p.index++
			break
		}

		// si le prochain caractère n'est pas fermant,
		// il s'agit d'un attribut. récupère la key
		// celle-ci doit être suivit quote avec la value
		// ou d'une expression.
		key, err := p.name()
		if err != nil {
			return nil, err
		}

		p.skipWhitespace()
		if err = p.expect("="); err != nil {
			return nil, err
		}

		p.skipWhitespace()

		var value string
		switch {
		case p.peek("\""):
			value, err = p.delimited('"', '"')
		case p.peek("'"):
			value, err = p.delimited('\'', '\'')
		case p.peek("{"):
			value, err = p.delimited('{', '}')
		default:
			return nil, fmt.Errorf("valeur d'attribut attendue à la position %d", p.index)
		}

		if err != nil {
			return nil, err
		}

		// il faudrait probablement faire quelque chose pour les empty value.
		node.Attrs = append(node.Attrs, Attr{Name: key, Value: value})
		p.index++
	}

	children := make([]*Node, 0)
	for !p.peek("</") {
		if p.index >= len(p.content) {
			return nil, errors.New("balise fermante manquante")
		}
		switch {
		case p.peek("<"):
			child, err := p.element()
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		case p.peek("{"):
			expr, err := p.delimited('{', '}')
			if err != nil {
				return nil, err
			}
			children = append(children, &Node{Kind: KindExpr, Text: expr})
		default:
			text := p.takeWhile(func(b byte) bool { return b != '<' && b != '{' })
			children = append(children, &Node{Kind: KindText, Text: text})
		}
	}

	p.index += 2
	if name, err := p.name(); name != nodeTag || err != nil {
		return nil, errors.New(fmt.Sprintf("closing tag does not match the opening <%v>", nodeTag))
	}

	p.skipWhitespace()
	if err = p.expect(">"); err != nil {
		return nil, err
	}

	node.Children = children
	return node, nil
}

func (p *Parser) startsElement() bool {
	return p.index+1 < len(p.content) &&
		p.content[p.index] == '<' &&
		isLetter(p.content[p.index+1])
}

func parse(content string) (ParseResult, error) {
	p := NewParser(content)
	p.skipWhitespace()

	paren := p.peek("(")
	if paren {
		p.index++
		p.skipWhitespace()
	}

	if !p.startsElement() {
		return ParseResult{isXML: false, End: p.index}, nil
	}

	root, err := p.element()
	if err != nil {
		return ParseResult{isXML: true}, err
	}

	if paren {
		p.skipWhitespace()
		if err := p.expect(")"); err != nil {
			return ParseResult{isXML: true}, err
		}
	}

	return ParseResult{isXML: true, Node: root, End: p.index}, nil
}
