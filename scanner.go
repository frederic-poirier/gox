package main

import (
	"fmt"
	"go/scanner"
	"go/token"
	"os"
)

type (
	returnContent []string
	returnKind    int
)

const (
	returnKindBlock returnKind = iota
	returnKindInline
	returnKindEmpty
)

func main() {
	exempleFilePath := "./examples/play.gox"
	content, err := os.ReadFile(exempleFilePath)
	if err != nil {
		fmt.Println("Could not read the file content")
		os.Exit(1)
	}

	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("play", fset.Base(), len(content))
	s.Init(file, content, nil, scanner.ScanComments)

	returns := make(returnContent, 1)
	isFollowingReturnStatement := false
	kind := returnKindEmpty
	lparenCount := 0
	for {

		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		if tok == token.COMMENT || tok == token.PACKAGE {
			continue
		}

		if tok == token.RETURN {
			isFollowingReturnStatement = true
			continue
		}

		// if inside an inline return and you encounter something else
		// then semicolon, append it to the return content, otherwise we
		// are out of the return content.
		if kind == returnKindInline {
			if tok != token.SEMICOLON {
				if lit == "" {
					returns[len(returns)-1] += tok.String()
				} else {
					returns[len(returns)-1] += lit
				}
			} else {
				kind = returnKindEmpty
			}
		}

		// if inside a return block and you encounter a left parenthesis
		// add it to the count of nested parenthese.
		// each time you encounter a right parenthese if there is no nesting, it is
		// the end of the return statement, else decrease the nesting.
		if kind == returnKindBlock {
			if tok == token.SEMICOLON {
				continue
			}

			if tok == token.LPAREN {
				lparenCount++
			}

			if tok == token.RPAREN {
				if lparenCount == 0 {
					kind = returnKindEmpty
				} else {
					lparenCount--
				}
			}

			if lit == "" {
				returns[len(returns)-1] += tok.String()
			} else {
				returns[len(returns)-1] += lit
			}
		}

		// if is after return, is it block or inline.
		// set after return to false
		// if not empty add a new return string.
		if isFollowingReturnStatement {
			isFollowingReturnStatement = false
			if tok == token.LPAREN {
				kind = returnKindBlock
			} else if tok != token.GTR {
				kind = returnKindInline
			} else {
				kind = returnKindEmpty
				continue
			}

			if lit == "" {
				returns = append(returns, tok.String())
			} else {
				returns = append(returns, lit)
			}
		}
	}

	for i := range returns {
		fmt.Println(returns[i])
	}
}
