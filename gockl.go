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

	return me.shiftUntil("<")
}

func (me *Tokenizer) shiftUntil(next string) string {
	if pos := strings.Index(me.Input[me.Position+1:], next); pos > -1 {
		r := me.Input[me.Position : me.Position+pos+1]
		me.Position += pos + 1
		return r
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
			raw := me.shift(">")

			if len(raw) >= 3 && raw[len(raw)-2] == '/' {
				return EmptyElementToken(raw), nil
			}

			return StartElementToken(raw), nil
		}
	}

dunno:

	return TextToken(me.shiftUntil("<")), nil
}
