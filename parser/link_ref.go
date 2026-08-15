package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type linkReferenceParagraphTransformer struct {
}

// LinkReferenceParagraphTransformer is a ParagraphTransformer implementation
// that parses and extracts link reference from paragraphs.
var LinkReferenceParagraphTransformer ParagraphTransformer = &linkReferenceParagraphTransformer{}

func (p *linkReferenceParagraphTransformer) Transform(node *ast.Paragraph, reader text.Reader, pc Context) {
	lines := node.Source()
	block := text.NewBlockReader(reader.Source(), lines, reader.Decoder())
	removes := [][2]int{}
	for {
		ref, start, end := parseLinkReferenceDefinition(block, pc)
		if start > -1 {
			if start == 0 {
				ref.SetBlankPreviousLines(node.HasBlankPreviousLines())
			}
			ref.SetPos(lines[start].Start)
			node.Parent().InsertBefore(node, ref)
			if start == end {
				end++
			}
			removes = append(removes, [2]int{start, end})
			continue
		}
		break
	}

	offset := 0
	for _, remove := range removes {
		if len(lines) == 0 {
			break
		}
		s := lines[remove[1]-offset:]
		lines = lines[0 : remove[0]-offset]
		lines = append(lines, s...)
		offset = remove[1]
	}

	if len(lines) == 0 {
		node.Parent().RemoveChild(node)
		return
	}

	node.SetSource(lines)
}

func parseLinkReferenceDefinition(block text.Reader, pc Context) (ast.BlockNode, int, int) {
	block.SkipSpaces()
	line, _ := block.PeekLine()
	if line == nil {
		return nil, -1, -1
	}
	startLine, _ := block.Position()
	width, pos := util.IndentWidth(line, 0)
	if width > 3 {
		return nil, -1, -1
	}
	if width != 0 {
		pos++
	}
	if line[pos] != '[' {
		return nil, -1, -1
	}
	block.Advance(pos + 1)
	labelVal, found := findClosure(block, '[', ']')
	if !found {
		return nil, -1, -1
	}
	if util.IsBlank(labelVal.Bytes(block.Source())) {
		return nil, -1, -1
	}
	if block.Peek() != ':' {
		return nil, -1, -1
	}
	block.Advance(1)
	block.SkipSpaces()
	destination, ok := parseLinkDestination(block)
	if !ok {
		return nil, -1, -1
	}
	line, _ = block.PeekLine()
	isNewLine := line == nil || util.IsBlank(line)

	endLine, _ := block.Position()
	_, spaces, _ := block.SkipSpaces()
	opener := block.Peek()
	if opener != '"' && opener != '\'' && opener != '(' {
		if !isNewLine {
			return nil, -1, -1
		}
		ref := ast.NewLinkReferenceDefinition(
			labelVal,
			destination,
		)
		pc.AddLinkDefinition(newLinkDefinitionFromNode(ref, block.Source()))
		return ref, startLine, endLine + 1
	}
	if spaces == 0 {
		return nil, -1, -1
	}
	block.Advance(1)
	closer := opener
	if opener == '(' {
		closer = ')'
	}
	titleVal, found := findClosure(block, opener, closer)
	if !found {
		if !isNewLine {
			return nil, -1, -1
		}
		ref := ast.NewLinkReferenceDefinition(
			labelVal,
			destination,
		)
		pc.AddLinkDefinition(newLinkDefinitionFromNode(ref, block.Source()))
		block.AdvanceLine()
		return ref, startLine, endLine + 1
	}

	line, _ = block.PeekLine()
	if line != nil && !util.IsBlank(line) {
		if !isNewLine {
			return nil, -1, -1
		}
		ref := ast.NewLinkReferenceDefinition(
			labelVal,
			destination,
			ast.WithLinkTitle(titleVal),
		)
		pc.AddLinkDefinition(newLinkDefinitionFromNode(ref, block.Source()))
		return ref, startLine, endLine
	}

	endLine, _ = block.Position()
	ref := ast.NewLinkReferenceDefinition(
		labelVal,
		destination,
		ast.WithLinkTitle(titleVal),
	)
	pc.AddLinkDefinition(newLinkDefinitionFromNode(ref, block.Source()))
	return ref, startLine, endLine + 1
}
