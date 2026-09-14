package parser

import (
	"bytes"

	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func hasPrefixFold(b []byte, s string) bool {
	if len(b) < len(s) {
		return false
	}
	for i := range len(s) {
		c, d := b[i], s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if d >= 'A' && d <= 'Z' {
			d += 'a' - 'A'
		}
		if c != d {
			return false
		}
	}
	return true
}

func containsFold(b []byte, s string) bool {
	if len(s) == 0 {
		return true
	}
	for i := 0; i+len(s) <= len(b); i++ {
		if hasPrefixFold(b[i:], s) {
			return true
		}
	}
	return false
}

func isAttrSeparator(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func isAttrNameStartByte(c byte) bool {
	return isAlpha(c) || c == '_' || c == ':'
}

func isAttrNameByte(c byte) bool {
	return util.IsAlphaNumeric(c) || c == ':' || c == '.' || c == '_' || c == '-'
}

func isUnquotedAttrValueByte(c byte) bool {
	return c > 0x20 && c < 0xff && c != '"' && c != '\'' && c != '=' && c != '<' && c != '>' && c != '`'
}

func scanTagNameBytes(line []byte, pos int) (end int, ok bool) {
	if pos >= len(line) || !isAlpha(line[pos]) {
		return pos, false
	}
	end = pos + 1
	for end < len(line) && (util.IsAlphaNumeric(line[end]) || line[end] == '-') {
		end++
	}
	return end, true
}

func scanHTMLAttributesBytes(line []byte, pos int) (end int, hasAttr bool) {
	for {
		i := pos
		for i < len(line) && isAttrSeparator(line[i]) {
			i++
		}
		if i == pos || i >= len(line) || !isAttrNameStartByte(line[i]) {
			return pos, hasAttr
		}
		i++
		for i < len(line) && isAttrNameByte(line[i]) {
			i++
		}
		j := i
		for j < len(line) && isAttrSeparator(line[j]) {
			j++
		}
		if j >= len(line) || line[j] != '=' {
			pos, hasAttr = i, true
			continue
		}
		j++
		for j < len(line) && isAttrSeparator(line[j]) {
			j++
		}
		if j >= len(line) {
			// No value follows '=': the whole "=value" part is optional in
			// attributePattern, so back off to the attribute name only.
			return i, true
		}
		switch q := line[j]; {
		case q == '\'' || q == '"':
			k := j + 1
			for k < len(line) && line[k] != q {
				k++
			}
			if k >= len(line) {
				return i, true // unterminated quote: back off to name only.
			}
			pos, hasAttr = k+1, true
		case isUnquotedAttrValueByte(q):
			k := j + 1
			for k < len(line) && isUnquotedAttrValueByte(line[k]) {
				k++
			}
			pos, hasAttr = k, true
		default:
			return i, true // invalid value start: back off to name only.
		}
	}
}

func skipAttrNameBytesReader(r text.Reader) {
	line, _ := r.PeekLine()
	i := 0
	for i < len(line) && isAttrNameByte(line[i]) {
		i++
	}
	if i > 0 {
		r.Advance(i)
	}
}

func skipUnquotedAttrValueBytesReader(r text.Reader) {
	line, _ := r.PeekLine()
	i := 0
	for i < len(line) && isUnquotedAttrValueByte(line[i]) {
		i++
	}
	if i > 0 {
		r.Advance(i)
	}
}

func skipAttrSeparatorBytesReader(r text.Reader) bool {
	consumed := false
	for {
		line, _ := r.PeekLine()
		if line == nil {
			return consumed
		}
		i := 0
		for i < len(line) && isAttrSeparator(line[i]) {
			i++
		}
		if i > 0 {
			r.Advance(i)
			consumed = true
		}
		if i < len(line) {
			return consumed
		}
	}
}

func skipToClosingQuoteBytesReader(r text.Reader, q byte) bool {
	for {
		line, _ := r.PeekLine()
		if line == nil {
			return false
		}
		i := bytes.IndexByte(line, q)
		if i >= 0 {
			r.Advance(i)
			return true
		}
		r.Advance(len(line))
	}
}

func scanTagNameReader(r text.Reader) bool {
	line, _ := r.PeekLine()
	end, ok := scanTagNameBytes(line, 0)
	if !ok {
		return false
	}
	r.Advance(end)
	return true
}

func skipAttrSeparatorsReader(r text.Reader) {
	for {
		line, _ := r.PeekLine()
		if line == nil {
			return
		}
		i := 0
		for i < len(line) {
			c := line[i]
			if c != ' ' && c != '\t' && c != '\n' {
				break
			}
			i++
		}
		if i > 0 {
			r.Advance(i)
		}
		if i >= len(line) {
			continue
		}
		if line[i] != '\r' {
			return
		}
		savedLine, savedPos := r.Position()
		r.Advance(1)
		if r.Peek() != '\n' {
			r.SetPosition(savedLine, savedPos)
			return
		}
		r.Advance(1)
	}
}

func scanHTMLAttributesReader(r text.Reader) {
	for {
		hasSep := skipAttrSeparatorBytesReader(r)
		if !hasSep || !isAttrNameStartByte(r.Peek()) {
			return
		}
		r.Advance(1)
		skipAttrNameBytesReader(r)
		nameEndLine, nameEndPos := r.Position()
		skipAttrSeparatorBytesReader(r)
		if r.Peek() != '=' {
			r.SetPosition(nameEndLine, nameEndPos)
			continue
		}
		r.Advance(1)
		skipAttrSeparatorBytesReader(r)
		switch q := r.Peek(); {
		case q == '\'' || q == '"':
			r.Advance(1)
			if !skipToClosingQuoteBytesReader(r, q) {
				r.SetPosition(nameEndLine, nameEndPos)
				continue
			}
			r.Advance(1)
		case isUnquotedAttrValueByte(q):
			r.Advance(1)
			skipUnquotedAttrValueBytesReader(r)
		default:
			r.SetPosition(nameEndLine, nameEndPos)
		}
	}
}
