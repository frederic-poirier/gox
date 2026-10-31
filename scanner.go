package main

import "strings"

type state int

const (
	InRegularCode state = iota
	InBlockComment
	InInlineComment
	InBackticksString
	InDoubleQuotedString
	InRune
)

var (
	BackSlash   byte = '\\'
	Slash       byte = '/'
	Asterik     byte = '*'
	Space       byte = ' '
	DoubleQuote byte = '"'
	SingleQuote byte = '\''
	BackTick    byte = '`'
	LineBreak   byte = '\n'
)

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func isFollowedByReturn(content string, i int) bool {
	// si pas précédé de vide ou un ident char, il n'est pas followed.
	if i > 0 && isIdentChar(content[i-1]) {
		return false
	}
	// si pas suivit de return il ne l'est pas followed
	if !strings.HasPrefix(content[i:], "return") {
		return false
	}
	// si le return est suivit d'assez de caractère
	// et que celui-ci n'est pas ident il est followed.
	if len(content) > i+6 && isIdentChar(content[i+6]) {
		return false
	}

	return true
}

func findReturns(content string) (position int, found bool) {
	state := InRegularCode

	for i := 0; i < len(content); i++ {
		switch state {
		case InBlockComment:
			if i+1 < len(content) && content[i] == Asterik && content[i+1] == Slash {
				state = InRegularCode
				i++
			}
		case InInlineComment:
			if content[i] == LineBreak {
				state = InRegularCode
			}
		case InRune:
			if content[i] == BackSlash {
				i++
				continue
			}
			if content[i] == SingleQuote {
				state = InRegularCode
			}
		case InDoubleQuotedString:
			if content[i] == BackSlash {
				i++
				continue
			}
			if content[i] == DoubleQuote {
				state = InRegularCode
			}
		case InBackticksString:
			if content[i] == BackTick {
				state = InRegularCode
			}
		case InRegularCode:
			if isFollowedByReturn(content, i) {
				return i + 6, true
			}

			switch content[i] {
			case Slash:
				if len(content) <= i+1 {
					break
				}

				if content[i+1] == Asterik {
					state = InBlockComment
					i++
				} else if content[i+1] == Slash {
					state = InInlineComment
					i++
				}

			case DoubleQuote:
				state = InDoubleQuotedString
			case SingleQuote:
				state = InRune
			case BackTick:
				state = InBackticksString
			}
		}
	}

	return position, found
}
