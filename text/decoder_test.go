package text_test

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark/v2/text"
)

func TestDecoder(t *testing.T) {
	s := "foo\\:\\) &amp;ab&#x3A; &#58;"
	d := text.NewDecoder()
	if !bytes.Equal(d.Decode([]byte(s)), []byte("foo:) &ab: :")) {
		t.Errorf("Decode failed: %s, should be 'foo:) &ab: :', actual: %s", s, d.Decode([]byte(s)))
	}
}
