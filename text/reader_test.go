package text

import (
	"regexp"
	"testing"
)

func TestFindSubMatchReader(t *testing.T) {
	s := "微笑"
	r := NewReader([]byte(":" + s + ":"))
	reg := regexp.MustCompile(`:(\p{L}+):`)
	match := r.FindSubMatch(reg)
	if len(match) != 2 || string(match[1]) != s {
		t.Fatal("no match cjk")
	}
}
