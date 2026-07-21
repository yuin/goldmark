package parser

import (
	"bytes"

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
		attr, ok := parseAttribute(reader)
		if !ok {
			reader.SetPosition(savedLine, savedPosition)
			return nil, false
		}
		if attr.Name == "class" {
			updated := false
			for i, a := range attrs {
				if a.Name == "class" {
					existing := a.Value.Str(nil)
					newVal := attr.Value.Str(nil)
					attrs[i].Value = text.NewStringMultilineValue(existing + " " + newVal)
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
		for i < len(line) && (!util.IsSpace(line[i]) &&
			(!util.IsPunct(line[i]) || line[i] == '_' ||
				line[i] == '-' || line[i] == ':' || line[i] == '.')) {
			i++
		}
		name := "class"
		if c == '#' {
			name = "id"
		}
		reader.Advance(i)
		return gast.Attribute{
			Name:  name,
			Value: text.NewStringMultilineValue(string(line[0:i])),
		}, true
	}
	line, _ := reader.PeekLine()
	if len(line) == 0 {
		return gast.Attribute{}, false
	}
	c = line[0]
	if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
		c != '_' && c != ':' {
		return gast.Attribute{}, false
	}
	i := 0
	for ; i < len(line); i++ {
		c = line[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') &&
			c != '_' && c != ':' && c != '.' && c != '-' {
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
	// class values stay letter-initial words or quoted strings (CommonMark/XHTML-ish).
	// Other attributes also accept numbers and simple arrays (Hugo highlighter opts; #563).
	var value text.MultilineValue
	var ok bool
	if util.BytesToReadOnlyString(name) == "class" {
		c = reader.Peek()
		if c == '"' {
			value, ok = parseAttributeString(reader)
		} else {
			value, ok = parseAttributeWordStrict(reader)
		}
	} else {
		value, ok = parseAttributeValue(reader)
	}
	if !ok {
		return gast.Attribute{}, false
	}
	return gast.Attribute{
		Name:  util.BytesToReadOnlyString(name),
		Value: value,
	}, true
}

func parseAttributeValue(reader text.Reader) (text.MultilineValue, bool) {
	reader.SkipSpaces()
	c := reader.Peek()
	switch c {
	case text.EOF:
		return text.MultilineValue{}, false
	case '"':
		return parseAttributeString(reader)
	case '[':
		return parseAttributeArray(reader)
	default:
		// Accept unquoted words and numbers (e.g. linenostart=199).
		// Previously only letter-initial words were accepted, which dropped
		// numeric values used by tools such as Hugo (#563).
		return parseAttributeWord(reader)
	}
}

func parseAttributeString(reader text.Reader) (text.MultilineValue, bool) {
	reader.Advance(1) // skip "
	line, _ := reader.PeekLine()
	i := 0
	l := len(line)
	var buf bytes.Buffer
	for i < l {
		c := line[i]
		if c == '\\' && i != l-1 {
			n := line[i+1]
			switch n {
			case '"', '/', '\\':
				buf.WriteByte(n)
				i += 2
			case 'b':
				buf.WriteString("\b")
				i += 2
			case 'f':
				buf.WriteString("\f")
				i += 2
			case 'n':
				buf.WriteString("\n")
				i += 2
			case 'r':
				buf.WriteString("\r")
				i += 2
			case 't':
				buf.WriteString("\t")
				i += 2
			default:
				buf.WriteByte('\\')
				i++
			}
			continue
		}
		if c == '"' {
			reader.Advance(i + 1)
			return text.NewStringMultilineValue(buf.String()), true
		}
		buf.WriteByte(c)
		i++
	}
	return text.MultilineValue{}, false
}

func parseAttributeWord(reader text.Reader) (text.MultilineValue, bool) {
	line, _ := reader.PeekLine()
	if len(line) == 0 {
		return text.MultilineValue{}, false
	}
	c := line[0]
	// Allow letter-initial words and numeric / signed numeric values.
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		c == '_' || c == ':' || c == '+' || c == '-' || util.IsNumeric(c)) {
		return text.MultilineValue{}, false
	}
	i := 0
	for ; i < len(line); i++ {
		c := line[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') &&
			c != '_' && c != ':' && c != '.' && c != '-' && c != '+' {
			break
		}
	}
	if i == 0 {
		return text.MultilineValue{}, false
	}
	// Reject a lone sign with no digits.
	if i == 1 && (line[0] == '+' || line[0] == '-') {
		return text.MultilineValue{}, false
	}
	reader.Advance(i)
	return text.NewStringMultilineValue(string(line[:i])), true
}

// parseAttributeWordStrict is the historical word rule (letter/_/:-initial only).
func parseAttributeWordStrict(reader text.Reader) (text.MultilineValue, bool) {
	line, _ := reader.PeekLine()
	if len(line) == 0 {
		return text.MultilineValue{}, false
	}
	c := line[0]
	if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
		c != '_' && c != ':' {
		return text.MultilineValue{}, false
	}
	i := 0
	for ; i < len(line); i++ {
		c := line[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') &&
			c != '_' && c != ':' && c != '.' && c != '-' {
			break
		}
	}
	reader.Advance(i)
	return text.NewStringMultilineValue(string(line[:i])), true
}

// parseAttributeArray parses [v1, v2, ...] into a space-separated string value.
// Nested structured types are flattened to their string form because v2
// attribute values are MultilineValue strings.
func parseAttributeArray(reader text.Reader) (text.MultilineValue, bool) {
	reader.Advance(1) // skip [
	var parts []string
	for i := 0; ; i++ {
		reader.SkipSpaces()
		c := reader.Peek()
		if c == ']' {
			reader.Advance(1)
			return text.NewStringMultilineValue(joinAttributeArrayParts(parts)), true
		}
		if i != 0 {
			if c != ',' {
				return text.MultilineValue{}, false
			}
			reader.Advance(1)
			reader.SkipSpaces()
			if reader.Peek() == ']' { // trailing comma
				return text.MultilineValue{}, false
			}
		}
		value, ok := parseAttributeValue(reader)
		if !ok {
			return text.MultilineValue{}, false
		}
		parts = append(parts, value.Str(nil))
	}
}

func joinAttributeArrayParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for i, p := range parts {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(p)
	}
	return buf.String()
}
