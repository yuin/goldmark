package goldmark_test

import (
	"testing"

	"github.com/yuin/goldmark"
)

func TestConvertNilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked: %v", r)
		}
	}()
	err := goldmark.Convert([]byte("# hi"), nil)
	if err == nil {
		t.Fatal("want error")
	}
}
