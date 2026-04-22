package extension

import "github.com/yuin/goldmark/v2/renderer/html"

type withXHTML struct {
	value bool
}

func (o *withXHTML) applyTableHTMLRendererOption(c *tableHTMLRendererConfig) {
	c.XHTML = o.value
}

func (o *withXHTML) applyTaskListItemHTMLRendererOption(c *taskListItemHTMLRendererConfig) {
	c.XHTML = o.value
}

func (o *withXHTML) applyFootnoteHTMLRendererOption(c *footnoteHTMLRendererConfig) {
	c.XHTML = true
}

// WithXHTML is a functional option that indicates whether the table should be rendered as XHTML.
func WithXHTML() interface {
	TableHTMLRendererOption
	TaskListItemHTMLRendererOption
	FootnoteHTMLRendererOption
} {
	return &withXHTML{value: true}
}

type withIsInTightBlockFunc struct {
	IsInTightBlockFunc html.IsInTightBlockFunc
}

func (o *withIsInTightBlockFunc) applyTaskListItemHTMLRendererOption(c *taskListItemHTMLRendererConfig) {
	c.IsInTightBlockFunc = o.IsInTightBlockFunc
}

// WithIsInTightBlockFunc is a functional option that sets the function to determine
// whether the list item is in a tight block.
func WithIsInTightBlockFunc(f html.IsInTightBlockFunc) interface {
	TaskListItemHTMLRendererOption
} {
	return &withIsInTightBlockFunc{IsInTightBlockFunc: f}
}
