// Package util provides utility functions for the goldmark.
package util

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// IsBlank returns true if given string is all space characters.
func IsBlank(bs []byte) bool {
	for _, b := range bs {
		if IsSpace(b) {
			continue
		}
		return false
	}
	return true
}

// DedentPosition dedents lines by given width.
func DedentPosition(bs []byte, width int) (pos, padding int) {
	i := 0
	l := len(bs)
	w := 0
	for ; i < l && w < width; i++ {
		b := bs[i]
		if b == ' ' {
			w++
		} else if b == '\t' {
			w += 4
		} else {
			break
		}
	}
	padding = w - width
	if padding < 0 {
		padding = 0
	}
	return i, padding
}

// VisualizeSpaces visualize invisible space characters.
func VisualizeSpaces(bs []byte) []byte {
	bs = bytes.Replace(bs, []byte(" "), []byte("[SPACE]"), -1)
	bs = bytes.Replace(bs, []byte("\t"), []byte("[TAB]"), -1)
	bs = bytes.Replace(bs, []byte("\n"), []byte("[NEWLINE]\n"), -1)
	return bs
}

// TabWidth calculates actual width of a tab at given position.
func TabWidth(currentPos int) int {
	return 4 - currentPos%4
}

// IndentPosition searches an indent position with given width for given line.
// If the line contains tab characters, paddings may be not zero.
// currentPos==0 and width==2:
//
//     position: 0    1
//               [TAB]aaaa
//     width:    1234 5678
//
// width=2 is in the tab character. In this case, IndentPosition returns
// (pos=1, padding=2)
func IndentPosition(bs []byte, currentPos, width int) (pos, padding int) {
	w := 0
	l := len(bs)
	for i := 0; i < l; i++ {
		b := bs[i]
		if b == ' ' {
			w++
		} else if b == '\t' {
			w += TabWidth(currentPos + w)
		} else {
			break
		}
		if w >= width {
			return i + 1, w - width
		}
	}
	return -1, -1
}

// IndentWidth calculate an indent width for given line.
func IndentWidth(bs []byte, currentPos int) (width, pos int) {
	l := len(bs)
	for i := 0; i < l; i++ {
		b := bs[i]
		if b == ' ' {
			width++
			pos++
		} else if b == '\t' {
			width += TabWidth(currentPos + width)
			pos++
		} else {
			break
		}
	}
	return
}

// FirstNonSpacePosition returns a potisoin line that is a first nonspace
// character.
func FirstNonSpacePosition(bs []byte) int {
	i := 0
	for ; i < len(bs); i++ {
		c := bs[i]
		if c == ' ' || c == '\t' {
			continue
		}
		if c == '\n' {
			return -1
		}
		return i
	}
	return -1
}

// FindClosure returns a position that closes given opener.
// If codeSpan is set true, it ignores characters in code spans.
// If allowNesting is set true, closures correspond to nested opener will be
// ignored.
func FindClosure(bs []byte, opener, closure byte, codeSpan, allowNesting bool) int {
	i := 0
	opened := 1
	codeSpanOpener := 0
	for i < len(bs) {
		c := bs[i]
		if codeSpan && codeSpanOpener != 0 && c == '`' {
			codeSpanCloser := 0
			for ; i < len(bs); i++ {
				if bs[i] == '`' {
					codeSpanCloser++
				} else {
					break
				}
			}
			if codeSpanCloser == codeSpanOpener {
				codeSpanOpener = 0
			}
		} else if c == '\\' && i < len(bs)-1 && IsPunct(bs[i+1]) {
			i += 2
			continue
		} else if codeSpan && codeSpanOpener == 0 && c == '`' {
			for ; i < len(bs); i++ {
				if bs[i] == '`' {
					codeSpanOpener++
				} else {
					break
				}
			}
		} else if (codeSpan && codeSpanOpener == 0) || !codeSpan {
			if c == closure {
				opened--
				if opened == 0 {
					return i
				}
			} else if c == opener {
				if !allowNesting {
					return -1
				}
				opened++
			}
		}
		i++
	}
	return -1
}

