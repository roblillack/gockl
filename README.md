gockl [![Build Status](https://secure.travis-ci.org/roblillack/gockl.png?branch=master)](http://travis-ci.org/roblillack/gockl)
[![GoDoc](http://godoc.org/github.com/roblillack/gockl?status.png)](http://godoc.org/github.com/roblillack/gockl)
[![Coverage Status](https://coveralls.io/repos/github/roblillack/gockl/badge.svg)](https://coveralls.io/github/roblillack/gockl)
[![Go Report Card](https://goreportcard.com/badge/github.com/roblillack/gockl)](https://goreportcard.com/report/github.com/roblillack/gockl)
=======

gockl is a minimal XML processor for Go that does not to fuck with your markup.

Supported & tested Go versions are: 1.2 – 1.18.

#### Usage

Transparently decode XML string `input` and re-encode to string `output` without affecting
the underlying structure of the original file:

```go
buf := bytes.Buffer{}
z := gockl.New(input)

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
```

#### Why?

- To ease creating XML document diffs, if only minor changes to a document are done
- To not run into over-escaping of text data in `encoding/xml`: https://github.com/golang/go/issues/9204
- To not run into broken namespace handling: https://github.com/golang/go/issues/9519
- To not run into errors when parsing DOCTYPEs with subsets: https://github.com/golang/go/issues/10158

#### License

[MIT/X11](https://github.com/roblillack/gockl/blob/master/LICENSE.txt).
