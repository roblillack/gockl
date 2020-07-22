package gockl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"reflect"
	"strings"
	"testing"
)

type DocumentInfo struct {
	Data         string
	ElementNames []string
}

var documents map[string]DocumentInfo = map[string]DocumentInfo{
	// taken from https://github.com/golang/go/issues/10158
	"doctype subset": DocumentInfo{
		Data: `<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE doc [
    <!ELEMENT doc ANY>
]>
<doc>
</doc>`,
		ElementNames: []string{"doc", "doc"},
	},
	"simple-svg": DocumentInfo{
		Data: `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" version="1.1" width="100%" height="100%" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 1920 1080">
  <style>
/* This is a comment. */
.test {
	fill: 'black';
}
  </style>
  <rect width="1920" height="1080" class="test" fill="red"></rect>
  <defs>
    <linearGradient id="grad">
      <stop stop-color="white" offset="0"></stop>
      <stop stop-opacity="0" stop-color="white" offset="1"></stop>
    </linearGradient>
  </defs>
</svg>`,
		ElementNames: []string{
			"svg", "style", "style", "rect", "rect", "defs", "linearGradient", "stop", "stop", "stop", "stop", "linearGradient", "defs", "svg",
		},
	},
}

func passthrough(data string) string {
	buf := bytes.Buffer{}
	z := New(data)

	for {
		t, err := z.Next()
		if err != nil {
			break
		}

		buf.WriteString(t.Raw())
	}

	return buf.String()
}

func elements(data string) []string {
	r := []string{}
	z := New(data)

	for {
		t, err := z.Next()
		if err != nil {
			break
		}

		if el, ok := t.(ElementToken); ok {
			r = append(r, el.Name())
		}
	}

	return r
}

func Test_NoChange(t *testing.T) {
	for name, info := range documents {
		if info.Data != passthrough(info.Data) {
			t.Errorf("Error processing document '%s'", name)
		}
	}
}

func Test_ElementNames(t *testing.T) {
	for name, info := range documents {
		elements := elements(info.Data)
		for pos, expected := range info.ElementNames {
			if pos >= len(elements) {
				t.Errorf("Element pos %d not existing for document %s", pos, name)
			} else if actual := elements[pos]; actual != expected {
				t.Errorf("Element name not matching at pos %d for document %s: %s (actual) != %s (expected)", pos, name, actual, expected)
			}
		}
	}
}

func Test_BrokenStartElement(t *testing.T) {
	input := "<elem"
	decoder := New(input)
	tok, err := decoder.Next()
	if err != nil {
		t.Error("Error while getting token.")
	}
	if _, ok := tok.(StartElementToken); !ok {
		t.Errorf("Not a start element token: %s, %s", input, reflect.TypeOf(tok))
	}
	if tok.Raw() != input {
		t.Errorf("Token text not matching: %s (expected) != %s (actual)", input, tok.Raw())
	}
	if next, err := decoder.Next(); next != nil || err != io.EOF {
		t.Errorf("Wanted EOF, got: '%s'/%s", next, err)
	}
}

func Test_BrokenTextElement(t *testing.T) {
	input := "/asdkjlh"
	decoder := New(input)
	tok, err := decoder.Next()
	if err != nil {
		t.Error("Error while getting token.")
	}
	if _, ok := tok.(TextToken); !ok {
		t.Errorf("Not a text token: %s, %s", input, reflect.TypeOf(tok))
	}
	if tok.Raw() != input {
		t.Errorf("Token text not matching: %s (expected) != %s (actual)", input, tok.Raw())
	}
	if next, err := decoder.Next(); next != nil || err != io.EOF {
		t.Errorf("Wanted EOF, got: '%s'/%s", next, err)
	}
}

// I have a directory of ~15.000 XML files created using: find ~ -name '*.xml' -exec cp {} ./XML \;
func Test_RealLifeFiles(t *testing.T) {
	dir := "/Users/rob/dev/gockl/XML"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, fi := range files {
		raw, err := ioutil.ReadFile(path.Join(dir, fi.Name()))
		if err != nil {
			t.Logf("Error reading %s: %s", fi.Name(), err)
			continue
		}

		input := string(raw)
		if output := passthrough(input); output != input {
			outfile := fmt.Sprintf("%s.tmp", path.Join(dir, fi.Name()))
			t.Errorf("Error processing document '%s', writing output to %s for comparing to ", fi.Name(), outfile)
			ioutil.WriteFile(outfile, []byte(output), 0644)
			return
		}

		//t.Logf("%d/%d files checked.", i+1, len(files))
	}
}

