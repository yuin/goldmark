package parser

import (
	"bytes"
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

var htmlBlockType2Close = []byte{'-', '-', '>'}
var htmlBlockType3Close = []byte{'?', '>'}
var htmlBlockType4Close = []byte{'>'}
var htmlBlockType5Close = []byte{']', ']', '>'}

func countLeadingSpaces(line []byte) (int, bool) {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	return i, i <= 3
}

func scanHTMLBlockOpen1(line []byte) bool {
	i, ok := countLeadingSpaces(line)
	if !ok || i >= len(line) || line[i] != '<' {
		return false
	}
	rest := line[i+1:]
	var n int
	switch {
	case hasPrefixFold(rest, "textarea"):
		n = len("textarea")
	case hasPrefixFold(rest, "script"):
		n = len("script")
	case hasPrefixFold(rest, "style"):
		n = len("style")
	case hasPrefixFold(rest, "pre"):
		n = len("pre")
	default:
		return false
	}
	rest = rest[n:]
	if len(rest) == 0 {
		return true
	}
	c := rest[0]
	return util.IsSpace(c) || c == '>' || (c == '/' && len(rest) > 1 && rest[1] == '>')
}

func scanHTMLBlockClose1(line []byte) bool {
	return containsFold(line, "</script>") || containsFold(line, "</pre>") ||
		containsFold(line, "</style>") || containsFold(line, "</textarea>")
}

func scanHTMLBlockOpen2(line []byte) bool {
	i, ok := countLeadingSpaces(line)
	return ok && bytes.HasPrefix(line[i:], []byte("<!--"))
}

func scanHTMLBlockOpen3(line []byte) bool {
	i, ok := countLeadingSpaces(line)
	return ok && bytes.HasPrefix(line[i:], []byte("<?"))
}

func scanHTMLBlockOpen4(line []byte) bool {
	i, ok := countLeadingSpaces(line)
	if !ok || !bytes.HasPrefix(line[i:], []byte("<!")) {
		return false
	}
	i += 2
	return i < len(line) && line[i] >= 'A' && line[i] <= 'Z'
}

func scanHTMLBlockOpen5(line []byte) bool {
	i, ok := countLeadingSpaces(line)
	return ok && bytes.HasPrefix(line[i:], []byte("<![CDATA["))
}

func scanHTMLBlockOpen6(line []byte) (tagName []byte, ok bool) {
	i, ok2 := countLeadingSpaces(line)
	if !ok2 || i >= len(line) || line[i] != '<' {
		return nil, false
	}
	i++
	if i < len(line) && line[i] == '/' {
		i++
		for i < len(line) && line[i] == ' ' {
			i++
		}
	}
	start := i
	end, found := scanTagNameBytes(line, i)
	if !found {
		return nil, false
	}
	if end < len(line) {
		c := line[end]
		if c != ' ' && c != '>' && (c != '/' || end+1 >= len(line) || line[end+1] != '>') {
			return nil, false
		}
	}
	return line[start:end], true
}

func scanHTMLBlockOpen7(line []byte) (tagName []byte, isCloseTag, hasAttr, ok bool) {
	i, ok2 := countLeadingSpaces(line)
	if !ok2 || i >= len(line) || line[i] != '<' {
		return nil, false, false, false
	}
	i++
	if i < len(line) && line[i] == '/' {
		isCloseTag = true
		i++
		for i < len(line) && line[i] == ' ' {
			i++
		}
	}
	start := i
	end, found := scanTagNameBytes(line, i)
	if !found {
		return nil, false, false, false
	}
	tagName = line[start:end]
	end, hasAttr = scanHTMLAttributesBytes(line, end)
	for end < len(line) && line[end] == ' ' {
		end++
	}
	if end >= len(line) {
		return tagName, isCloseTag, hasAttr, false
	}
	switch line[end] {
	case '/':
		if end+1 >= len(line) || line[end+1] != '>' {
			return tagName, isCloseTag, hasAttr, false
		}
		end += 2
	case '>':
		end++
	default:
		return tagName, isCloseTag, hasAttr, false
	}
	for end < len(line) && line[end] == ' ' {
		end++
	}
	if end < len(line) && line[end] == '\r' && end+1 < len(line) && line[end+1] == '\n' {
		end += 2
	} else if end < len(line) && line[end] == '\n' {
		end++
	}
	return tagName, isCloseTag, hasAttr, end == len(line)
}

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

	if scanHTMLBlockOpen1(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind1)
	} else if scanHTMLBlockOpen2(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind2)
	} else if scanHTMLBlockOpen3(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind3)
	} else if scanHTMLBlockOpen4(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind4)
	} else if scanHTMLBlockOpen5(line) {
		node = ast.NewHTMLBlock(ast.HTMLBlockKind5)
	} else if rawTagName, isCloseTag, hasAttr, ok := scanHTMLBlockOpen7(line); ok {
		tagName := strings.ToLower(string(rawTagName))
		_, ok := allowedBlockTags[tagName]
		if ok {
			node = ast.NewHTMLBlock(ast.HTMLBlockKind6)
		} else if tagName != "script" && tagName != "style" &&
			tagName != "pre" && !ast.IsParagraph(last) && (!isCloseTag || !hasAttr) { // type 7 can not interrupt paragraph
			node = ast.NewHTMLBlock(ast.HTMLBlockKind7)
		}
	}
	if node == nil {
		if tagName, ok := scanHTMLBlockOpen6(line); ok {
			_, ok := allowedBlockTags[strings.ToLower(string(tagName))]
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
			if scanHTMLBlockClose1(firstLine.Bytes(reader.Source())) {
				return Close
			}
		}
		if scanHTMLBlockClose1(line) {
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
