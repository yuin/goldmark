package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var linkLabelStateKey = NewContextKey()

type linkLabelState struct {
	ast.BaseInline

	value text.Segment

	IsImage bool

	Prev *linkLabelState

	Next *linkLabelState

	First *linkLabelState

	Last *linkLabelState
}

func newLinkLabelState(segment text.Segment, isImage bool) *linkLabelState {
	return &linkLabelState{
		value:   segment,
		IsImage: isImage,
	}
}

func (s *linkLabelState) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(s, map[string]any{
		"IsImage": s.IsImage,
	})
}

var kindLinkLabelState = ast.NewNodeKind("LinkLabelState")

func (s *linkLabelState) Kind() ast.NodeKind {
	return kindLinkLabelState
}

func linkLabelStateLength(v *linkLabelState) int {
	if v == nil || v.Last == nil || v.First == nil {
		return 0
	}
	return v.Last.value.Stop - v.First.value.Start
}

func pushLinkLabelState(pc Context, v *linkLabelState) {
	tlist := pc.Get(linkLabelStateKey)
	var list *linkLabelState
	if tlist == nil {
		list = v
		v.First = v
		v.Last = v
		pc.Set(linkLabelStateKey, list)
	} else {
		list = tlist.(*linkLabelState)
		l := list.Last
		list.Last = v
		l.Next = v
		v.Prev = l
	}
}

func removeLinkLabelState(pc Context, d *linkLabelState) {
	tlist := pc.Get(linkLabelStateKey)
	var list *linkLabelState
	if tlist == nil {
		return
	}
	list = tlist.(*linkLabelState)

	if d.Prev == nil {
		list = d.Next
		if list != nil {
			list.First = d
			list.Last = d.Last
			list.Prev = nil
			pc.Set(linkLabelStateKey, list)
		} else {
			pc.Set(linkLabelStateKey, nil)
		}
	} else {
		d.Prev.Next = d.Next
		if d.Next != nil {
			d.Next.Prev = d.Prev
		}
	}
	if list != nil && d.Next == nil {
		list.Last = d.Prev
	}
	d.Next = nil
	d.Prev = nil
	d.First = nil
	d.Last = nil
}

type linkParser struct {
}

var defaultLinkParser = &linkParser{}

// NewLinkParser return a new InlineParser that parses links.
func NewLinkParser() InlineParser {
	return defaultLinkParser
}

func (s *linkParser) Trigger() []byte {
	return []byte{'!', '[', ']'}
}

var linkBottom = NewContextKey()

func (s *linkParser) Parse(parent ast.Node, block text.Reader, pc Context) ast.Node {
	line, segment := block.PeekLine()
	if line[0] == '!' {
		if len(line) > 1 && line[1] == '[' {
			block.Advance(1)
			pushLinkBottom(pc)
			return processLinkLabelOpen(block, segment.Start+1, true, pc)
		}
		return nil
	}
	if line[0] == '[' {
		pushLinkBottom(pc)
		return processLinkLabelOpen(block, segment.Start, false, pc)
	}

	// line[0] == ']'
	tlist := pc.Get(linkLabelStateKey)
	if tlist == nil {
		return nil
	}
	last := tlist.(*linkLabelState).Last
	if last == nil {
		_ = popLinkBottom(pc)
		return nil
	}
	block.Advance(1)
	removeLinkLabelState(pc, last)
	// CommonMark spec says:
	//  > A link label can have at most 999 characters inside the square brackets.
	if linkLabelStateLength(tlist.(*linkLabelState)) > 998 {
		ast.MergeOrReplaceTextSegment(last.Parent(), last, last.value)
		_ = popLinkBottom(pc)
		return nil
	}

	if !last.IsImage && s.containsLink(last) { // a link in a link text is not allowed
		ast.MergeOrReplaceTextSegment(last.Parent(), last, last.value)
		_ = popLinkBottom(pc)
		return nil
	}

	c := block.Peek()
	l, pos := block.Position()
	var link *ast.Link
	var hasValue bool
	switch c {
	case '(':
		link = s.parseLink(parent, last, block, pc)
	case '[':
		link, hasValue = s.parseReferenceLink(parent, last, block, pc)
		if link == nil && hasValue {
			ast.MergeOrReplaceTextSegment(last.Parent(), last, last.value)
			_ = popLinkBottom(pc)
			return nil
		}
	}

	if link == nil {
		// maybe shortcut reference link
		block.SetPosition(l, pos)
		ssegment := text.NewSegment(last.value.Stop, segment.Start)
		maybeReferenceValue := block.ValueBetween(ssegment.Start, ssegment.Stop)
		maybeReference := maybeReferenceValue.Bytes(block.Source())
		// CommonMark spec says:
		//  > A link label can have at most 999 characters inside the square brackets.
		if len(maybeReference) > 999 {
			ast.MergeOrReplaceTextSegment(last.Parent(), last, last.value)
			_ = popLinkBottom(pc)
			return nil
		}

		ref, ok := pc.LinkDefinition(util.ToLinkReference(maybeReference))
		if !ok {
			ast.MergeOrReplaceTextSegment(last.Parent(), last, last.value)
			_ = popLinkBottom(pc)
			return nil
		}
		link = referenceLink(ref, ast.ReferenceLinkKindShortcut, maybeReferenceValue)
		s.processLinkLabel(parent, link, last, pc)
	}
	var n ast.Node
	if last.IsImage {
		last.Parent().RemoveChild(last)
		img := ast.NewImage(link.Destination, ast.WithLinkTitle(link.Title))
		img.Reference = link.Reference
		for c := link.FirstChild(); c != nil; {
			next := c.NextSibling()
			link.RemoveChild(c)
			img.AppendChild(c)
			c = next
		}
		n = img
	} else {
		last.Parent().RemoveChild(last)
		n = link
	}
	n.(interface{ SetPos(int) }).SetPos(last.value.Start)
	return n
}

