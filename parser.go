package main

import (
	"errors"
	"fmt"
	"strings"
)

type Node interface{ isNode() }
type AttrValue interface{ isAttrValue() }

type Expression string
type Text string

func (Expression) isAttrValue() {}
func (Text) isAttrValue()       {}

func (Element) isNode()    {}
func (Expression) isNode() {}
func (Text) isNode()       {}

type Element struct {
	tag      string
	attrs    []Attr
	children []Node
}

type Attr struct {
	Name  string
	Value AttrValue
}

type ParseResult struct {
	isXML bool
	Node  Node
	End   int
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
		return fmt.Errorf("wanted %v at %d", prefix, p.index)
	}
	p.index += len(prefix)
	return nil
}

func (p *Parser) element() (Element, error) {
	p.expect("<")
	element := Element{}
	p.skipWhitespace()

	attrs := make([]Attr, 0)
	tag, err := p.name()
	if err != nil {
		return element, err
	}

	element.tag = tag
	element.attrs = attrs

	for {
		p.skipWhitespace()

		// fin du node autofermant,
		// on return car aucun enfant.
		if p.peek("/>") {
			p.index += 2
			return element, nil
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
			return element, err
		}

		p.skipWhitespace()
		if err = p.expect("="); err != nil {
			return element, err
		}

		p.skipWhitespace()

		var value string
		switch {
		case p.peek("\""):
			value, err = p.delimited('"', '"')
			if err != nil {
				return element, err
			}

			element.attrs = append(element.attrs, Attr{
				Name:  key,
				Value: Text(value),
			})
		case p.peek("'"):
			value, err = p.delimited('\'', '\'')
			if err != nil {
				return element, err
			}

			element.attrs = append(element.attrs, Attr{
				Name:  key,
				Value: Text(value),
			})
		case p.peek("{"):
			value, err = p.delimited('{', '}')
			if err != nil {
				return element, err
			}

			element.attrs = append(element.attrs, Attr{
				Name:  key,
				Value: Expression(value),
			})
		default:
			return element, fmt.Errorf("valeur d'attribut attendue à la position %d", p.index)
		}
		p.index++
	}

	children := make([]Node, 0)
	for !p.peek("</") {
		if p.index >= len(p.content) {
			return element, errors.New("balise fermante manquante")
		}
		switch {
		case p.peek("<"):
			child, err := p.element()
			if err != nil {
				return element, err
			}
			children = append(children, child)
		case p.peek("{"):
			expr, err := p.delimited('{', '}')
			if err != nil {
				return element, err
			}
			children = append(children, Expression(expr))
		default:
			text := p.takeWhile(func(b byte) bool { return b != '<' && b != '{' })
			children = append(children, Text(text))
		}
	}

	p.index += 2
	if name, err := p.name(); name != tag || err != nil {
		return element, fmt.Errorf("closing tag does not match the opening <%v>", tag)
	}

	p.skipWhitespace()
	if err = p.expect(">"); err != nil {
		return element, err
	}

	element.children = children
	return element, nil
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

func (e Element) String() string {
	return fmt.Sprintf("Element(tag=%q, attrs=%v, children=%v)", e.tag, e.attrs, e.children)
}

func (e Expression) String() string {
	return fmt.Sprintf("Expression(%q)", string(e))
}

func (t Text) String() string {
	return fmt.Sprintf("Text(%q)", string(t))
}

func (a Attr) String() string {
	return fmt.Sprintf("Attr(Name=%+q, Value=%v)", a.Name, a.Value)
}
