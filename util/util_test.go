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
