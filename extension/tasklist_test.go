package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"testing"
)

func TestTaskList(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			TaskList,
		),
	)
	goldmark.DoTestCaseFile(markdown, "_test/tasklist.txt", t)
}
