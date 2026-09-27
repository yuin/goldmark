package parser

import (
	"bytes"

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
	ast.Text

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
	d.Value = d.Value.WithStop(d.Value.Index().Start + d.Length)
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

func (d *Delimiter) toText() {
	p := d.Parent()
	if r, ok := p.(interface {
		ReplaceChildInPlace(target, insertee ast.Node)
	}); ok {
		r.ReplaceChildInPlace(d, &d.Text)
		return
	}
	next := d.NextSibling()
	p.RemoveChild(d)
	if next != nil {
		p.InsertBefore(next, &d.Text)
	} else {
		p.AppendChild(&d.Text)
	}
}

// NewDelimiter returns a new Delimiter node.
func NewDelimiter(canOpen, canClose bool, length int, char byte, processor DelimiterProcessor) *Delimiter {
	return &Delimiter{
		CanOpen:        canOpen,
		CanClose:       canClose,
		Length:         length,
		OriginalLength: length,
		Char:           char,
		Processor:      processor,
	}
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

// ParseDelimiterFunc scans a delimiter from block, and if found sets its segment,
// advances the reader, pushes it onto the delimiter list, and returns it.
type ParseDelimiterFunc = func(block text.Reader, minimum int, processor DelimiterProcessor, pc Context) ast.Node

// ParseDelimiter is a default implementation of [ParseDelimiterFunc] that
// follows the CommonMark spec.
func ParseDelimiter(block text.Reader, minimum int, processor DelimiterProcessor, pc Context) ast.Node {
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

	canOpen, canClose := parseDelimiterOpenClose(before, after, c)
	if !canOpen && !canClose { // not a delimiter, just a text
		return consumeText(j)
	}

	// Fast path: simple emphasis
	if canOpen && j >= minimum && (!canClose || !hasSameCharOpener(pc, c)) {
		if fast := parseSimpleDelimitedText(block, line, segment, c, j, canOpen, canClose,
			processor, pc); fast != nil {
			return fast
		}
	}

	node := NewDelimiter(canOpen, canClose, j, c, processor)
	node.Value = text.NewSingleLineValueFromSegment(segment.WithStop(segment.Start+j), block.Decoder())
	block.Advance(j)
	pc.PushDelimiter(node)
	return node
}

func parseSimpleDelimitedText(block text.Reader, line []byte, segment text.Segment,
	c byte, openerLen int, canOpen, canClose bool,
	processor DelimiterProcessor, pc Context) ast.Node {
	if !canOpen {
		return nil
	}
	pctx, ok := pc.(*parseContext)
	if !ok || pctx.inlineParsers == nil {
		return nil
	}
	parsers := pctx.inlineParsers

	// hop to the next occurrence of c and check whether it can close our opener.
	scanStart := openerLen
	var absIdx, closerLen int
	for {
		idx := bytes.IndexByte(line[scanStart:], c)
		if idx < 0 {
			return nil
		}
		absIdx = scanStart + idx

		k := absIdx
		for k < len(line) && line[k] == c {
			k++
		}
		closerLen = k - absIdx

		// Opener-shorter-than-closer: fast path unsupported.
		if closerLen > openerLen {
			return nil
		}

		// Cheap byte-level pre-filter. `absIdx >= openerLen >= 1`,
		// so absIdx-1 is always a valid index.
		if prev := line[absIdx-1]; prev == ' ' || prev == '\t' || prev == '\\' {
			scanStart = k
			continue
		}

		// Strict rune-based flanking.
		beforeRune := util.ToRune(line, absIdx-1)
		afterRune := rune(' ')
		if k < len(line) && line[k] != '\n' {
			afterRune = util.ToRune(line, k)
		}
		closerCanOpen, closerCanClose := parseDelimiterOpenClose(beforeRune, afterRune, c)
		if !closerCanClose {
			scanStart = k
			continue
		}
		// CommonMark Rule 9
		if (canClose || closerCanOpen) && (openerLen+closerLen)%3 == 0 && closerLen%3 != 0 {
			return nil
		}
		break
	}

	k := absIdx + closerLen

	// fast path supports only single-text node.
	for i := openerLen; i < absIdx; i++ {
		ch := line[i]
		if ch == '\\' && i+1 < absIdx {
			i++
			continue
		}
		if ch == '\n' {
			return nil
		}
		if parsers[ch] != nil {
			return nil
		}
	}

	if leftover := openerLen - closerLen; leftover > 0 {
		return nil
	}

	contentSeg := text.NewSegment(segment.Start+openerLen, segment.Start+absIdx)
	var inner ast.Node = ast.NewText(text.NewSingleLineValueFromSegment(contentSeg, block.Decoder()))
	remaining := closerLen
	for remaining > 0 {
		consume := 1
		if remaining >= 2 {
			consume = 2
		}
		container := processor.OnMatch(consume)
		if container == nil {
			return nil
		}
		container.AppendChild(inner)
		container.SetPos(segment.Start + (openerLen - closerLen))
		inner = container
		remaining -= consume
	}
	block.Advance(k)
	return inner
}

func parseDelimiterOpenClose(before, after rune, c byte) (canOpen, canClose bool) {
	beforeIsSpace, beforeIsPunct := runeClass(before)
	afterIsSpace, afterIsPunct := runeClass(after)
	isLeft := !afterIsSpace && (!afterIsPunct || beforeIsSpace || beforeIsPunct)
	isRight := !beforeIsSpace && (!beforeIsPunct || afterIsSpace || afterIsPunct)
	if c == '_' {
		canOpen = isLeft && (!isRight || beforeIsPunct)
		canClose = isRight && (!isLeft || afterIsPunct)
	} else {
		canOpen = isLeft
		canClose = isRight
	}
	return
}

// ParseDelimiterSimple is a simpler implementation of [ParseDelimiterFunc].
//
// This function determines flankings based on original Markdown like simple rules:
//
//   - If the character after the delimiter is a space, the delimiter cannot open a span.
//   - If the character before the delimiter is a space, the delimiter cannot close a span.
//
// These rules are easy to understand even for non-engineer writers.
// While CommonMark rules do not work well with CJK, these rules often work well with CJK.
func ParseDelimiterSimple(block text.Reader, minimum int, processor DelimiterProcessor, pc Context) ast.Node {
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

	last := pc.LastDelimiter()
	beforeIsDelimiter := false
	if last != nil && last.Value.Index().Stop == segment.Start {
		beforeIsDelimiter = true
		last.CanClose = true
	}
	isLeft := !util.IsSpaceRune(after)
	isRight := !util.IsSpaceRune(before) || beforeIsDelimiter
	var canOpen, canClose bool
	if c == '_' {
		canOpen = isLeft && (!isRight || beforeIsDelimiter)
		canClose = isRight && !isLeft
	} else {
		canOpen = isLeft
		canClose = isRight
	}

	node := NewDelimiter(canOpen, canClose, j, c, processor)
	node.Value = text.NewSingleLineValueFromSegment(segment.WithStop(segment.Start+j), block.Decoder())
	block.Advance(j)
	pc.PushDelimiter(node)
	return node
}

func hasSameCharOpener(pc Context, c byte) bool {
	for d := pc.FirstDelimiter(); d != nil; d = d.NextDelimiter {
		if d.Char == c && d.CanOpen {
			return true
		}
	}
	return false
}

func runeClass(r rune) (isSpace, isPunct bool) {
	if r < 0x80 {
		flags := charFlags[byte(r)]
		return flags&charFlagSpace != 0 || r == '\n' || r == '\r', flags&charFlagPunct != 0
	}
	return util.IsSpaceRune(r), util.IsPunctRune(r)
}

// delimiterClassCount is the size of the openersBottom table in
// ProcessDelimiters: one slot per (Char, CanOpen, Length%3) combination.
const delimiterClassCount = 256 * 2 * 3

type delimiterOpenersBottoms struct {
	values   [8]delimiterOpenerBottom
	length   int
	overflow *[delimiterClassCount]int
}

type delimiterOpenerBottom struct {
	index      int
	lowerBound int
}

func (d *delimiterOpenersBottoms) get(index int) (int, bool) {
	for _, v := range d.values[:d.length] {
		if v.index == index {
			return v.lowerBound, true
		}
	}
	if d.overflow == nil || d.overflow[index] == 0 {
		return 0, false
	}
	return d.overflow[index] - 1, true
}

func (d *delimiterOpenersBottoms) set(index, lowerBound int) {
	for i := range d.values[:d.length] {
		if d.values[i].index == index {
			d.values[i].lowerBound = lowerBound
			return
		}
	}
	if d.length < len(d.values) {
		d.values[d.length] = delimiterOpenerBottom{index, lowerBound}
		d.length++
		return
	}
	if d.overflow == nil {
		d.overflow = new([delimiterClassCount]int)
		for _, v := range d.values {
			d.overflow[v.index] = v.lowerBound + 1
		}
	}
	d.overflow[index] = lowerBound + 1
}

func delimiterClassIndex(char byte, canOpen bool, lengthMod3 int) int {
	idx := int(char) * 6
	if canOpen {
		idx += 3
	}
	return idx + lengthMod3
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

	var openersBottom delimiterOpenersBottoms

	for closer != nil {
		if !closer.CanClose {
			closer = closer.NextDelimiter
			continue
		}
		idx := delimiterClassIndex(closer.Char, closer.CanOpen, closer.Length%3)
		lowerBound, hasLowerBound := openersBottom.get(idx)

		consume := 0
		found := false
		maybeOpener := false
		var opener *Delimiter
		for opener = closer.PreviousDelimiter; opener != nil && opener != bottom &&
			(!hasLowerBound || opener.Value.Index().Start >= lowerBound); opener = opener.PreviousDelimiter {
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
			openersBottom.set(idx, closer.Value.Index().Start)
			closer = next
			continue
		}
		opener.ConsumeCharacters(consume)
		closer.ConsumeCharacters(consume)

		node := opener.Processor.OnMatch(consume)
		node.SetPos(opener.Value.Index().Start)

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
