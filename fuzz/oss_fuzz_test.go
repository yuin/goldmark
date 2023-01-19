package fuzz

import (
	"testing"
)

func FuzzOss(f *testing.F) {
	fuzz(f)
}
