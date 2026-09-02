package parser

import (
	"bytes"
	"io"
	"slices"

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
	builder.Decoder(nil)
	// firstLine is true while no line has been skipped yet (i.e. we are still
	// looking at the line right after the opener). The overwhelmingly common
	// case is a closer found on this very line, e.g. `code`; that case is
	// handled without ever touching the ValueBuilder, avoiding the slice
	// allocation multiLineCodeSpanValue would otherwise require.
	firstLine := true
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
					block.Advance(i)
					if firstLine {
						return ast.NewCodeSpan(singleLineCodeSpanValue{start: segment.Start, stop: segment.Start + i - closure})
					}
					index := text.NewIndex(segment.Start, segment.Start+i-closure)
					if !index.IsEmpty() {
						builder.AddIndex(index)
					}
					goto end
				}
			}
		}
		builder.AddSegment(segment)
		block.AdvanceLine()
		firstLine = false
	}
end:
	value := builder.BuildMultiLine()
	indices := value.Indices()
	if len(indices) == 1 {
		return ast.NewCodeSpan(singleLineCodeSpanValue{start: indices[0].Start, stop: indices[0].Stop})
	}
	return ast.NewCodeSpan(multiLineCodeSpanValue{indices: indices})
}

var _ text.Value = (*singleLineCodeSpanValue)(nil)

// singleLineCodeSpanValue is a [text.Value] implementation for inline code spans.
//
// this [text.Value] normalizes the value of inline code spans by trimming leading and trailing spaces and
// replacing newlines with spaces, as per the CommonMark specification.
//
// singleLineCodeSpanValue is optimized to avoid allocations; this struct is small enough to be
// assigned to an interface-type variable without allocations.
type singleLineCodeSpanValue struct {
	start int
	stop  int
}

func (v singleLineCodeSpanValue) Index() text.Index {
	return text.NewIndex(v.start, v.stop)
}

func (v singleLineCodeSpanValue) Indices() []text.Index {
	return []text.Index{v.Index()}
}

func (v singleLineCodeSpanValue) IsEmpty() bool {
	return v.start >= v.stop
}

func (v singleLineCodeSpanValue) IsOwned() bool {
	return false
}

func (v singleLineCodeSpanValue) Value(source []byte) string {
	return v.Str(source)
}

func (v singleLineCodeSpanValue) WriteTo(w io.Writer, source []byte) (int, error) {
	i := v.trimmedIndex(source)
	b := source[i.Start:i.Stop]
	if len(b) == 0 {
		return 0, nil
	}
	return writeNewlinesWithSpaces(w, b)
}

func (v singleLineCodeSpanValue) Bytes(source []byte) []byte {
	i := v.trimmedIndex(source)
	b := source[i.Start:i.Stop]
	if len(b) == 0 {
		return nil
	}
	return replaceNewlinesWithSpaces(b)
}

func (v singleLineCodeSpanValue) Str(source []byte) string {
	return util.BytesToReadOnlyString(v.Bytes(source))
}

func (v singleLineCodeSpanValue) shouldTrimSpaces(source []byte) bool {
	b := source[v.start:v.stop]
	if len(b) < 2 {
		return false
	}
	if util.IsBlank(b) {
		return false
	}
	return isSpaceOrNewline(b[0]) && isSpaceOrNewline(b[len(b)-1])
}

func (v singleLineCodeSpanValue) trimmedIndex(source []byte) text.Index {
	if v.shouldTrimSpaces(source) {
		start := v.start + 1
		stop := v.stop
		if v.stop > v.start {
			stop--
		}
		return text.NewIndex(start, stop)
	}
	return text.NewIndex(v.start, v.stop)
}

var _ text.Value = (*multiLineCodeSpanValue)(nil)

// multiLineCodeSpanValue is a text.Value implementation for inline code spans.
//
// this [text.Value] normalizes the value of inline code spans by trimming leading and trailing spaces and
// replacing newlines with spaces, as per the CommonMark specification.
type multiLineCodeSpanValue struct {
	indices []text.Index
}

