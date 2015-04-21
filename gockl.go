package gockl

import (
	"io"
	"strings"
)

type Token interface {
	Raw() string
}

type ElementToken interface {
	Name() string
}

type TextToken string

func (t TextToken) Raw() string {
	return string(t)
}

type CDATAToken string

func (t CDATAToken) Raw() string {
	return string(t)
}

type CommentToken string

func (t CommentToken) Raw() string {
	return string(t)
}

type DirectiveToken string

func (t DirectiveToken) Raw() string {
	return string(t)
}

type ProcInstToken string

func (t ProcInstToken) Raw() string {
	return string(t)
}

type StartElementToken string

func (t StartElementToken) Raw() string {
	return string(t)
}

func (t StartElementToken) Name() string {
	if idx := strings.IndexAny(string(t)[1:], " \t\r\n>/"); idx > -1 {
		return string(t)[1 : 1+idx]
	}
	return string(t)[1:]
}

type EndElementToken string

func (t EndElementToken) Raw() string {
	return string(t)
}

func (t EndElementToken) Name() string {
	return string(t)[2 : len(t)-1]
}

type StartEndElementToken string

func (t StartEndElementToken) Raw() string {
	return string(t)
}

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
			if me.has("<![CDATA[") {
				return CDATAToken(me.shift("]]>")), nil
			}
			if me.Input[me.Position+2:me.Position+4] != "--" {
				goto dunno
			}
			return CommentToken(me.shift("-->")), nil
		case '/':
			return EndElementToken(me.shift(">")), nil
		default:
			raw := me.shift(">")

			if len(raw) >= 3 && raw[len(raw)-2] == '/' {
				return StartEndElementToken(raw), nil
			}

			return StartElementToken(raw), nil
		}
	}

dunno:

	return TextToken(me.shiftUntil("<")), nil
}
