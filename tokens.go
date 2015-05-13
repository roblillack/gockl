package gockl

import (
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

func (t StartElementToken) Attributes() []Attribute {
	list := []Attribute{}

	z := &attributeTokenizer{Input: string(t)[1 : len(t)-1]}
	// eat the element name
	z.shiftUntilSpace()

	for {
		a, err := z.Next()
		if err != nil {
			break
		}
		list = append(list, a)
	}

	return list
}

type EndElementToken string

func (t EndElementToken) Raw() string {
	return string(t)
}

func (t EndElementToken) Name() string {
	return string(t)[2 : len(t)-1]
}

type EmptyElementToken string

func (t EmptyElementToken) Raw() string {
	return string(t)
}
