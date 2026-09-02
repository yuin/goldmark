package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// A DelimiterProcessor interface provides a set of functions about
// Delimiter nodes.
type DelimiterProcessor interface {
	// IsDelimiter returns true if given character is a delimiter, otherwise false.
	IsDelimiter(byte) bool

	// CanOpenCloser returns true if given opener can be matched with given closer to
	// open a span that the closer ends, otherwise false.
	CanOpenCloser(opener, closer *Delimiter) bool

	// OnMatch will be called when new matched delimiter found.
	// OnMatch should return a new Node correspond to the matched delimiter.
	OnMatch(consumes int) ast.Node
}

// A Delimiter struct represents a delimiter like '*' of the Markdown text.
type Delimiter struct {
	ast.BaseInline

	value text.Segment

	decoder text.Decoder

	// CanOpen is set true if this delimiter can open a span for a new node.
	// See https://spec.commonmark.org/0.30/#can-open-emphasis for details.
	CanOpen bool

	// CanClose is set true if this delimiter can close a span for a new node.
	// See https://spec.commonmark.org/0.30/#can-open-emphasis for details.
	CanClose bool

	// Length is a remaining length of this delimiter.
	Length int

	// OriginalLength is a original length of this delimiter.
	OriginalLength int

	// Char is a character of this delimiter.
	Char byte

	// PreviousDelimiter is a previous sibling delimiter node of this delimiter.
	PreviousDelimiter *Delimiter

	// NextDelimiter is a next sibling delimiter node of this delimiter.
	NextDelimiter *Delimiter

	// Processor is a DelimiterProcessor associated with this delimiter.
	Processor DelimiterProcessor
}

// Inline implements Inline.Inline.
func (d *Delimiter) Inline() {}

// Dump implements Node.Dump.
func (d *Delimiter) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(d, map[string]any{
		"CanOpen":        d.CanOpen,
		"CanClose":       d.CanClose,
		"OriginalLength": d.OriginalLength,
		"Char":           string(d.Char),
	})
}

var kindDelimiter = ast.NewNodeKind("Delimiter")

// Kind implements Node.Kind.
func (d *Delimiter) Kind() ast.NodeKind {
	return kindDelimiter
}

// ConsumeCharacters consumes delimiters.
func (d *Delimiter) ConsumeCharacters(n int) {
	d.Length -= n
	d.value = d.value.WithStop(d.value.Start + d.Length)
}

// CalcConsumption calculates how many characters should be used for opening
// a new span correspond to given closer.
func (d *Delimiter) CalcConsumption(closer *Delimiter) int {
	if (d.CanClose || closer.CanOpen) && (d.OriginalLength+closer.OriginalLength)%3 == 0 && closer.OriginalLength%3 != 0 {
		return 0
	}
	if d.Length >= 2 && closer.Length >= 2 {
		return 2
	}
	return 1
}

// NewDelimiter returns a new Delimiter node.
func NewDelimiter(canOpen, canClose bool, length int, char byte, processor DelimiterProcessor) *Delimiter {
	c := &Delimiter{
		BaseInline:        ast.BaseInline{},
		CanOpen:           canOpen,
		CanClose:          canClose,
		Length:            length,
		OriginalLength:    length,
		Char:              char,
		PreviousDelimiter: nil,
		NextDelimiter:     nil,
		Processor:         processor,
	}
	return c
}

// IsLeftFlankingDelimiterRun returns true if the position represents a
// left-flanking delimiter run as defined by the CommonMark spec.
// before is the character preceding the delimiter run and after is the
// character immediately following it.
func IsLeftFlankingDelimiterRun(before, after rune) bool {
	afterIsWhitespace := util.IsSpaceRune(after)
	afterIsPunctuation := util.IsPunctRune(after)
	beforeIsWhitespace := util.IsSpaceRune(before)
	beforeIsPunctuation := util.IsPunctRune(before)
	return !afterIsWhitespace && (!afterIsPunctuation || beforeIsWhitespace || beforeIsPunctuation)
}

