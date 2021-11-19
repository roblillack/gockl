package gockl

import (
	"strings"
)

type Token interface {
	Raw() string
}

type ElementToken interface {
	Token
	Name() string
}

type StartOrEmptyElementToken interface {
	ElementToken
	Attributes() []Attribute
	Attribute(name string) (string, bool)
}

type TextToken string

var _ Token = TextToken("")

func (t TextToken) Raw() string {
	return string(t)
}

type CDATAToken string

var _ Token = CDATAToken("")

func (t CDATAToken) Raw() string {
	return string(t)
}

type CommentToken string

func (t CommentToken) Raw() string {
	return string(t)
}

type DirectiveToken string

var _ Token = DirectiveToken("")

func (t DirectiveToken) Raw() string {
	return string(t)
}

type ProcInstToken string

var _ Token = ProcInstToken("")

func (t ProcInstToken) Raw() string {
	return string(t)
}

type StartElementToken string

var _ StartOrEmptyElementToken = StartElementToken("")

func (t StartElementToken) Raw() string {
	return string(t)
}

func (t StartElementToken) Name() string {
	if idx := strings.IndexAny(string(t)[1:], " \t\r\n>/"); idx > -1 {
		return string(t)[1 : 1+idx]
	}
	return string(t)[1:]
}

func (t StartElementToken) Attributes() []Attribute {
	if len(t) <= 1 {
		return []Attribute{}
	}

	return getAttributes(string(t)[1 : len(t)-1])
}

func (t StartElementToken) Attribute(name string) (string, bool) {
	return getAttribute(string(t)[1:len(t)-1], name)
}

type EndElementToken string

var _ EndElementToken = EndElementToken("")

func (t EndElementToken) Raw() string {
	return string(t)
}

func (t EndElementToken) Name() string {
	if len(t) <= 2 {
		return ""
	}

	return string(t)[2 : len(t)-1]
}

type EmptyElementToken string

var _ StartOrEmptyElementToken = EmptyElementToken("")

func (t EmptyElementToken) Raw() string {
	return string(t)
}

func (t EmptyElementToken) Name() string {
	return StartElementToken(t).Name()
}

func (t EmptyElementToken) Attributes() []Attribute {
	return getAttributes(string(t)[1 : len(t)-2])
}

func (t EmptyElementToken) Attribute(name string) (string, bool) {
	return getAttribute(string(t)[1:len(t)-2], name)
}
