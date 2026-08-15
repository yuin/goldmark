package text_test

import (
	"regexp"
	"testing"

	"github.com/yuin/goldmark/v2/text"
)

func TestFindSubMatchReader(t *testing.T) {
	s := "微笑"
	r := text.NewReader([]byte(":"+s+":"), text.NewDecoder())
	reg := regexp.MustCompile(`:(\p{L}+):`)
	match := r.FindSubMatch(reg)
	if len(match) != 2 || string(match[1]) != s {
		t.Fatal("no match cjk")
	}
}
