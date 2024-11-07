package gockl

import (
	"io"
	"strings"
)

var spaceChars = " \t\r\n"

type Tokenizer struct {
	Input    string
	Position int
}

func New(input string) *Tokenizer {
	return &Tokenizer{Input: input}
}

func (me *Tokenizer) shift(end string) string {
	if pos := strings.Index(me.Input[me.Position:], end); pos > -1 {
		r := me.Input[me.Position : me.Position+pos+len(end)]
		me.Position += pos + len(end)
		return r
	}

	return me.shiftUntil('<')
}

func (me *Tokenizer) shiftToTagEnd() string {
	type state uint8
	const (
		tagname     state = iota // we're in the name of the tag still
		tagcontent               // past the name, looking for attributes or the end of the tag
		attribname               // ok, started the attribute name
		attribvalue              // started an attribute value
		doublequote              // inside a double quoted attribute value
		singlequote              // inside a single quoted attribute value
	)

	var s state = tagname
	pos := me.Position
	len := len(me.Input)
	for i := pos + 1; i < len; i++ {
		curr := me.Input[i]
		if s == doublequote {
			if curr == '"' {
				s = tagname
			}
			continue
		} else if s == singlequote {
			if curr == '\'' {
				s = tagname
			}
			continue
		}

		// tagname, tagcontent, attribname, attribvalue here

		switch curr {
		case ' ', '\t', '\r', '\n':
			if s == tagname {
				s = tagcontent
			} else if s == attribvalue {
				s = tagcontent
			}
		case '=':
			if s == attribname {
				s = attribvalue
			}
		case '<':
			me.Position = i
			return me.Input[pos:i]
		case '>':
			me.Position = i + 1
			return me.Input[pos : i+1]
		case '"':
			if s == attribvalue {
				s = doublequote
			}
		case '\'':
			if s == attribvalue {
				s = singlequote
			}
		default:
			if s == tagcontent {
				s = attribname
			}
		}
	}

	// eof
	me.Position = len
	return me.Input[pos:]
}

func (me *Tokenizer) shiftUntil(next rune) string {
	if me.Position < len(me.Input) {
		if pos := strings.IndexRune(me.Input[me.Position+1:], next); pos > -1 {
			r := me.Input[me.Position : me.Position+pos+1]
			me.Position += pos + 1
			return r
		}
	}

	r := me.Input[me.Position:]
	me.Position = len(me.Input)
	return r
}

func (me *Tokenizer) has(next string) bool {
	return me.Position+len(next) <= len(me.Input) && me.Input[me.Position:me.Position+len(next)] == next
}

func (me *Tokenizer) Next() (Token, error) {
	if me.Position >= len(me.Input) {
		return nil, io.EOF
	}

	if me.Position >= len(me.Input)-3 {
		goto dunno
	}

	switch me.Input[me.Position] {
	case '<':
		switch me.Input[me.Position+1] {
		case '?':
			return ProcInstToken(me.shift("?>")), nil
		case '!':
			if me.has("<!--") {
				return CommentToken(me.shift("-->")), nil
			}

			if me.has("<![CDATA[") {
				return CDATAToken(me.shift("]]>")), nil
			}

			r := me.shift(">")
			if strings.HasPrefix(r, "<!DOCTYPE") && strings.Contains(r, "[") {
				r += me.shift("]") + me.shift(">")
			}

			return DirectiveToken(r), nil
		case '/':
			return EndElementToken(me.shift(">")), nil
		default:
			raw := me.shiftToTagEnd()

			if len(raw) >= 3 && raw[len(raw)-2] == '/' {
				return EmptyElementToken(raw), nil
			}

			return StartElementToken(raw), nil
		}
	}

dunno:

	return TextToken(me.shiftUntil('<')), nil
}
