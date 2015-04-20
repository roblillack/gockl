package gockl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"reflect"
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
    <linearGradient id="SvgjsLinearGradient1070">
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
