package parser

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var allowedBlockTags = map[string]bool{
	"address":    true,
	"article":    true,
	"aside":      true,
	"base":       true,
	"basefont":   true,
	"blockquote": true,
	"body":       true,
	"caption":    true,
	"center":     true,
	"col":        true,
	"colgroup":   true,
	"dd":         true,
	"details":    true,
	"dialog":     true,
	"dir":        true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"fieldset":   true,
	"figcaption": true,
	"figure":     true,
	"footer":     true,
	"form":       true,
	"frame":      true,
	"frameset":   true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"head":       true,
	"header":     true,
	"hr":         true,
	"html":       true,
	"iframe":     true,
	"legend":     true,
	"li":         true,
	"link":       true,
	"main":       true,
	"menu":       true,
	"menuitem":   true,
	"meta":       true,
	"nav":        true,
	"noframes":   true,
	"ol":         true,
	"optgroup":   true,
	"option":     true,
	"p":          true,
	"param":      true,
	"search":     true,
	"section":    true,
	"summary":    true,
	"table":      true,
	"tbody":      true,
	"td":         true,
	"tfoot":      true,
	"th":         true,
	"thead":      true,
	"title":      true,
	"tr":         true,
	"track":      true,
	"ul":         true,
}

var htmlBlockType1OpenRegexp = regexp.MustCompile(`(?i)^[ ]{0,3}<(script|pre|style|textarea)(?:\s.*|>.*|/>.*|)(?:\r\n|\n)?$`) //nolint:lll
var htmlBlockType1CloseRegexp = regexp.MustCompile(`(?i)^.*</(?:script|pre|style|textarea)>.*`)

var htmlBlockType2OpenRegexp = regexp.MustCompile(`^[ ]{0,3}<!\-\-`)
var htmlBlockType2Close = []byte{'-', '-', '>'}

var htmlBlockType3OpenRegexp = regexp.MustCompile(`^[ ]{0,3}<\?`)
var htmlBlockType3Close = []byte{'?', '>'}

var htmlBlockType4OpenRegexp = regexp.MustCompile(`^[ ]{0,3}<![A-Z]+.*(?:\r\n|\n)?$`)
var htmlBlockType4Close = []byte{'>'}

var htmlBlockType5OpenRegexp = regexp.MustCompile(`^[ ]{0,3}<\!\[CDATA\[`)
var htmlBlockType5Close = []byte{']', ']', '>'}

var htmlBlockType6Regexp = regexp.MustCompile(`^[ ]{0,3}<(?:/[ ]*)?([a-zA-Z]+[a-zA-Z0-9\-]*)(?:[ ].*|>.*|/>.*|)(?:\r\n|\n)?$`) //nolint:lll

var htmlBlockType7Regexp = regexp.MustCompile(`^[ ]{0,3}<(/[ ]*)?([a-zA-Z]+[a-zA-Z0-9\-]*)(` + attributePattern + `*)[ ]*(?:>|/>)[ ]*(?:\r\n|\n)?$`) //nolint:lll

type htmlBlockParser struct {
}

var defaultHTMLBlockParser = &htmlBlockParser{}

// NewHTMLBlockParser return a new BlockParser that can parse html
// blocks.
func NewHTMLBlockParser() BlockParser {
	return defaultHTMLBlockParser
}

func (b *htmlBlockParser) Trigger() []byte {
	return []byte{'<'}
}

func (b *htmlBlockParser) Open(_ ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	var node *ast.HTMLBlock
	line, segment := reader.PeekLine()
	last := pc.LastOpenedBlock().Node

	if m := htmlBlockType1OpenRegexp.FindSubmatchIndex(line); m != nil {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind1)
	} else if htmlBlockType2OpenRegexp.Match(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind2)
	} else if htmlBlockType3OpenRegexp.Match(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind3)
	} else if htmlBlockType4OpenRegexp.Match(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind4)
	} else if htmlBlockType5OpenRegexp.Match(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind5)
	} else if match := htmlBlockType7Regexp.FindSubmatchIndex(line); match != nil {
		isCloseTag := match[2] > -1 && bytes.Equal(line[match[2]:match[3]], []byte("/"))
		hasAttr := match[6] != match[7]
		tagName := strings.ToLower(string(line[match[4]:match[5]]))
		_, ok := allowedBlockTags[tagName]
		if ok {
			node = ast.NewHTMLBlock(ast.HTMLBlockKind6)
		} else if tagName != "script" && tagName != "style" &&
			tagName != "pre" && !ast.IsParagraph(last) && (!isCloseTag || !hasAttr) { // type 7 can not interrupt paragraph
			node = ast.NewHTMLBlock(ast.HTMLBlockKind7)
		}
	}
	if node == nil {
		if match := htmlBlockType6Regexp.FindSubmatchIndex(line); match != nil {
			tagName := string(line[match[2]:match[3]])
			_, ok := allowedBlockTags[strings.ToLower(tagName)]
			if ok {
				node = ast.NewHTMLBlock(ast.HTMLBlockKind6)
			}
		}
	}
	if node != nil {
		reader.AdvanceToEOL()
		node.Value.AppendSegment(segment)
		return node, NoChildren
	}
	return nil, NoChildren
}

func (b *htmlBlockParser) Continue(node ast.Node, reader text.Reader, _ Context) State {
	htmlBlock := node.(*ast.HTMLBlock)
	line, segment := reader.PeekLine()
	var closurePattern []byte

	switch htmlBlock.HTMLBlockKind {
	case ast.HTMLBlockKind1:
		if len(htmlBlock.Value.Segments()) == 1 {
			firstLine := htmlBlock.Value.Segments()[0]
			if htmlBlockType1CloseRegexp.Match(firstLine.Bytes(reader.Source())) {
				return Close
			}
		}
		if htmlBlockType1CloseRegexp.Match(line) {
			htmlBlock.Value.AppendSegment(segment)
			reader.AdvanceToEOL()
			return Close
		}
	case ast.HTMLBlockKind2:
		closurePattern = htmlBlockType2Close
		fallthrough
	case ast.HTMLBlockKind3:
		if closurePattern == nil {
			closurePattern = htmlBlockType3Close
		}
		fallthrough
	case ast.HTMLBlockKind4:
		if closurePattern == nil {
			closurePattern = htmlBlockType4Close
		}
		fallthrough
	case ast.HTMLBlockKind5:
		if closurePattern == nil {
			closurePattern = htmlBlockType5Close
		}

		if len(htmlBlock.Value.Segments()) == 1 {
			firstLine := htmlBlock.Value.Segments()[0]
			if bytes.Contains(firstLine.Bytes(reader.Source()), closurePattern) {
				return Close
			}
		}
		if bytes.Contains(line, closurePattern) {
			htmlBlock.Value.AppendSegment(segment)
			reader.AdvanceToEOL()
			return Close
		}

	case ast.HTMLBlockKind6, ast.HTMLBlockKind7:
		if util.IsBlank(line) {
			return Close
		}
	}
	htmlBlock.Value.AppendSegment(segment)
	reader.AdvanceToEOL()
	return Continue | NoChildren
}

func (b *htmlBlockParser) Close(_ ast.Node, _ text.Reader, _ Context) {
	// nothing to do
}

func (b *htmlBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *htmlBlockParser) CanAcceptIndentedLine() bool {
	return false
}
