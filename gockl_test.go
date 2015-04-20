package gockl

import (
	"bytes"
	"testing"
)

var documents map[string]string = map[string]string{
	"simple-svg": `<?xml version="1.0" encoding="UTF-8"?>
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

func Test_NoChange(t *testing.T) {
	for name, raw := range documents {
		if raw != passthrough(raw) {
			t.Errorf("Error processing document '%s'", name)
		}
	}
}
