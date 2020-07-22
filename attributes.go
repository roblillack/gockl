package gockl

import (
	"io"
	"strings"
)

type Attribute struct {
	Name    string
	Content string
}

type attributeTokenizer struct {
	Input    string
	Position int
}

func (me *attributeTokenizer) shiftUntil(next string) string {
	if pos := strings.Index(me.Input[me.Position+1:], next); pos > -1 {
		r := me.Input[me.Position : me.Position+pos+1]
		me.Position += pos + 1
		return r
	}

	r := me.Input[me.Position:]
	me.Position = len(me.Input)
	return r
}

func (me *attributeTokenizer) shiftUntilSpace() string {
	if me.Position+1 >= len(me.Input) {
		goto whaa
	}

	if pos := strings.IndexAny(me.Input[me.Position+1:], spaceChars); pos > -1 {
		r := me.Input[me.Position : me.Position+pos+1]
		me.Position += pos + 1
		return r
	}

whaa:
	r := me.Input[me.Position:]
	me.Position = len(me.Input)
	return r
}

func (me *attributeTokenizer) eatSpace() string {
	if me.Position > len(me.Input) {
		return ""
	}

	old := me.Position

	for ; me.Position < len(me.Input); me.Position++ {
		if !strings.Contains(spaceChars, me.Input[me.Position:me.Position+1]) {
			break
		}
	}

	return me.Input[old:me.Position]
}

func (me *attributeTokenizer) shiftValue() string {
	value := me.shiftUntilSpace()
	quoteChars := `"'`

	if value == "" || !strings.ContainsAny(value[0:1], quoteChars) {
		return value
	}

	q := value[0:1]
	for value[len(value)-1:] != q {
		part := me.eatSpace() + me.shiftUntilSpace()
		if part == "" {
			break
		}
		value += part
	}

	return strings.Trim(value, q)
}

func (me *attributeTokenizer) Next() (Attribute, error) {
	me.eatSpace()

	if me.Position >= len(me.Input) {
		return Attribute{}, io.EOF
	}

	key := me.shiftUntil("=")
	me.Position++
	me.eatSpace()

	if me.Position >= len(me.Input) {
		return Attribute{key, ""}, nil
	}

	return Attribute{key, me.shiftValue()}, nil
}

func getAttribute(rawInput, name string) (string, bool) {
	name = strings.ToLower(name)

	z := &attributeTokenizer{Input: rawInput}
	// eat the element name
	z.shiftUntilSpace()

	for {
		a, err := z.Next()
		if err != nil {
			break
		}
		if strings.ToLower(a.Name) == name {
			return a.Content, true
		}
	}

	return "", false
}

func getAttributes(rawInput string) []Attribute {
	list := []Attribute{}

	z := &attributeTokenizer{Input: rawInput}
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
