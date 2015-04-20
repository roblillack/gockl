gockl [![Build Status](https://secure.travis-ci.org/roblillack/gockl.png?branch=master)](http://travis-ci.org/roblillack/gockl) [![GoDoc](http://godoc.org/github.com/roblillack/gockl?status.png)](http://godoc.org/github.com/roblillack/gockl)
=======

gockl is a minimal XML processor for Go that does not to fuck with your markup.

#### Usage ####

Transparently “convert” string `input` to `output` and log names for all element
tokens:

	buf := bytes.Buffer{}
	z := New(input)

	for {
		t, err := z.Next()
		if err != nil {
			break
		}

		if el, ok := t.(gockl.ElementToken); ok {
			log.Println(el.Name())
		}
		buf.WriteString(t.Raw())
	}

	output := buf.String()

#### License ####

[MIT/X11](https://github.com/roblillack/gockl/blob/master/LICENSE.txt).