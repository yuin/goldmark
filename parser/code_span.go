package parser

import (
	"bytes"
	"io"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type codeSpanParser struct {
}

var defaultCodeSpanParser = &codeSpanParser{}

// NewCodeSpanParser return a new InlineParser that parses inline codes
// surrounded by '`' .
func NewCodeSpanParser() InlineParser {
	return defaultCodeSpanParser
}

func (s *codeSpanParser) Trigger() []byte {
	return []byte{'`'}
}

func (s *codeSpanParser) Parse(_ ast.Node, block text.Reader, _ Context) ast.Node {
	line, startSegment := block.PeekLine()
	opener := 0
	for opener < len(line) && line[opener] == '`' {
		opener++
	}
	block.Advance(opener)
	l, pos := block.Position()
	var builder text.ValueBuilder
	builder.Decoder(codeSpanDecoderInstance)
	for {
		line, segment := block.PeekLine()
		if line == nil {
			block.SetPosition(l, pos)
			return ast.NewText(text.NewSingleLineValueFromSegment(
				startSegment.WithStop(startSegment.Start+opener), block.Decoder()))
		}
		for i := 0; i < len(line); i++ {
			c := line[i]
			if c == '`' {
				oldi := i
				for i < len(line) && line[i] == '`' {
					i++
				}
				closure := i - oldi
				if closure == opener && (i >= len(line) || line[i] != '`') {
					index := text.NewIndex(segment.Start, segment.Start+i-closure)
					if !index.IsEmpty() {
						builder.AddIndex(index)
					}
					block.Advance(i)
					goto end
				}
			}
		}
		builder.AddSegment(segment)
		block.AdvanceLine()
	}
end:
	value := builder.BuildMultiLine()
	// trim leading and trailing space if applicable
	v := value.Bytes(block.Source())
	if !util.IsBlank(v) && len(v) >= 2 &&
		isSpaceOrNewline(v[0]) && isSpaceOrNewline(v[len(v)-1]) {
		indices := value.Indices()
		if len(indices) == 1 {
			indices[0].Start++
			if indices[0].Stop > indices[0].Start {
				indices[0].Stop--
			}
		} else if len(indices) > 1 {
			indices[0].Start++
			if indices[len(indices)-1].Stop > indices[len(indices)-1].Start {
				indices[len(indices)-1].Stop--
			}
		}
		value = text.NewMultiLineValueFromIndices(indices, codeSpanDecoderInstance)
	}
	return ast.NewCodeSpan(value)
}

func isSpaceOrNewline(c byte) bool {
	return c == ' ' || c == '\n'
}

type codeSpanDecoder struct {
}

func (d *codeSpanDecoder) Decode(b []byte) []byte {
	if bytes.IndexByte(b, '\n') >= 0 {
		return bytes.ReplaceAll(b, []byte{'\n'}, []byte{' '})
	}
	return b
}

func (d *codeSpanDecoder) DecodeTo(w io.Writer, b []byte) (int, error) {
	return w.Write(d.Decode(b))
}

var codeSpanDecoderInstance text.Decoder = &codeSpanDecoder{}
