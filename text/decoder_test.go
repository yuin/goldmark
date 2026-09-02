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

// TestDecoderCursorInvalidation exercises Decode/DecodeTo's jump-based '&'/'\'
// cursor tracking, where consuming one match can invalidate (and require
// recomputing) the cached position of the other, or leave it untouched.
func TestDecoderCursorInvalidation(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"backslash escapes ampersand, entity not reprocessed", `\&amp;`, "&amp;"},
		{"failed entity immediately followed by real entity", "&&amp;", "&&"},
		{"backslash-escaped backslash leaves following entity intact", `\\&amp;`, `\&`},
		{"trailing lone backslash", `abc\`, `abc\`},
		{"trailing bare ampersand", "abc&", "abc&"},
		{"hex entity followed immediately by escape", `&#x26;\*`, "&*"},
		{"consecutive entities", "&amp;&amp;&amp;", "&&&"},
		{"ampersand immediately followed by backslash escape", `&*\*`, "&**"},
		{"decimal entity then bare ampersand", "&#38;&", "&&"},
		{"named entity with no trailing semicolon", "&amp", "&amp"},
	}

	d := text.NewDecoder()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := d.Decode([]byte(c.in))
			if !bytes.Equal(got, []byte(c.want)) {
				t.Errorf("Decode(%q) = %q, want %q", c.in, got, c.want)
			}
			var buf bytes.Buffer
			if _, err := d.DecodeTo(&buf, []byte(c.in)); err != nil {
				t.Fatalf("DecodeTo(%q) returned error: %v", c.in, err)
			}
			if buf.String() != c.want {
				t.Errorf("DecodeTo(%q) = %q, want %q", c.in, buf.String(), c.want)
			}
		})
	}
}

func TestDecoderEscapedSpace(t *testing.T) {
	d := text.NewDecoder(text.WithEscapedSpace())
	in := "a\\ b\\&c"
	want := "ab&c"
	got := d.Decode([]byte(in))
	if !bytes.Equal(got, []byte(want)) {
		t.Errorf("Decode(%q) = %q, want %q", in, got, want)
	}
	var buf bytes.Buffer
	if _, err := d.DecodeTo(&buf, []byte(in)); err != nil {
		t.Fatalf("DecodeTo(%q) returned error: %v", in, err)
	}
	if buf.String() != want {
		t.Errorf("DecodeTo(%q) = %q, want %q", in, buf.String(), want)
	}
}
