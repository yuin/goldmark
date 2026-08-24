//go:build goldmark_v1_attribute

package parser

import (
	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var attrNameID = []byte("id")
var attrNameClass = []byte("class")

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
					attrs[i].Value = text.NewMultiLineValueFromString(existing+" "+newVal, text.IdentityDecoder)
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
		if reader.Peek() == ',' {
			reader.Advance(1)
			reader.SkipSpaces()
		}
	}
}

func parseAttribute(reader text.Reader) (gast.Attribute, bool) {
	reader.SkipSpaces()
	c := reader.Peek()
	if c == '#' || c == '.' {
		reader.Advance(1)
		line, _ := reader.PeekLine()
		i := 0
		// HTML5 allows any kind of characters as id, but XHTML restricts characters for id.
		// CommonMark is basically defined for XHTML(even though it is legacy).
		// So we restrict id characters.
		for ; i < len(line) && !util.IsSpace(line[i]) &&
			(!util.IsPunct(line[i]) || line[i] == '_' ||
				line[i] == '-' || line[i] == ':' || line[i] == '.'); i++ {
		}
		name := attrNameClass
		if c == '#' {
			name = attrNameID
		}
		reader.Advance(i)
		return gast.Attribute{
			Name:  util.BytesToReadOnlyString(name),
			Value: text.NewMultiLineValue(line[0:i], text.IdentityDecoder),
		}, true
	}
	line, _ := reader.PeekLine()
	if len(line) == 0 {
		return gast.Attribute{}, false
	}
	c = line[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		c == '_' || c == ':') {
		return gast.Attribute{}, false
	}
	i := 0
	for ; i < len(line); i++ {
		c = line[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' || c == ':' || c == '.' || c == '-') {
			break
		}
	}
	name := line[:i]
	reader.Advance(i)
	reader.SkipSpaces()
	c = reader.Peek()
	if c != '=' {
		return gast.Attribute{}, false
	}
	reader.Advance(1)
	reader.SkipSpaces()
	_, pos1 := reader.Position()
	_, ok := text.ParseAttributeValue(reader)
	if !ok {
		return gast.Attribute{}, false
	}
	_, pos2 := reader.Position()
	value := reader.ValueBetween(pos1.Start, pos2.Start)

	return gast.Attribute{
		Name:  util.BytesToReadOnlyString(name),
		Value: value,
	}, true
}