func (s *linkParser) containsLink(n ast.Node) bool {
	if n == nil {
		return false
	}
	for c := n; c != nil; c = c.NextSibling() {
		if _, ok := c.(*ast.Link); ok {
			return true
		}
		if s.containsLink(c.FirstChild()) {
			return true
		}
	}
	return false
}

func processLinkLabelOpen(block text.Reader, pos int, isImage bool, pc Context) *linkLabelState {
	start := pos
	if isImage {
		start--
	}
	state := newLinkLabelState(text.NewSegment(start, pos+1), isImage)
	pushLinkLabelState(pc, state)
	block.Advance(1)
	return state
}

func (s *linkParser) processLinkLabel(parent ast.Node, link *ast.Link, last *linkLabelState, pc Context) {
	bottom := popLinkBottom(pc)
	ProcessDelimiters(bottom, pc)
	for c := last.NextSibling(); c != nil; {
		next := c.NextSibling()
		parent.RemoveChild(c)
		link.AppendChild(c)
		c = next
	}
}

func findClosure(r text.Reader, opener, closer byte) (text.MultiLineValue, bool) {
	orgLine, orgPos := r.Position()
	var segs []text.Segment
	for {
		bs, seg := r.PeekLine()
		if bs == nil {
			break
		}
		for i := 0; i < len(bs); i++ {
			c := bs[i]
			if c == '\\' && i < len(bs)-1 && util.IsPunct(bs[i+1]) {
				i++
				continue
			}
			if c == closer {
				segs = append(segs, seg.WithStop(seg.Start+i-seg.Padding))
				r.Advance(i + 1)
				var b text.ValueBuilder
				for _, s := range segs {
					b.AddSegment(s)
				}
				return b.BuildMultiLine(), true
			}
			if c == opener {
				r.SetPosition(orgLine, orgPos)
				return text.MultiLineValue{}, false
			}
		}
		r.AdvanceLine()
		segs = append(segs, seg)
	}
	r.SetPosition(orgLine, orgPos)
	return text.MultiLineValue{}, false
}

func (s *linkParser) parseReferenceLink(parent ast.Node, last *linkLabelState,
	block text.Reader, pc Context) (*ast.Link, bool) {
	_, orgpos := block.Position()
	block.Advance(1) // skip '['
	maybeReferenceValue, found := findClosure(block, '[', ']')
	if !found {
		return nil, false
	}

	refType := ast.ReferenceLinkKindFull
	maybeReference := maybeReferenceValue.Bytes(block.Source())
	if util.IsBlank(maybeReference) { // collapsed reference link
		maybeReference = block.ValueBetween(last.value.Stop, orgpos.Start-1).Bytes(block.Source())
		refType = ast.ReferenceLinkKindCollapsed
	}
	// CommonMark spec says:
	//  > A link label can have at most 999 characters inside the square brackets.
	if len(maybeReference) > 999 {
		return nil, true
	}

	ref, ok := pc.LinkDefinition(util.ToLinkReference(maybeReference))
	if !ok {
		return nil, true
	}
	link := referenceLink(ref, refType, maybeReferenceValue)
	s.processLinkLabel(parent, link, last, pc)
	return link, true
}