// IsRightFlankingDelimiterRun returns true if the position represents a
// right-flanking delimiter run as defined by the CommonMark spec.
// before is the character preceding the delimiter run and after is the
// character immediately following it.
func IsRightFlankingDelimiterRun(before, after rune) bool {
	afterIsWhitespace := util.IsSpaceRune(after)
	afterIsPunctuation := util.IsPunctRune(after)
	beforeIsWhitespace := util.IsSpaceRune(before)
	beforeIsPunctuation := util.IsPunctRune(before)
	return !beforeIsWhitespace && (!beforeIsPunctuation || afterIsWhitespace || afterIsPunctuation)
}

// ParseDelimiter scans a delimiter from block, and if found sets its segment,
// advances the reader, pushes it onto the delimiter list, and returns it.
func ParseDelimiter(block text.Reader, minimum int, processor DelimiterProcessor, pc Context) *Delimiter {
	before := block.PrecedingCharacter()
	line, segment := block.PeekLine()
	if len(line) == 0 {
		return nil
	}
	c := line[0]
	if !processor.IsDelimiter(c) {
		return nil
	}
	j := 0
	for j < len(line) && line[j] == c {
		j++
	}
	if j < minimum {
		return nil
	}
	after := rune(' ')
	if j < len(line) {
		after = util.ToRune(line, j)
	}
	isLeft := IsLeftFlankingDelimiterRun(before, after)
	isRight := IsRightFlankingDelimiterRun(before, after)
	var canOpen, canClose bool
	if c == '_' {
		beforeIsPunctuation := util.IsPunctRune(before)
		afterIsPunctuation := util.IsPunctRune(after)
		canOpen = isLeft && (!isRight || beforeIsPunctuation)
		canClose = isRight && (!isLeft || afterIsPunctuation)
	} else {
		canOpen = isLeft
		canClose = isRight
	}
	node := NewDelimiter(canOpen, canClose, j, c, processor)
	node.value = segment.WithStop(segment.Start + j)
	node.decoder = block.Decoder()
	block.Advance(j)
	pc.PushDelimiter(node)
	return node
}

// ProcessDelimiters processes the delimiter list in the context.
// Processing will be stop when reaching the bottom.
//
// If you implement an inline parser that can have other inline nodes as
// children, you should call this function when nesting span has closed.
func ProcessDelimiters(bottom ast.Node, pc Context) {
	lastDelimiter := pc.LastDelimiter()
	if lastDelimiter == nil {
		return
	}

	var closer *Delimiter
	if b, ok := bottom.(*Delimiter); ok && b != nil {
		closer = b.NextDelimiter
	} else {
		closer = pc.FirstDelimiter()
	}
	if closer == nil {
		pc.ClearDelimiters(bottom)
		return
	}
	for closer != nil {
		if !closer.CanClose {
			closer = closer.NextDelimiter
			continue
		}
		consume := 0
		found := false
		maybeOpener := false
		var opener *Delimiter
		for opener = closer.PreviousDelimiter; opener != nil && opener != bottom; opener = opener.PreviousDelimiter {
			if opener.CanOpen && opener.Processor.CanOpenCloser(opener, closer) {
				maybeOpener = true
				consume = opener.CalcConsumption(closer)
				if consume > 0 {
					found = true
					break
				}
			}
		}
		if !found {
			next := closer.NextDelimiter
			if !maybeOpener && !closer.CanOpen {
				pc.RemoveDelimiter(closer)
			}
			closer = next
			continue
		}
		opener.ConsumeCharacters(consume)
		closer.ConsumeCharacters(consume)

		node := opener.Processor.OnMatch(consume)
		node.(interface{ SetPos(int) }).SetPos(opener.value.Start)

		parent := opener.Parent()
		child := opener.NextSibling()

		for child != nil && child != closer {
			next := child.NextSibling()
			node.AppendChild(child)
			child = next
		}
		parent.InsertAfter(opener, node)

		for c := opener.NextDelimiter; c != nil && c != closer; {
			next := c.NextDelimiter
			pc.RemoveDelimiter(c)
			c = next
		}

		if opener.Length == 0 {
			pc.RemoveDelimiter(opener)
		}

		if closer.Length == 0 {
			next := closer.NextDelimiter
			pc.RemoveDelimiter(closer)
			closer = next
		}
	}
	pc.ClearDelimiters(bottom)
}
