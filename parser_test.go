package main

import (
	"reflect"
	"testing"
)

type ParseJSXReturnTest struct {
	name    string
	in      string // contenu qui suit directement le mot-clé "return"
	wantJSX bool
	wantEnd int
	wantErr bool
	wantAST *Node
}

var parseJSXReturnTests = []ParseJSXReturnTest{
	// --- 1. Early exits : ce n'est PAS du JSX (isJSX = false, err = nil) ---
	{
		name:    "empty return before newline",
		in:      "\n}",
		wantJSX: false,
	},
	{
		name:    "return go variable",
		in:      " x + 1 }",
		wantJSX: false,
	},
	{
		name:    "return go comparison",
		in:      " a < b }",
		wantJSX: false,
	},
	{
		name:    "return parenthesized go expression",
		in:      " (a + b) }",
		wantJSX: false,
	},
	{
		name:    "return parenthesized go comparison",
		in:      " (a < b) }",
		wantJSX: false,
	},
	{
		name:    "return string with html inside",
		in:      ` "<div>not jsx</div>" }`,
		wantJSX: false,
	},
	{
		name:    "less than followed by space is not a tag",
		in:      " < 5",
		wantJSX: false,
	},

	// --- 2. Vrais retours JSX valides ---
	{
		name:    "self closing tag",
		in:      " <hr />\n}",
		wantJSX: true,
		wantEnd: 7, // len(" <hr />")
		wantAST: &Node{
			Kind: KindElement,
			Tag:  "hr",
		},
	},
	{
		name:    "simple inline element with text",
		in:      " <h1>Hello</h1>\n}",
		wantJSX: true,
		wantEnd: 15, // len(" <h1>Hello</h1>")
		wantAST: &Node{
			Kind: KindElement,
			Tag:  "h1",
			Children: []*Node{
				{Kind: KindText, Text: "Hello"},
			},
		},
	},
	{
		name:    "element with string and go-expr attributes",
		in:      ` <Footer class="box" year={2026} />`,
		wantJSX: true,
		wantEnd: 35,
		wantAST: &Node{
			Kind: KindElement,
			Tag:  "Footer",
			Attrs: []Attr{
				{Name: "class", Value: `"box"`},
				{Name: "year", Value: `{2026}`},
			},
		},
	},
	{
		name:    "element with go expression child",
		in:      " <div>Hi {name}</div>",
		wantJSX: true,
		wantEnd: 21,
		wantAST: &Node{
			Kind: KindElement,
			Tag:  "div",
			Children: []*Node{
				{Kind: KindText, Text: "Hi "},
				{Kind: KindExpr, Text: "name"},
			},
		},
	},
	{
		name: "parenthesized block return with nested children",
		in: ` (
		<div class="box">
			<h1>{title}</h1>
			<hr />
		</div>
	)
}`,
		wantJSX: true,
		wantEnd: 64, // s'arrête juste après la parenthèse fermante ')'
		wantAST: &Node{
			Kind:  KindElement,
			Tag:   "div",
			Attrs: []Attr{{Name: "class", Value: `"box"`}},
			Children: []*Node{
				{
					Kind:     KindElement,
					Tag:      "h1",
					Children: []*Node{{Kind: KindExpr, Text: "title"}},
				},
				{
					Kind: KindElement,
					Tag:  "hr",
				},
			},
		},
	},

	// --- 3. Cas piégés dans le JSX (apostrophes, URLs, accolades imbriquées) ---
	{
		name:    "apostrophe and url in text",
		in:      " <p>C'est http://example.com</p>\n\tcase 3:",
		wantJSX: true,
		wantEnd: 32, // s'arrête pile après </p>
		wantAST: &Node{
			Kind: KindElement,
			Tag:  "p",
			Children: []*Node{
				{Kind: KindText, Text: "C'est http://example.com"},
			},
		},
	},
	{
		name:    "closing tag inside go string expression",
		in:      ` <div>{fmt.Sprintf("</div> }")}</div>}`,
		wantJSX: true,
		wantEnd: 37, // s'arrête après le vrai </div>, sans compter le '}' final de la func Go
		wantAST: &Node{
			Kind: KindElement,
			Tag:  "div",
			Children: []*Node{
				{Kind: KindExpr, Text: `fmt.Sprintf("</div> }")`},
			},
		},
	},

	// --- 4. Erreurs de syntaxe JSX (isJSX = true, err != nil) ---
	{
		name:    "mismatched closing tag",
		in:      " <div><h1>oups</div>",
		wantJSX: true,
		wantErr: true,
	},
	{
		name:    "unclosed tag at EOF",
		in:      " <div><h1>hello</h1>",
		wantJSX: true,
		wantErr: true,
	},
	{
		name:    "unclosed parenthesis in block return",
		in:      " (\n\t<div>hello</div>\n}",
		wantJSX: true,
		wantErr: true,
	},
}

func TestParseJSXReturn(t *testing.T) {
	for _, tt := range parseJSXReturnTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parse(tt.in)
			isJSX := result.isXML
			endPos := result.End
			node := result.Node

			if isJSX != tt.wantJSX {
				t.Fatalf("isJSX = %v, want %v", isJSX, tt.wantJSX)
			}

			if !tt.wantJSX {
				return
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if endPos != tt.wantEnd {
				t.Errorf("endPos = %d (%q), want %d (%q)", endPos, tt.in[:endPos], tt.wantEnd, tt.in[:tt.wantEnd])
			}
			if !reflect.DeepEqual(node, tt.wantAST) {
				t.Errorf("AST mismatch:\ngot:  %+v\nwant: %+v", node, tt.wantAST)
			}
		})
	}
}