func (s *linkParser) parseLink(parent ast.Node, last *linkLabelState, block text.Reader, pc Context) *ast.Link {
	block.Advance(1) // skip '('
	block.SkipSpaces()
	var title text.MultiLineValue
	var destination text.SingleLineValue
	var ok bool
	if block.Peek() == ')' { // empty link like '[link]()'
		block.Advance(1)
	} else {
		destination, ok = parseLinkDestination(block)
		if !ok {
			return nil
		}
		block.SkipSpaces()
		if block.Peek() == ')' {
			block.Advance(1)
		} else {
			title, ok = parseLinkTitle(block)
			if !ok {
				return nil
			}
			block.SkipSpaces()
			if block.Peek() == ')' {
				block.Advance(1)
			} else {
				return nil
			}
		}
	}

	link := ast.NewLink(destination, ast.WithLinkTitle(title))
	s.processLinkLabel(parent, link, last, pc)
	return link
}

func parseLinkDestination(block text.Reader) (text.SingleLineValue, bool) {
	block.SkipSpaces()
	line, segment := block.PeekLine()
	if block.Peek() == '<' {
		i := 1
		for i < len(line) {
			c := line[i]
			if c == '\\' && i < len(line)-1 && util.IsPunct(line[i+1]) {
				i += 2
				continue
			} else if c == '>' {
				block.Advance(i + 1)
				return text.NewIndexSingleLineValue(text.NewIndex(segment.Start+1, segment.Start+i)), true
			}
			i++
		}
		return text.SingleLineValue{}, false
	}
	opened := 0
	i := 0
	for i < len(line) {
		c := line[i]
		if c == '\\' && i < len(line)-1 && util.IsPunct(line[i+1]) {
			i += 2
			continue
		} else if c == '(' {
			opened++
		} else if c == ')' {
			opened--
			if opened < 0 {
				break
			}
		} else if util.IsSpace(c) {
			break
		}
		i++
	}
	block.Advance(i)
	if i == 0 {
		return text.SingleLineValue{}, false
	}
	return text.NewIndexSingleLineValue(text.NewIndex(segment.Start, segment.Start+i)), true
}

func parseLinkTitle(block text.Reader) (text.MultiLineValue, bool) {
	block.SkipSpaces()
	opener := block.Peek()
	if opener != '"' && opener != '\'' && opener != '(' {
		return text.MultiLineValue{}, false
	}
	closer := opener
	if opener == '(' {
		closer = ')'
	}
	block.Advance(1)
	mv, found := findClosure(block, opener, closer)
	if found {
		return mv, true
	}
	return text.MultiLineValue{}, false
}

func pushLinkBottom(pc Context) {
	bottoms := pc.Get(linkBottom)
	b := pc.LastDelimiter()
	if bottoms == nil {
		pc.Set(linkBottom, b)
		return
	}
	if s, ok := bottoms.([]ast.Node); ok {
		pc.Set(linkBottom, append(s, b))
		return
	}
	pc.Set(linkBottom, []ast.Node{bottoms.(ast.Node), b})
}

func popLinkBottom(pc Context) ast.Node {
	bottoms := pc.Get(linkBottom)
	if bottoms == nil {
		return nil
	}
	if v, ok := bottoms.(ast.Node); ok {
		pc.Set(linkBottom, nil)
		return v
	}
	s := bottoms.([]ast.Node)
	v := s[len(s)-1]
	n := s[0 : len(s)-1]
	switch len(n) {
	case 0:
		pc.Set(linkBottom, nil)
	case 1:
		pc.Set(linkBottom, n[0])
	default:
		pc.Set(linkBottom, s[0:len(s)-1])
	}
	return v
}

func referenceLink(def LinkDefinition, kind ast.ReferenceLinkKind, refvalue text.MultiLineValue) *ast.Link {
	var link *ast.Link
	if ld, ok := def.(*linkDefinition); ok && ld.node != nil {
		link = ast.NewLink(
			ld.node.Destination,
			ast.WithLinkTitle(ld.node.Title),
			ast.WithLinkReference(kind, refvalue),
		)
	} else {
		link = ast.NewLink(
			text.NewStringSingleLineValue(string(def.Destination())),
			ast.WithLinkTitle(text.NewStringMultiLineValue(string(def.Title()))),
			ast.WithLinkReference(kind, refvalue),
		)
	}
	return link
}

func (s *linkParser) CloseBlock(_ ast.Node, _ text.Reader, pc Context) {
	pc.Set(linkBottom, nil)
	tlist := pc.Get(linkLabelStateKey)
	if tlist == nil {
		return
	}
	for s := tlist.(*linkLabelState); s != nil; {
		next := s.Next
		removeLinkLabelState(pc, s)
		s.Parent().ReplaceChild(s, ast.NewSegmentText(s.value))
		s = next
	}
}
