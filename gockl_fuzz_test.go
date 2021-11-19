//go:build go1.18
// +build go1.18

package gockl

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var seed = []string{
	"<doc></doc>",
	`<?xml version="1.0" encoding="UTF-8"?>
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
}

func uniqueAttributes(tok StartOrEmptyElementToken) []Attribute {
	seen := map[string]struct{}{}
	r := []Attribute{}

	for _, i := range tok.Attributes() {
		low := strings.ToLower(i.Name)
		if _, ok := seen[low]; ok {
			continue
		}
		r = append(r, i)
		seen[low] = struct{}{}
	}

	return r
}

func process(str string) error {
	z := New(str)

	for {
		t, err := z.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if tok, ok := t.(ElementToken); ok {
			_ = tok.Name()
		}

		if tok, ok := t.(StartOrEmptyElementToken); ok {
			for _, i := range uniqueAttributes(tok) {
				c, ok := tok.Attribute(i.Name)
				if !ok {
					return fmt.Errorf("unable to find previously found attribute %s", i.Name)
				}
				if c != i.Content {
					return fmt.Errorf("attribute content does not match for %s: %s != %s", i.Name, i.Content, c)
				}
			}
		}
	}
}

func decode(data string) ([]Token, error) {
	r := []Token{}
	z := New(data)

	for {
		t, err := z.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return r, nil
		}
		r = append(r, t)
	}
}

func encode(tokens []Token) (string, error) {
	b := strings.Builder{}
	for _, t := range tokens {
		if _, err := b.WriteString(t.Raw()); err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

func FuzzParsing(f *testing.F) {
	for _, i := range seed {
		f.Add(i)
	}

	files, err := ioutil.ReadDir(filepath.Join("testdata", "xml"))
	if err != nil {
		return
	}

	for _, fi := range files {
		raw, err := ioutil.ReadFile(filepath.Join("testdata", "xml", fi.Name()))
		if err != nil {
			f.Logf("Error reading %s: %s", fi.Name(), err)
			continue
		}

		f.Add(string(raw))
	}

	f.Fuzz(func(t *testing.T, documentContent string) {
		if err := process(documentContent); err != nil {
			t.Fatalf("Unable to decode doc %s: %v", documentContent, err)
		}
	})
}

func FuzzRoundtrip(f *testing.F) {
	for _, i := range seed {
		f.Add(i)
	}
	f.Fuzz(func(t *testing.T, document1 string) {
		tokens1, err := decode(document1)
		if err != nil {
			t.Fatalf("Unable to decode doc %s: %v", document1, err)
		}

		document2, err := encode(tokens1)
		if err != nil {
			t.Fatalf("Unable to encode document %s we decoded before: %v", document1, err)
		}

		if !reflect.DeepEqual(document1, document2) {
			t.Errorf("%s != %s", document1, document2)
		}
	})
}
