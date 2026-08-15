//go:build !goldmark_v1_attribute

package parser

import (
	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// ParseAttributes parses attributes into a slice of ast.Attribute.
// ParseAttributes returns a parsed attributes and true if could parse
// attributes, otherwise nil and false.
func ParseAttributes(reader text.Reader) ([]gast.Attribute, bool) {
	savedLine, savedPosition := reader.Position()
	reader.SkipSpaces()
	if reader.Peek() != '{' {
		reader.SetPosition(savedLine, savedPosition)
		return nil, false
	}
	reader.Advance(1)
	var attrs []gast.Attribute
	for {
		if reader.Peek() == '}' {
			reader.Advance(1)
			return attrs, true
		}
		if reader.Peek() == text.EOF {
			return nil, false
		}
		attr, ok := parseAttribute(reader)
		if !ok {
			reader.SetPosition(savedLine, savedPosition)
			return nil, false
		}
		if attr.Name == "class" {
			updated := false
			for i, a := range attrs {
				if a.Name == "class" {
					existing := a.Value.Str(reader.Source())
					newVal := attr.Value.Str(reader.Source())
					attrs[i].Value = text.NewMultiLineValueFromString(existing+" "+newVal, reader.Decoder())
					updated = true
					break
				}
			}
			if !updated {
				attrs = append(attrs, attr)
			}
		} else {
			attrs = append(attrs, attr)
		}
		reader.SkipSpaces()
	}
}

func parseAttribute(reader text.Reader) (gast.Attribute, bool) {
	reader.SkipSpaces()
	c := reader.Peek()
	if c == '#' || c == '.' {
		reader.Advance(1)
		line, seg := reader.PeekLine()
		i := 0
		for i < len(line) && !util.IsSpace(line[i]) &&
			(!util.IsPunct(line[i]) || line[i] == '_' ||
				line[i] == '-' || line[i] == ':' || line[i] == '.') {
			i++
		}
		if i == 0 {
			return gast.Attribute{}, false
		}
		name := "class"
		if c == '#' {
			name = "id"
		}
		reader.Advance(i)
		return gast.Attribute{
			Name:  name,
			Value: text.NewMultiLineValueFromIndex(text.NewIndex(seg.Start, seg.Start+i), reader.Decoder()),
		}, true
	}
	line, seg := reader.PeekLine()
	if len(line) == 0 {
		return gast.Attribute{}, false
	}
	c = line[0]
	if util.IsSpace(c) || c == '=' || c == '/' || c == '}' {
		return gast.Attribute{}, false
	}
	i := 0
	for ; i < len(line); i++ {
		c = line[i]
		if util.IsSpace(c) || c == '=' || c == '/' || c == '}' {
			break
		}
	}
	name := line[:i]
	reader.Advance(i)
	reader.SkipSpaces()
	c = reader.Peek()
	if c != '=' {
		return gast.Attribute{
			Name:  util.BytesToReadOnlyString(name),
			Value: text.NewMultiLineValueFromIndex(text.NewIndex(seg.Start, seg.Start+i), reader.Decoder()),
		}, true
	}
	reader.Advance(1)
	reader.SkipSpaces()
	value, ok := parseAttributeValue(reader)
	if !ok {
		return gast.Attribute{}, false
	}
	return gast.Attribute{
		Name:  util.BytesToReadOnlyString(name),
		Value: value,
	}, true
}

func parseAttributeValue(reader text.Reader) (text.MultiLineValue, bool) {
	reader.SkipSpaces()
	c := reader.Peek()
	switch c {
	case text.EOF:
		return text.MultiLineValue{}, false
	case '"':
		return parseAttributeQuoted(reader, '"')
	case '\'':
		return parseAttributeQuoted(reader, '\'')
	default:
		return parseAttributeUnquoted(reader)
	}
}

func parseAttributeQuoted(reader text.Reader, q byte) (text.MultiLineValue, bool) {
	reader.Advance(1) // skip "/'
	var lines []text.Segment
	for {
		line, seg := reader.PeekLine()
		if len(line) == 0 {
			break
		}

		i := 0
		offset := 0
		for ; i < len(line); i++ {
			c := line[i]
			if c == q {
				offset = 1
				break
			}
		}

		reader.Advance(i + offset)
		lines = append(lines, seg.WithStop(seg.Start+i))
		if offset == 1 {
			var b text.ValueBuilder
			b.Decoder(reader.Decoder())
			for _, line := range lines {
				b.AddSegment(line)
			}
			return b.BuildMultiLine(), true
		}
	}
	return text.MultiLineValue{}, false
}

func parseAttributeUnquoted(reader text.Reader) (text.MultiLineValue, bool) {
	line, seg := reader.PeekLine()
	i := 0
	for ; i < len(line); i++ {
		c := line[i]
		if util.IsSpace(c) || c == '}' {
			break
		}
		// ", \, >, <, =, `, and ' are not allowed in unquoted HTML attribute values.
		// But most of implementations ignore this rule, so we also ignore it.
	}
	reader.Advance(i)
	return text.NewMultiLineValueFromIndex(text.NewIndex(seg.Start, seg.Start+i), reader.Decoder()), true
}
