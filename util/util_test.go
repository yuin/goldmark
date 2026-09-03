package util

import (
	"testing"
)

func chr(c int) string { return string([]byte{byte(c)}) }

func TestURLEscape(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// %xx is kept as is.
		{"/a%41b", "/a%41b"},
		{"/a%c3%a9b", "/a%c3%a9b"},
		// A % that does not start a complete triple is escaped.
		{"/a%zzb", "/a%25zzb"},
		{"/a%4", "/a%254"},
		// The byte after %<hex> must be escaped like any other byte.
		{"/x%A<y", "/x%25A%3Cy"},
		{"/x%A y", "/x%25A%20y"},
		{"/x%Aéy", "/x%25A%C3%A9y"},
		// Malformed utf8 is passed through, not dropped or query escaped.
		{"/x\xffy", "/x\xffy"},
		{"a b\xc3", "a%20b\xc3"},
		{"/x\xe2\x82 y", "/x\xe2\x82%20y"},
		{"/x\ufffdy", "/x%EF%BF%BDy"},
	}
	for _, tt := range tests {
		if got := string(URLEscape([]byte(tt.in), false)); got != tt.want {
			t.Errorf("URLEscape(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// How a byte is escaped must not depend on it following a %<hex> pair.
func TestURLEscapePercentDoesNotSkipNextByte(t *testing.T) {
	prefix := string(URLEscape([]byte("/x%A"), false))
	for c := 0; c < 256; c++ {
		if IsHexDecimal(byte(c)) {
			continue // %A followed by a hexdecimal is a complete triple
		}
		elsewhere := string(URLEscape([]byte("/xA"+chr(c)+"Z"), false))
		in := "/x%A" + chr(c) + "Z"
		want := prefix + elsewhere[len("/xA"):]
		if got := string(URLEscape([]byte(in), false)); got != want {
			t.Errorf("URLEscape(%q) = %q, want %q (from URLEscape(%q) = %q)",
				in, got, want, "/xA"+chr(c)+"Z", elsewhere)
		}
	}
}

// A truncated utf8 sequence must be passed through like an invalid leading
// byte is. "a b" makes the copy on write buffer copy, which exposes the drop.
func TestURLEscapeTruncatedUTF8(t *testing.T) {
	for c := 0xc2; c <= 0xf4; c++ {
		in := "a b" + chr(c)
		want := "a%20b" + chr(c)
		if got := string(URLEscape([]byte(in), false)); got != want {
			t.Errorf("URLEscape(%q) = %q, want %q", in, got, want)
		}
	}
}
