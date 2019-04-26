package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

type gfm struct {
}

// GFM is an extension that provides Github Flavored markdown functionalities.
var GFM = &gfm{}

var filterTags = []string{
	"title",
	"textarea",
	"style",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"script",
	"plaintext",
}

func (e *gfm) Extend(m goldmark.Markdown) {
	m.Parser().AddOption(parser.WithFilterTags(filterTags...))
	Linkify.Extend(m)
	Table.Extend(m)
	Strikethrough.Extend(m)
	TaskList.Extend(m)
}