func Test_Attributes(t *testing.T) {
	svg := StartElementToken(`<svg xmlns="http://www.w3.org/2000/svg" version=1.1 width='100%' height='a + b' xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 1920 1080" bla=blub bla>`)
	attrs := []Attribute{
		Attribute{"xmlns", "http://www.w3.org/2000/svg"},
		Attribute{"version", "1.1"},
		Attribute{"width", "100%"},
		Attribute{"height", "a + b"},
		Attribute{"xmlns:xlink", "http://www.w3.org/1999/xlink"},
		Attribute{"viewBox", "0 0 1920 1080"},
		Attribute{"bla", "blub"},
		Attribute{"bla", ""},
	}

	result := svg.Attributes()
	if len(result) != len(attrs) {
		t.Error("Attribute count not matching")
		t.Errorf("Expected: %v", attrs)
		t.Errorf("Got: %v", result)
	}

	for i, a := range result {
		if i >= len(attrs) {
			t.Errorf("Unexpected attribute: %s", a)
		} else if attrs[i] != a {
			t.Errorf("Attributes not matching. Expected %s, got %s", attrs[i], a)
		}
	}
}

func Test_AttributesInEmptyElements(t *testing.T) {
	svg := EmptyElementToken(`<circle cx="50" cy="25" r="20" fill="yellow" />`)
	attrs := []Attribute{
		Attribute{"cx", "50"},
		Attribute{"cy", "25"},
		Attribute{"r", "20"},
		Attribute{"fill", "yellow"},
	}

	if "circle" != svg.Name() {
		t.Error("No a circle")
	}

	result := svg.Attributes()
	if len(result) != len(attrs) {
		t.Error("Attribute count not matching")
		t.Errorf("Expected: %v", attrs)
		t.Errorf("Got: %v", result)
	}

	for i, a := range result {
		if i >= len(attrs) {
			t.Errorf("Unexpected attribute: %s", a)
		} else if attrs[i] != a {
			t.Errorf("Attributes not matching. Expected %s, got %s", attrs[i], a)
		}
	}
}

func TestGettingAttributesByName(t *testing.T) {
	type AttribTest struct {
		Token         Token
		Attributes    []Attribute
		NonAttributes []string
	}
	testdata := []AttribTest{
		AttribTest{
			EmptyElementToken(`<circle cx="50" cy="25" r="20" fill="yellow" />`),
			[]Attribute{
				Attribute{Name: "cx", Content: "50"},
				Attribute{Name: "R", Content: "20"},
			},
			[]string{"bla"},
		},
		AttribTest{
			StartElementToken(`<group style="fill: none;" style="nope">`),
			[]Attribute{
				Attribute{Name: "style", Content: "fill: none;"},
			},
			[]string{"r"},
		},
	}

	for _, i := range testdata {
		tok, ok := i.Token.(StartOrEmptyElementToken)
		if !ok {
			t.FailNow()
		}

		for _, a := range i.Attributes {
			for _, fn := range []func(string) string{
				strings.ToLower,
				strings.ToUpper,
				func(x string) string { return x },
			} {
				name := fn(a.Name)
				res, ok := tok.Attribute(name)
				if !ok {
					t.Errorf("Missing attribute %s in token %s", name, i.Token)
					continue
				}
				if res != a.Content {
					t.Errorf("Wrong attribute %s in token %s: Expected %s, got %s", name, i.Token, a.Content, res)
				}
			}
		}
	}
}

func Test_CDATA(t *testing.T) {
	doc := New(`<p><![CDATA[</p>]]><!-- </p> --></p>`)

	if tok, err := doc.Next(); err != nil {
		t.Error(err)
	} else if start, ok := tok.(StartElementToken); !ok {
		t.Errorf("Expected start element token, got: %s", start)
	}

	if tok, err := doc.Next(); err != nil {
		t.Error(err)
	} else if cdata, ok := tok.(CDATAToken); !ok {
		t.Errorf("Expected CDATA token, got: %s", cdata)
	} else if raw := cdata.Raw(); raw != `<![CDATA[</p>]]>` {
		t.Errorf("Wrong content for CDATA token: %s", raw)
	}

	if tok, err := doc.Next(); err != nil {
		t.Error(err)
	} else if comment, ok := tok.(CommentToken); !ok {
		t.Errorf("Expected end element token, got: %s", comment)
	} else if raw := comment.Raw(); raw != `<!-- </p> -->` {
		t.Errorf("Wrong content for end token: %s", raw)
	}

	if tok, err := doc.Next(); err != nil {
		t.Error(err)
	} else if end, ok := tok.(EndElementToken); !ok {
		t.Errorf("Expected end element token, got: %s", end)
	} else if raw := end.Raw(); raw != `</p>` {
		t.Errorf("Wrong content for end token: %s", raw)
	}
}
