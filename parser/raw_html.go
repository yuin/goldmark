package parser

import (
	"bytes"
	"regexp"

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

func (s *rawHTMLParser) Parse(_ ast.Node, block text.Reader, pc Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) > 1 && util.IsAlphaNumeric(line[1]) {
		return s.parseMultiLineRegexp(openTagRegexp, block, pc)
	}
	if len(line) > 2 && line[1] == '/' && util.IsAlphaNumeric(line[2]) {
		return s.parseMultiLineRegexp(closeTagRegexp, block, pc)
	}
	if bytes.HasPrefix(line, openComment) {
		return s.parseComment(block, pc)
	}
	if bytes.HasPrefix(line, openProcessingInstruction) {
		return s.parseUntil(block, closeProcessingInstruction, pc)
	}
	if len(line) > 2 && line[1] == '!' && line[2] >= 'A' && line[2] <= 'Z' {
		return s.parseUntil(block, closeDecl, pc)
	}
	if bytes.HasPrefix(line, openCDATA) {
		return s.parseUntil(block, closeCDATA, pc)
	}
	return nil
}

var tagnamePattern = `([A-Za-z][A-Za-z0-9-]*)`
var spaceOrOneNewline = `(?:[ \t]|(?:\r\n|\n){0,1})`
var attributePattern = `(?:[\r\n \t]+[a-zA-Z_:][a-zA-Z0-9:._-]*(?:[\r\n \t]*=[\r\n \t]*(?:[^\"'=<>` + "`" + `\x00-\x20]+|'[^']*'|"[^"]*"))?)` //nolint:lll
var openTagRegexp = regexp.MustCompile("^<" + tagnamePattern + attributePattern + `*` + spaceOrOneNewline + `*/?>`)
var closeTagRegexp = regexp.MustCompile("^</" + tagnamePattern + spaceOrOneNewline + `*>`)

var openProcessingInstruction = []byte("<?")
var closeProcessingInstruction = []byte("?>")
var openCDATA = []byte("<![CDATA[")
var closeCDATA = []byte("]]>")
var closeDecl = []byte(">")
var emptyComment1 = []byte("<!-->")
var emptyComment2 = []byte("<!--->")
var openComment = []byte("<!--")
var closeComment = []byte("-->")

func (s *rawHTMLParser) parseComment(block text.Reader, _ Context) ast.Node {
	savedLine, savedSegment := block.Position()
	line, segment := block.PeekLine()
	if bytes.HasPrefix(line, emptyComment1) {
		stop := segment.Start + len(emptyComment1)
		block.Advance(len(emptyComment1))
		return ast.NewRawHTML(text.NewIndexMultiLineValue(text.NewIndex(segment.Start, stop)))
	}
	if bytes.HasPrefix(line, emptyComment2) {
		stop := segment.Start + len(emptyComment2)
		block.Advance(len(emptyComment2))
		return ast.NewRawHTML(text.NewIndexMultiLineValue(text.NewIndex(segment.Start, stop)))
	}
	offset := len(openComment)
	line = line[offset:]
	var indices []text.Index
	for {
		index := bytes.Index(line, closeComment)
		if index > -1 {
			stop := segment.Start + offset + index + len(closeComment)
			indices = append(indices, text.NewIndex(segment.Start, stop))
			block.Advance(offset + index + len(closeComment))
			return ast.NewRawHTML(text.NewIndicesMultiLineValue(indices))
		}
		offset = 0
		indices = append(indices, text.NewIndex(segment.Start, segment.Stop))
		block.AdvanceLine()
		line, segment = block.PeekLine()
		if line == nil {
			break
		}
	}
	block.SetPosition(savedLine, savedSegment)
	return nil
}

func (s *rawHTMLParser) parseUntil(block text.Reader, closer []byte, _ Context) ast.Node {
	savedLine, savedSegment := block.Position()
	var indices []text.Index
	for {
		line, segment := block.PeekLine()
		if line == nil {
			break
		}
		index := bytes.Index(line, closer)
		if index > -1 {
			stop := segment.Start + index + len(closer)
			indices = append(indices, text.NewIndex(segment.Start, stop))
			block.Advance(index + len(closer))
			return ast.NewRawHTML(text.NewIndicesMultiLineValue(indices))
		}
		indices = append(indices, text.NewIndex(segment.Start, segment.Stop))
		block.AdvanceLine()
	}
	block.SetPosition(savedLine, savedSegment)
	return nil
}

func (s *rawHTMLParser) parseMultiLineRegexp(reg *regexp.Regexp, block text.Reader, _ Context) ast.Node {
	sline, ssegment := block.Position()
	if block.Match(reg) {
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
		return ast.NewRawHTML(text.NewIndicesMultiLineValue(indices))
	}
	return nil
}
