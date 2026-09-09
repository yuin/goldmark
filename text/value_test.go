package text_test

import (
	"testing"

	"github.com/yuin/goldmark/v2/text"
)

func TestSegmentTrimLeftSpaceWidth(t *testing.T) {
	source := []byte("\t  foo")
	for _, c := range []struct {
		name       string
		segment    text.Segment
		width      int
		currentPos int
		expected   string
	}{
		// A tab at the third column is only two columns wide, so trimming four
		// columns eats the two spaces that follow it as well.
		{"tab at the start of a line", text.NewSegment(0, len(source)), 4, 0, "  foo"},
		{"tab at the third column", text.NewSegment(0, len(source)), 4, 2, "foo"},
		{"tab after two columns of padding", text.NewSegmentPadding(0, len(source), 2), 4, 0, "  foo"},
		{"spaces", text.NewSegment(1, len(source)), 1, 0, " foo"},
	} {
		got := string(c.segment.TrimLeftSpaceWidthAt(c.width, c.currentPos, source).Bytes(source))
		if got != c.expected {
			t.Errorf("%s: TrimLeftSpaceWidthAt(%d, %d): expected %q, got %q", c.name, c.width, c.currentPos, c.expected, got)
		}
		if c.currentPos != 0 {
			continue
		}
		got = string(c.segment.TrimLeftSpaceWidth(c.width, source).Bytes(source))
		if got != c.expected {
			t.Errorf("%s: TrimLeftSpaceWidth(%d): expected %q, got %q", c.name, c.width, c.expected, got)
		}
	}
}
