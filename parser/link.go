package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var linkLabelStateKey = NewContextKey()

type linkLabelState struct {
	ast.Text

	bottom *Delimiter

	IsImage           bool
	hasLinkDescendant bool
}

func (s *linkLabelState) value() text.Index {
	return s.Value.Index()
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

func (s *linkLabelState) toText() {
	p := s.Parent()
	if r, ok := p.(interface {
		ReplaceChildInPlace(target, insertee ast.Node)
	}); ok {
		r.ReplaceChildInPlace(s, &s.Text)
		return
	}
	next := s.NextSibling()
	p.RemoveChild(s)
	if next != nil {
		p.InsertBefore(next, &s.Text)
	} else {
		p.AppendChild(&s.Text)
	}
}

type linkLabelStateStack struct {
	states []*linkLabelState

	slab []linkLabelState
	slot int
}

func (s *linkLabelStateStack) newState() *linkLabelState {
	if s.slot >= len(s.slab) {
		n := max(len(s.slab)*2, 8)
		s.slab = make([]linkLabelState, n)
		s.slot = 0
	}
	st := &s.slab[s.slot]
	s.slot++
	return st
}

func linkLabelStateStackFor(pc Context) *linkLabelStateStack {
	v := pc.Get(linkLabelStateKey)
	if v != nil {
		return v.(*linkLabelStateStack)
	}
	s := &linkLabelStateStack{}
	pc.Set(linkLabelStateKey, s)
	return s
}

func linkLabelSpan(states []*linkLabelState) int {
	if len(states) == 0 {
		return 0
	}
	v := states[len(states)-1].value()
	return v.Stop - v.Start
}

func popLastLinkLabelState(pc Context) *linkLabelState {
	v := pc.Get(linkLabelStateKey)
	if v == nil {
		return nil
	}
	s := v.(*linkLabelStateStack)
	n := len(s.states)
	if n == 0 {
		return nil
	}
	last := s.states[n-1]
	s.states[n-1] = nil
	s.states = s.states[:n-1]
	return last
}

func openLinkLabelStates(pc Context) []*linkLabelState {
	v := pc.Get(linkLabelStateKey)
	if v == nil {
		return nil
	}
	return v.(*linkLabelStateStack).states
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

func (s *linkParser) Parse(parent ast.Node, block text.Reader, pc Context) ast.Node {
	line, segment := block.PeekLine()
	if line[0] == '!' {
		if len(line) > 1 && line[1] == '[' {
			if fast := s.parseSimpleLink(block, true, pc); fast != nil {
				return fast
			}
			block.Advance(1)
			return processLinkLabelOpen(block, segment.Start+1, true, pc)
		}
		return nil
	}
	if line[0] == '[' {
		if fast := s.parseSimpleLink(block, false, pc); fast != nil {
			return fast
		}
		return processLinkLabelOpen(block, segment.Start, false, pc)
	}

	// line[0] == ']'
	last := popLastLinkLabelState(pc)
	if last == nil {
		return nil
	}
	block.Advance(1)

	// CommonMark spec says:
	//  > A link label can have at most 999 characters inside the square brackets.
	if linkLabelSpan(openLinkLabelStates(pc)) > 998 {
		last.toText()
		return nil
	}

	if !last.IsImage && last.hasLinkDescendant { // a link in a link text is not allowed
		last.toText()
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
			last.toText()
			return nil
		}
	}

	if link == nil {
		// maybe shortcut reference link
		block.SetPosition(l, pos)
		labelIndex := text.NewIndex(last.value().Stop, segment.Start)
		link = parseShortcutLinkTail(labelIndex, block, pc)
		if link == nil {
			last.toText()
			return nil
		}
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
	n.(interface{ SetPos(int) }).SetPos(last.value().Start)
	return n
}

// parseSimpleLink is the fast path for inline and reference links /
// images whose label is a single run of plain characters.
func (s *linkParser) parseSimpleLink(block text.Reader, isImage bool, pc Context) ast.Node {
	line, segment := block.PeekLine()
	contentStart := 1
	if isImage {
		contentStart = 2 // skip `![`
	}
	if contentStart >= len(line) {
		return nil
	}

	pctx, ok := pc.(*parseContext)
	if !ok || pctx.inlineParsers == nil {
		return nil
	}
	parsers := pctx.inlineParsers

	closeBracket := -1
	for i := contentStart; i < len(line); i++ {
		ch := line[i]
		if ch == '\\' && i+1 < len(line) {
			i++
			continue
		}
		if ch == '\n' {
			return nil
		}
		if ch == ']' {
			closeBracket = i
			break
		}
		if ch == '[' {
			return nil
		}
		if parsers[ch] != nil {
			return nil
		}
	}
	if closeBracket < 0 {
		return nil
	}

	// Commit the reader past `[label]` and delegate the tail to the
	// shared helpers. Anything that fails there rewinds and lets the
	// standard `[…]` state-machine take another swing.
	labelIndex := text.NewIndex(segment.Start+contentStart, segment.Start+closeBracket)
	savedLine, savedPos := block.Position()
	block.Advance(closeBracket + 1) // past `[label]`

	var link *ast.Link
	switch block.Peek() {
	case '(':
		link = parseInlineLinkTail(block)
	case '[':
		link, _ = parseReferenceLinkTail(block, labelIndex, pc)
	default:
		link = parseShortcutLinkTail(labelIndex, block, pc)
	}
	if link == nil {
		block.SetPosition(savedLine, savedPos)
		return nil
	}

	labelText := ast.NewText(text.NewSingleLineValueFromSegment(
		text.NewSegment(labelIndex.Start, labelIndex.Stop), block.Decoder()))
	link.AppendChild(labelText)
	link.SetPos(segment.Start)

	if !isImage {
		// CommonMark disallows links inside link text; propagate the
		// descendant marker so any still-open outer `[...]` opener
		// turns into literal text when its closer is reached.
		for _, st := range openLinkLabelStates(pc) {
			st.hasLinkDescendant = true
		}
		return link
	}
	img := ast.NewImage(link.Destination)
	img.Title = link.Title
	img.Reference = link.Reference
	link.RemoveChild(labelText)
	img.AppendChild(labelText)
	img.SetPos(segment.Start)
	return img
}

func processLinkLabelOpen(block text.Reader, pos int, isImage bool, pc Context) ast.Node {
	start := pos
	if isImage {
		start--
	}
	stack := linkLabelStateStackFor(pc)
	st := stack.newState()
	st.Value = text.NewSingleLineValueFromSegment(text.NewSegment(start, pos+1), block.Decoder())
	st.IsImage = isImage
	st.bottom = pc.LastDelimiter()
	stack.states = append(stack.states, st)
	block.Advance(1)
	return st
}

func (s *linkParser) processLinkLabel(parent ast.Node, link *ast.Link, last *linkLabelState, pc Context) {
	ProcessDelimiters(last.bottom, pc)
	for c := last.NextSibling(); c != nil; {
		next := c.NextSibling()
		parent.RemoveChild(c)
		link.AppendChild(c)
		c = next
	}
	if !last.IsImage {
		for _, st := range openLinkLabelStates(pc) {
			st.hasLinkDescendant = true
		}
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
				b.Decoder(r.Decoder())
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

func parseInlineLinkTail(block text.Reader) *ast.Link {
	block.Advance(1) // skip '('
	block.SkipSpaces()
	var title text.MultiLineValue
	var destination text.SingleLineValue
	var ok bool
	if block.Peek() == ')' {
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
			if block.Peek() != ')' {
				return nil
			}
			block.Advance(1)
		}
	}
	link := &ast.Link{Destination: destination, Title: title}
	link.Init(link)
	return link
}

func parseReferenceLinkTail(block text.Reader, labelIndex text.Index, pc Context) (*ast.Link, bool) {
	block.Advance(1) // skip '['
	maybeReferenceValue, found := findClosure(block, '[', ']')
	if !found {
		return nil, false
	}
	refType := ast.ReferenceLinkKindFull
	maybeReference := maybeReferenceValue.Bytes(block.Source())
	if util.IsBlank(maybeReference) { // collapsed
		maybeReference = block.ValueBetween(labelIndex.Start, labelIndex.Stop).Bytes(block.Source())
		refType = ast.ReferenceLinkKindCollapsed
	}
	if len(maybeReference) > 999 {
		return nil, true
	}
	def, ok := pc.LinkDefinition(util.ToLinkReference(maybeReference))
	if !ok {
		return nil, true
	}
	return referenceLink(def, refType, maybeReferenceValue, block.Decoder()), true
}

func parseShortcutLinkTail(labelIndex text.Index, block text.Reader, pc Context) *ast.Link {
	refValue := block.ValueBetween(labelIndex.Start, labelIndex.Stop)
	maybeReference := refValue.Bytes(block.Source())
	if len(maybeReference) == 0 || len(maybeReference) > 999 {
		return nil
	}
	def, ok := pc.LinkDefinition(util.ToLinkReference(maybeReference))
	if !ok {
		return nil
	}
	return referenceLink(def, ast.ReferenceLinkKindShortcut, refValue, block.Decoder())
}

func (s *linkParser) parseReferenceLink(parent ast.Node, last *linkLabelState,
	block text.Reader, pc Context) (*ast.Link, bool) {
	_, orgpos := block.Position()
	labelIndex := text.NewIndex(last.value().Stop, orgpos.Start-1)
	link, hasValue := parseReferenceLinkTail(block, labelIndex, pc)
	if link == nil {
		return nil, hasValue
	}
	s.processLinkLabel(parent, link, last, pc)
	return link, true
}

func (s *linkParser) parseLink(parent ast.Node, last *linkLabelState, block text.Reader, pc Context) *ast.Link {
	link := parseInlineLinkTail(block)
	if link == nil {
		return nil
	}
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
				return text.NewSingleLineValueFromIndex(text.NewIndex(segment.Start+1, segment.Start+i), block.Decoder()), true
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
	return text.NewSingleLineValueFromIndex(text.NewIndex(segment.Start, segment.Start+i), block.Decoder()), true
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

func referenceLink(def LinkDefinition, kind ast.ReferenceLinkKind, refvalue text.MultiLineValue,
	decoder text.Decoder) *ast.Link {
	link := &ast.Link{
		Reference: &ast.ReferenceLink{ReferenceLinkKind: kind, Value: refvalue},
	}
	link.Init(link)
	if ld, ok := def.(*linkDefinition); ok && ld.node != nil {
		link.Destination = ld.node.Destination
		link.Title = ld.node.Title
	} else {
		link.Destination = text.NewSingleLineValueFromString(string(def.Destination()), decoder)
		link.Title = text.NewMultiLineValueFromString(string(def.Title()), decoder)
	}
	return link
}

func (s *linkParser) CloseBlock(_ ast.Node, _ text.Reader, pc Context) {
	v := pc.Get(linkLabelStateKey)
	if v == nil {
		return
	}
	stack := v.(*linkLabelStateStack)
	states := stack.states
	if len(states) == 0 {
		return
	}
	for _, state := range states {
		state.toText()
	}
	for i := range states {
		states[i] = nil
	}
	stack.states = states[:0]
}
