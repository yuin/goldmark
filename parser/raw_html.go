package parser

import (
	"bytes"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type rawHTMLParser struct {
}

var defaultRawHTMLParser = &rawHTMLParser{}

// NewRawHTMLParser return a new InlineParser that can parse
// inline htmls.
func NewRawHTMLParser() InlineParser {
	return defaultRawHTMLParser
}

func (s *rawHTMLParser) Trigger() []byte {
	return []byte{'<'}
}

var closeProcessingInstructionCacheKey = NewContextKey()
var closeDeclCacheKey = NewContextKey()
var closeCDATACacheKey = NewContextKey()
var closeCommentCacheKey = NewContextKey()

func (s *rawHTMLParser) Parse(_ ast.Node, block text.Reader, pc Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) > 1 && util.IsAlphaNumeric(line[1]) {
		return s.parseTag(scanOpenTag, block, pc)
	}
	if len(line) > 2 && line[1] == '/' && util.IsAlphaNumeric(line[2]) {
		return s.parseTag(scanCloseTag, block, pc)
	}
	if bytes.HasPrefix(line, openComment) {
		return s.parseComment(block, pc)
	}
	if bytes.HasPrefix(line, openProcessingInstruction) {
		return s.parseUntil(block, closeProcessingInstruction, closeProcessingInstructionCacheKey, pc)
	}
	if len(line) > 2 && line[1] == '!' && line[2] >= 'A' && line[2] <= 'Z' {
		return s.parseUntil(block, closeDecl, closeDeclCacheKey, pc)
	}
	if bytes.HasPrefix(line, openCDATA) {
		return s.parseUntil(block, closeCDATA, closeCDATACacheKey, pc)
	}
	return nil
}

func scanOpenTag(r text.Reader) bool {
	line, pos := r.Position()
	if r.Peek() != '<' {
		return false
	}
	r.Advance(1)
	if !scanTagNameReader(r) {
		r.SetPosition(line, pos)
		return false
	}
	scanHTMLAttributesReader(r)
	skipAttrSeparatorsReader(r)
	if r.Peek() == '/' {
		r.Advance(1)
	}
	if r.Peek() != '>' {
		r.SetPosition(line, pos)
		return false
	}
	r.Advance(1)
	return true
}

func scanCloseTag(r text.Reader) bool {
	line, pos := r.Position()
	if r.Peek() != '<' {
		return false
	}
	r.Advance(1)
	if r.Peek() != '/' {
		r.SetPosition(line, pos)
		return false
	}
	r.Advance(1)
	if !scanTagNameReader(r) {
		r.SetPosition(line, pos)
		return false
	}
	skipAttrSeparatorsReader(r)
	if r.Peek() != '>' {
		r.SetPosition(line, pos)
		return false
	}
	r.Advance(1)
	return true
}

var openProcessingInstruction = []byte("<?")
var closeProcessingInstruction = []byte("?>")
var openCDATA = []byte("<![CDATA[")
var closeCDATA = []byte("]]>")
var closeDecl = []byte(">")
var emptyComment1 = []byte("<!-->")
var emptyComment2 = []byte("<!--->")
var openComment = []byte("<!--")
var closeComment = []byte("-->")

func (s *rawHTMLParser) parseComment(block text.Reader, pctx Context) ast.Node {
	savedLine, savedSegment := block.Position()
	line, segment := block.PeekLine()
	if bytes.HasPrefix(line, emptyComment1) {
		stop := segment.Start + len(emptyComment1)
		block.Advance(len(emptyComment1))
		return ast.NewRawHTML(text.NewMultiLineValueFromIndex(text.NewIndex(segment.Start, stop), text.IdentityDecoder))
	}
	if bytes.HasPrefix(line, emptyComment2) {
		stop := segment.Start + len(emptyComment2)
		block.Advance(len(emptyComment2))
		return ast.NewRawHTML(text.NewMultiLineValueFromIndex(text.NewIndex(segment.Start, stop), text.IdentityDecoder))
	}
	cache := pctx.ComputeIfAbsent(closeCommentCacheKey, func() any { return &posCache{} }).(*posCache)
	if segment.Start < cache.pos {
		return nil
	}
	offset := len(openComment)
	line = line[offset:]
	var indices []text.Index
	var lastStop int
	for {
		index := bytes.Index(line, closeComment)
		if index > -1 {
			stop := segment.Start + offset + index + len(closeComment)
			indices = append(indices, text.NewIndex(segment.Start, stop))
			block.Advance(offset + index + len(closeComment))
			return ast.NewRawHTML(text.NewMultiLineValueFromIndices(indices, text.IdentityDecoder))
		}
		offset = 0
		indices = append(indices, text.NewIndex(segment.Start, segment.Stop))
		lastStop = segment.Stop
		block.AdvanceLine()
		line, segment = block.PeekLine()
		if line == nil {
			break
		}
	}
	cache.pos = lastStop
	block.SetPosition(savedLine, savedSegment)
	return nil
}

func (s *rawHTMLParser) parseUntil(block text.Reader, closer []byte, ck ContextKey, pctx Context) ast.Node {
	savedLine, savedSegment := block.Position()
	var indices []text.Index
	cache := pctx.ComputeIfAbsent(ck, func() any { return &posCache{} }).(*posCache)
	var lastStop int
	for {
		line, segment := block.PeekLine()
		if line == nil {
			break
		}
		if segment.Start < cache.pos {
			break
		}
		index := bytes.Index(line, closer)
		if index > -1 {
			stop := segment.Start + index + len(closer)
			indices = append(indices, text.NewIndex(segment.Start, stop))
			block.Advance(index + len(closer))
			return ast.NewRawHTML(text.NewMultiLineValueFromIndices(indices, text.IdentityDecoder))
		}
		indices = append(indices, text.NewIndex(segment.Start, segment.Stop))
		lastStop = segment.Stop
		block.AdvanceLine()
	}
	cache.pos = max(cache.pos, lastStop)
	block.SetPosition(savedLine, savedSegment)
	return nil
}

func (s *rawHTMLParser) parseTag(scan func(text.Reader) bool, block text.Reader, _ Context) ast.Node {
	sline, ssegment := block.Position()
	if scan(block) {
		eline, esegment := block.Position()
		block.SetPosition(sline, ssegment)
		var indices []text.Index
		for {
			line, segment := block.PeekLine()
			if line == nil {
				break
			}
			l, _ := block.Position()
			start := segment.Start
			if l == sline {
				start = ssegment.Start
			}
			end := segment.Stop
			if l == eline {
				end = esegment.Start
			}
			indices = append(indices, text.NewIndex(start, end))
			if l == eline {
				block.Advance(end - start)
				break
			}
			block.AdvanceLine()
		}
		return ast.NewRawHTML(text.NewMultiLineValueFromIndices(indices, text.IdentityDecoder))
	}
	return nil
}