func (v multiLineCodeSpanValue) Index() text.Index {
	if len(v.indices) > 0 {
		return v.indices[0]
	}
	return text.NewIndex(0, 0)
}

func (v multiLineCodeSpanValue) Indices() []text.Index {
	return v.indices
}

func (v multiLineCodeSpanValue) IsEmpty() bool {
	for _, idx := range v.indices {
		if idx.Start < idx.Stop {
			return false
		}
	}
	return true
}

func (v multiLineCodeSpanValue) IsOwned() bool {
	return false
}

func (v multiLineCodeSpanValue) Value(source []byte) string {
	return v.Str(source)
}

func (v multiLineCodeSpanValue) WriteTo(w io.Writer, source []byte) (int, error) {
	indices := v.trimmedIndices(source)
	if len(indices) == 0 {
		return 0, nil
	}
	if len(indices) == 1 {
		return writeNewlinesWithSpaces(w, source[indices[0].Start:indices[0].Stop])
	}
	n := int(0)
	for _, idx := range indices {
		written, err := writeNewlinesWithSpaces(w, source[idx.Start:idx.Stop])
		n += written
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (v multiLineCodeSpanValue) Bytes(source []byte) []byte {
	indices := v.trimmedIndices(source)
	if len(indices) == 0 {
		return nil
	}
	if len(indices) == 1 {
		return replaceNewlinesWithSpaces(source[indices[0].Start:indices[0].Stop])
	}
	buf := slices.Clone(replaceNewlinesWithSpaces(source[indices[0].Start:indices[0].Stop]))
	for _, idx := range indices[1:] {
		chunk := replaceNewlinesWithSpaces(source[idx.Start:idx.Stop])
		buf = append(buf, chunk...)
	}
	return buf
}

func (v multiLineCodeSpanValue) Str(source []byte) string {
	return util.BytesToReadOnlyString(v.Bytes(source))
}

func (v multiLineCodeSpanValue) shouldTrimSpaces(source []byte) bool {
	indices := v.Indices()
	if len(indices) == 0 {
		return false
	}
	if len(indices) == 1 {
		b := source[indices[0].Start:indices[0].Stop]
		if len(b) < 2 {
			return false
		}
		if util.IsBlank(b) {
			return false
		}
		return isSpaceOrNewline(b[0]) && isSpaceOrNewline(b[len(b)-1])
	}
	first := source[indices[0].Start:indices[0].Stop]
	last := source[indices[len(indices)-1].Start:indices[len(indices)-1].Stop]
	if len(first) < 1 || len(last) < 1 {
		return false
	}
	hasNonBlank := false
	for _, index := range indices {
		b := source[index.Start:index.Stop]
		if !util.IsBlank(b) {
			hasNonBlank = true
			break
		}
	}
	if !hasNonBlank {
		return false
	}
	return isSpaceOrNewline(first[0]) && isSpaceOrNewline(last[len(last)-1])
}

func (v multiLineCodeSpanValue) trimmedIndices(source []byte) []text.Index {
	if v.shouldTrimSpaces(source) {
		indices := v.Indices()
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
		return indices
	}
	return v.Indices()
}

func isSpaceOrNewline(c byte) bool {
	return c == ' ' || c == '\n'
}

func replaceNewlinesWithSpaces(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte{'\n'}, []byte{' '})
}

func writeNewlinesWithSpaces(w io.Writer, b []byte) (int, error) {
	if bytes.IndexByte(b, '\n') == -1 {
		return w.Write(b)
	}
	written := 0
	for {
		i := bytes.IndexByte(b, '\n')
		if i == -1 {
			wr, err := w.Write(b)
			written += wr
			return written, err
		}
		wr, err := w.Write(b[:i])
		written += wr
		if err != nil {
			return written, err
		}
		wr, err = w.Write([]byte{' '})
		written += wr
		if err != nil {
			return written, err
		}
		b = b[i+1:]
	}
}
