package extension

import (
	"io"
	"regexp"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// TaskStatus represents the status of a task in a task list.
type TaskStatus string

const (
	// TaskStatusActive indicates that the task is active (not completed).
	TaskStatusActive TaskStatus = "active"

	// TaskStatusCompleted indicates that the task is completed.
	TaskStatusCompleted TaskStatus = "completed"
)

const taskStatusAttributeName = "task-status"

// IsTask returns true if the given node is a task list item.
func IsTask(node gast.Node) bool {
	listItem, ok := node.(*gast.ListItem)
	if !ok {
		return false
	}
	_, ok = listItem.Attribute(taskStatusAttributeName)
	return ok
}

// TaskStatusOf returns the TaskStatus of the given node if it is a task list item, and
// a boolean indicating whether the status was found.
func TaskStatusOf(node gast.Node) (TaskStatus, bool) {
	listItem, ok := node.(*gast.ListItem)
	if !ok {
		return "", false
	}
	attr, ok := listItem.Attribute(taskStatusAttributeName)
	if !ok {
		return "", false
	}
	v := attr.Str(nil) // value is always an owned string, so this is safe
	return TaskStatus(v), true
}

var taskCheckboxRegexp = regexp.MustCompile(`^\[([\sxX])\]\s*`)

type taskListItemParser struct {
}

var defaultTaskListItemParser = &taskListItemParser{}

// newTaskListItemParser returns a new InlineParser that can parse
// checkboxes in list items.
// This parser must take precedence over the parser.LinkParser.
func newTaskListItemParser() parser.InlineParser {
	return defaultTaskListItemParser
}

func (s *taskListItemParser) Trigger() []byte {
	return []byte{'['}
}

func (s *taskListItemParser) Parse(parent gast.Node, block text.Reader, _ parser.Context) gast.Node {
	// Given AST structure must be like
	// - List
	//   - ListItem         : parent.Parent
	//     - Paragraph      : parent
	//       (current line)
	if parent.Parent() == nil || parent.Parent().FirstChild() != parent {
		return nil
	}

	if parent.HasChildren() {
		return nil
	}

	listItem, ok := parent.Parent().(*gast.ListItem)
	if !ok {
		return nil
	}
	if _, alreadySet := listItem.Attribute(taskStatusAttributeName); alreadySet {
		return nil
	}
	line, _ := block.PeekLine()
	m := taskCheckboxRegexp.FindSubmatchIndex(line)
	if m == nil {
		return nil
	}
	value := line[m[2]:m[3]][0]
	block.Advance(m[1])
	checked := value == 'x' || value == 'X'
	if checked {
		listItem.SetAttribute(taskStatusAttributeName,
			text.NewStringMultilineValue(string(TaskStatusCompleted)))
	} else {
		listItem.SetAttribute(taskStatusAttributeName,
			text.NewStringMultilineValue(string(TaskStatusActive)))
	}
	return parser.Nil
}

func (s *taskListItemParser) CloseBlock(_ gast.Node, _ parser.Context) {
	// nothing to do
}

type taskListItemHTMLRendererConfig struct {
	XHTML              bool
	IsInTightBlockFunc html.IsInTightBlockFunc
}

// TaskListItemHTMLRendererOption represents an option for configuring the task list item HTML renderer.
type TaskListItemHTMLRendererOption interface {
	applyTaskListItemHTMLRendererOption(*taskListItemHTMLRendererConfig)
}

type taskListItemHTMLRendererExtension struct {
	config taskListItemHTMLRendererConfig
}

// NewTaskListItemHTMLRenderer returns a new html.Extension for rendering task list items.
func NewTaskListItemHTMLRenderer(opts ...TaskListItemHTMLRendererOption) html.Extension {
	config := taskListItemHTMLRendererConfig{}
	for _, opt := range opts {
		opt.applyTaskListItemHTMLRendererOption(&config)
	}
	return &taskListItemHTMLRendererExtension{
		config: config,
	}
}

func (r *taskListItemHTMLRendererExtension) RendererOptions(c *html.Config) []html.Option {
	if c.XHTML {
		r.config.XHTML = true
	}
	if c.Paragraph.IsInTightBlockFunc != nil {
		r.config.IsInTightBlockFunc = c.Paragraph.IsInTightBlockFunc
	}
	return []html.Option{
		html.WithNodeRenderers(map[gast.NodeKind]html.NodeRenderer{
			gast.KindParagraph: html.NodeRendererFunc(r.renderParagraph),
		}),
	}
}

func (r *taskListItemHTMLRendererExtension) renderInputTag(w util.BufWriter, status TaskStatus) {
	if status == TaskStatusCompleted {
		_, _ = w.WriteString(`<input checked="" disabled="" type="checkbox"`)
	} else {
		_, _ = w.WriteString(`<input disabled="" type="checkbox"`)
	}
	if r.config.XHTML {
		_, _ = w.WriteString(" /> ")
	} else {
		_, _ = w.WriteString("> ")
	}
}

func (r *taskListItemHTMLRendererExtension) renderParagraph(
	writer io.Writer, _ []byte, node gast.Node, entering bool, _ renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*gast.Paragraph)

	// Determine whether this paragraph belongs to a task list item.
	status, isTask := func() (TaskStatus, bool) {
		parent := n.Parent()
		if parent == nil || parent.FirstChild() != n {
			return "", false
		}
		return TaskStatusOf(parent)
	}()

	inTight := r.config.IsInTightBlockFunc(n)

	if !isTask {
		// Standard paragraph rendering.
		if inTight {
			if !entering && n.NextSibling() != nil && n.FirstChild() != nil {
				_ = w.WriteByte('\n')
			}
			return gast.WalkContinue, nil
		}
		if entering {
			if n.Attributes() != nil {
				_, _ = w.WriteString("<p")
				html.RenderAttributes(w, n, html.ParagraphAttributeFilter)
				_ = w.WriteByte('>')
			} else {
				_, _ = w.WriteString("<p>")
			}
		} else {
			_, _ = w.WriteString("</p>\n")
		}
		return gast.WalkContinue, nil
	}

	// Task paragraph rendering.
	if inTight {
		if entering {
			r.renderInputTag(w, status)
		} else if n.NextSibling() != nil && n.FirstChild() != nil {
			_ = w.WriteByte('\n')
		}
		return gast.WalkContinue, nil
	}
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<p")
			html.RenderAttributes(w, n, html.ParagraphAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<p>")
		}
		r.renderInputTag(w, status)
	} else {
		_, _ = w.WriteString("</p>\n")
	}
	return gast.WalkContinue, nil
}

type taskCheckBoxParserExtension struct {
}

// NewTaskCheckBoxParser returns a new parser.Extension for parsing task checkboxes in list items.
func NewTaskCheckBoxParser() parser.Extension {
	return &taskCheckBoxParserExtension{}
}

func (e *taskCheckBoxParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(
			util.Prioritized(newTaskListItemParser(), 0),
		),
	}
}