// TrimLeft trims characters in given s from head of the source.
// bytes.TrimLeft offers same functionalities, but bytes.TrimLeft
// allocates new buffer for the result.
func TrimLeft(source, b []byte) []byte {
	i := 0
	for ; i < len(source); i++ {
		c := source[i]
		found := false
		for j := 0; j < len(b); j++ {
			if c == b[j] {
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return source[i:]
}

// TrimRight trims characters in given s from tail of the source.
func TrimRight(source, b []byte) []byte {
	i := len(source) - 1
	for ; i >= 0; i-- {
		c := source[i]
		found := false
		for j := 0; j < len(b); j++ {
			if c == b[j] {
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return source[:i+1]
}

// TrimLeftLength returns a length of leading specified characters.
func TrimLeftLength(source, s []byte) int {
	return len(source) - len(TrimLeft(source, s))
}

// TrimRightLength returns a length of trailing specified characters.
func TrimRightLength(source, s []byte) int {
	return len(source) - len(TrimRight(source, s))
}

// TrimLeftSpaceLength returns a length of leading space characters.
func TrimLeftSpaceLength(source []byte) int {
	return TrimLeftLength(source, spaces)
}

// TrimRightSpaceLength returns a length of trailing space characters.
func TrimRightSpaceLength(source []byte) int {
	return TrimRightLength(source, spaces)
}

// TrimLeftSpace returns a subslice of given string by slicing off all leading
// space characters.
func TrimLeftSpace(source []byte) []byte {
	return TrimLeft(source, spaces)
}

// TrimRightSpace returns a subslice of given string by slicing off all trailing
// space characters.
func TrimRightSpace(source []byte) []byte {
	return TrimRight(source, spaces)
}

// ReplaceSpaces replaces sequence of spaces with given repl.
func ReplaceSpaces(source []byte, repl byte) []byte {
	var ret []byte
	start := -1
	for i, c := range source {
		iss := IsSpace(c)
		if start < 0 && iss {
			start = i
			continue
		} else if start >= 0 && iss {
			continue
		} else if start >= 0 {
			if ret == nil {
				ret = make([]byte, 0, len(source))
				ret = append(ret, source[:start]...)
			}
			ret = append(ret, repl)
			start = -1
		}
		if ret != nil {
			ret = append(ret, c)
		}
	}
	if start >= 0 && ret != nil {
		ret = append(ret, repl)
	}
	if ret == nil {
		return source
	}
	return ret
}

// ToRune decode given bytes start at pos and returns a rune.
func ToRune(source []byte, pos int) rune {
	i := pos
	for ; i >= 0; i-- {
		if utf8.RuneStart(source[i]) {
			break
		}
	}
	r, _ := utf8.DecodeRune(source[i:])
	return r
}

// ToValidRune returns 0xFFFD if given rune is invalid, otherwise v.
func ToValidRune(v rune) rune {
	if v == 0 || !utf8.ValidRune(v) {
		return rune(0xFFFD)
	}
	return v
}

// ToLinkReference convert given bytes into a valid link reference string.
// ToLinkReference trims leading and trailing spaces and convert into lower
// case and replace spaces with a single space character.
func ToLinkReference(v []byte) string {
	v = TrimLeftSpace(v)
	v = TrimRightSpace(v)
	return strings.ToLower(string(ReplaceSpaces(v, ' ')))
}

var escapeRegex = regexp.MustCompile(`\\.`)
var hexRefRegex = regexp.MustCompile(`#[xX][\da-fA-F]+;`)
var numRefRegex = regexp.MustCompile(`#\d{1,7};`)
var entityRefRegex = regexp.MustCompile(`&([a-zA-Z\d]+);`)

var entityLt = []byte("&lt;")
var entityGt = []byte("&gt;")
var entityAmp = []byte("&amp;")
var entityQuot = []byte("&quot;")

// EscapeHTML escapes characters that should be escaped in HTML text.
func EscapeHTML(v []byte) []byte {
	result := make([]byte, 0, len(v)+10)
	for _, c := range v {
		switch c {
		case '<':
			result = append(result, entityLt...)
		case '>':
			result = append(result, entityGt...)
		case '&':
			result = append(result, entityAmp...)
		case '"':
			result = append(result, entityQuot...)
		default:
			result = append(result, c)
		}
	}
	return result
}

// UnescapePunctuations unescapes blackslash escaped punctuations.
func UnescapePunctuations(v []byte) []byte {
	return escapeRegex.ReplaceAllFunc(v, func(match []byte) []byte {
		if IsPunct(match[1]) {
			return []byte{match[1]}
		}
		return match
	})
}

// ResolveNumericReferences resolve numeric references like '&#1234;" .
func ResolveNumericReferences(v []byte) []byte {
	buf := make([]byte, 6, 6)
	v = hexRefRegex.ReplaceAllFunc(v, func(match []byte) []byte {
		v, _ := strconv.ParseUint(string(match[2:len(match)-1]), 16, 32)
		n := utf8.EncodeRune(buf, ToValidRune(rune(v)))
		return buf[:n]
	})
	return numRefRegex.ReplaceAllFunc(v, func(match []byte) []byte {
		v, _ := strconv.ParseUint(string(match[1:len(match)-1]), 0, 32)
		n := utf8.EncodeRune(buf, ToValidRune(rune(v)))
		return buf[:n]
	})
}

// ResolveEntityNames resolve entity references like '&ouml;" .
func ResolveEntityNames(v []byte) []byte {
	return entityRefRegex.ReplaceAllFunc(v, func(match []byte) []byte {
		entity, ok := LookUpHTML5EntityByName(string(match[1 : len(match)-1]))
		if ok {
			return entity.Characters
		}
		return match
	})
}

// URLEscape escape given URL.
// If resolveReference is set true:
//   1. unescape punctuations
//   2. resolve numeric references
//   3. resolve entity references
//
// URL encoded values (%xx) are keeped as is.
func URLEscape(v []byte, resolveReference bool) []byte {
	if resolveReference {
		v = UnescapePunctuations(v)
		v = ResolveNumericReferences(v)
		v = ResolveEntityNames(v)
	}
	result := make([]byte, 0, len(v)+10)
	for i := 0; i < len(v); {
		c := v[i]
		if urlEscapeTable[c] == 1 {
			result = append(result, c)
			i++
			continue
		}
		if c == '%' && i+2 < len(v) && IsHexDecimal(v[i+1]) && IsHexDecimal(v[i+1]) {
			result = append(result, c, v[i+1], v[i+2])
			i += 3
			continue
		}
		u8len := utf8lenTable[c]
		if u8len == 99 { // invalid utf8 leading byte, skip it
			result = append(result, c)
			i++
			continue
		}
		if c == ' ' {
			result = append(result, '%', '2', '0')
			i++
			continue
		}
		result = append(result, []byte(url.QueryEscape(string(v[i:i+int(u8len)])))...)
		i += int(u8len)
	}
	return result
}

// GenerateLinkID generates an ID for links.
func GenerateLinkID(value []byte, exists map[string]bool) []byte {
	value = TrimLeftSpace(value)
	value = TrimRightSpace(value)
	result := []byte{}
	for i := 0; i < len(value); {
		v := value[i]
		l := utf8lenTable[v]
		i += int(l)
		if l != 1 {
			continue
		}
		if IsAlphaNumeric(v) {
			result = append(result, v)
		} else if v == ' ' {
			result = append(result, '-')
		}
	}
	if len(result) == 0 {
		result = []byte("id")
	}
	if _, ok := exists[string(result)]; !ok {
		exists[string(result)] = true
		return result
	}
	for i := 1; ; i++ {
		newResult := fmt.Sprintf("%s%d", result, i)
		if _, ok := exists[newResult]; !ok {
			exists[newResult] = true
			return []byte(newResult)
		}

	}
}

var spaces = []byte(" \t\n\x0b\x0c\x0d")

var spaceTable = [256]int8{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var punctTable = [256]int8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// a-zA-Z0-9, ;/?:@&=+$,-_.!~*'()#
var urlEscapeTable = [256]int8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var utf8lenTable = [256]int8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 99, 99, 99, 99, 99, 99, 99, 99}

// IsPunct returns true if given character is a punctuation, otherwise false.
func IsPunct(c byte) bool {
	return punctTable[c] == 1
}

// IsSpace returns true if given character is a space, otherwise false.
func IsSpace(c byte) bool {
	return spaceTable[c] == 1
}

// IsNumeric returns true if given character is a numeric, otherwise false.
func IsNumeric(c byte) bool {
	return c >= '0' && c <= '9'
}

// IsHexDecimal returns true if given character is a hexdecimal, otherwise false.
func IsHexDecimal(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// IsAlphaNumeric returns true if given character is a alphabet or a numeric, otherwise false.
func IsAlphaNumeric(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// A BufWriter is a subset of the bufio.Writer .
type BufWriter interface {
	io.Writer
	Available() int
	Buffered() int
	Flush() error
	WriteByte(c byte) error
	WriteRune(r rune) (size int, err error)
	WriteString(s string) (int, error)
}

// A PrioritizedValue struct holds pair of an arbitary value and a priority.
type PrioritizedValue struct {
	// Value is an arbitary value that you want to prioritize.
	Value interface{}
	// Priority is a priority of the value.
	Priority int
}

// PrioritizedSlice is a slice of the PrioritizedValues
type PrioritizedSlice []PrioritizedValue

// Sort sorts the PrioritizedSlice in ascending order.
func (s PrioritizedSlice) Sort() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Priority < s[j].Priority
	})
}

// Remove removes given value from this slice.
func (s PrioritizedSlice) Remove(v interface{}) PrioritizedSlice {
	i := 0
	found := false
	for ; i < len(s); i++ {
		if s[i].Value == v {
			found = true
			break
		}
	}
	if !found {
		return s
	}
	return append(s[:i], s[i+1:]...)
}

// Prioritized returns a new PrioritizedValue.
func Prioritized(v interface{}, priority int) PrioritizedValue {
	return PrioritizedValue{v, priority}
}
