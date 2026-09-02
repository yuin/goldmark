package text

import (
	"bytes"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/yuin/goldmark/v2/util"
)

// A Decoder decodes a byte slice, e.g. resolving backslash escapes and
// character references, and returns or writes out the decoded bytes.
//
// Note that this interface is not intended to 'normalize' a byte slice, but to decode it.
// For example, CommonMark requires that code spans normalize leading spaces, trailing spaces
// and newlines, but this interface does not do that.
type Decoder interface {
	// Decode decodes the given byte slice and returns the decoded bytes.
	Decode(b []byte) []byte

	// DecodeTo decodes the given byte slice and writes the decoded bytes to w.
	DecodeTo(w io.Writer, b []byte) (int, error)
}

type decoderConfig struct {
	EscapedSpace bool
}

// DecoderOption is a function that configures a decoder.
type DecoderOption func(*decoderConfig)

// WithEscapedSpace configures the decoder to treat escaped spaces as an empty character.
func WithEscapedSpace() DecoderOption {
	return func(cfg *decoderConfig) {
		cfg.EscapedSpace = true
	}
}

var _ Decoder = (*DefaultDecoder)(nil)

// DefaultDecoder is the default implementation of the Decoder interface.
//
// CommonMark defines texts except some inline elements (e.g., code span, auto link, etc)
// can contain character references and backslash escapes. This decoder decodes them.
//
// - decodes entity references (e.g., `&amp; -> &`)
// - decodes numeric character references (e.g., `&#x26; -> &`)
// - decodes `\` escaped punctuations (e.g., `\* -> *`)
//   - If WithEscapedSpace is set, it will also decode `\ ` to an empty character.
type DefaultDecoder struct {
	cfg decoderConfig
}

// NewDecoder creates a new Decoder with the given options.
func NewDecoder(opts ...DecoderOption) *DefaultDecoder {
	cfg := decoderConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &DefaultDecoder{cfg: cfg}
}

// indexByteFrom returns the index of the first occurrence of c in b at or
// after from, or -1 if c is not present in b[from:]. It is used to keep the
// cached '&'/'\\' cursors in Decode/DecodeTo up to date without rescanning
// bytes that have already been consumed.
func indexByteFrom(b []byte, from int, c byte) int {
	if from >= len(b) {
		return -1
	}
	if idx := bytes.IndexByte(b[from:], c); idx != -1 {
		return from + idx
	}
	return -1
}

// Decode implements the [Decoder] interface.
func (d *DefaultDecoder) Decode(b []byte) []byte {
	limit := len(b)
	ampPos := bytes.IndexByte(b, '&')
	bsPos := bytes.IndexByte(b, '\\')
	if ampPos == -1 && bsPos == -1 {
		return b
	}
	cob := util.NewCopyOnWriteBuffer(b)
	n := 0
	// Instead of scanning every byte, jump directly between the cached
	// positions of the next '&' and '\\', re-deriving via IndexByte only the
	// cursor that was actually consumed (or overrun) by the last match.
	for ampPos != -1 || bsPos != -1 {
		var i int
		if bsPos != -1 && (ampPos == -1 || bsPos < ampPos) {
			i = bsPos
		} else {
			i = ampPos
		}
		next := i + 1
		if b[i] == '\\' {
			if i < limit-1 && (util.IsPunct(b[next]) || d.cfg.EscapedSpace && b[next] == ' ') {
				cob.Write(b[n:i])
				if b[next] != ' ' {
					_ = cob.WriteByte(b[next])
				}
				n = next + 1
				next = n
			}
		} else { // b[i] == '&'
			pos := i
			handled := false
			if next < limit {
				if b[next] == '#' {
					nnext := next + 1
					if nnext < limit {
						nc := b[nnext]
						// code point like #x22;
						if nnext < limit && nc == 'x' || nc == 'X' {
							start := nnext + 1
							j, ok := util.ReadWhile(b, [2]int{start, limit}, util.IsHexDecimal)
							if ok && j < limit && b[j] == ';' && j-start < 7 {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:j]), 16, 32)
								cob.Write(b[n:pos])
								n = j + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								cob.Write(buf[:runeSize])
								next = n
								handled = true
							}
							// code point like #1234;
						} else if nc >= '0' && nc <= '9' {
							start := nnext
							j, ok := util.ReadWhile(b, [2]int{start, limit}, util.IsNumeric)
							if ok && j < limit && j-start < 8 && b[j] == ';' {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:j]), 10, 32)
								cob.Write(b[n:pos])
								n = j + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								cob.Write(buf[:runeSize])
								next = n
								handled = true
							}
						}
					}
				} else {
					start := next
					j, ok := util.ReadWhile(b, [2]int{start, limit}, util.IsAlphaNumeric)
					if ok && j < limit && b[j] == ';' {
						name := util.BytesToReadOnlyString(b[start:j])
						entity, found := util.LookUpHTML5EntityByName(name)
						if found {
							cob.Write(b[n:pos])
							n = j + 1
							cob.Write(entity.Characters)
							next = n
							handled = true
						}
					}
				}
			}
			if !handled {
				next = pos + 1
			}
		}
		if ampPos != -1 && ampPos < next {
			ampPos = indexByteFrom(b, next, '&')
		}
		if bsPos != -1 && bsPos < next {
			bsPos = indexByteFrom(b, next, '\\')
		}
	}
	if cob.IsCopied() {
		cob.Write(b[n:])
	}
	return cob.Bytes()
}

// DecodeTo implements the [Decoder] interface.
func (d *DefaultDecoder) DecodeTo(w io.Writer, b []byte) (int, error) {
	limit := len(b)
	ampPos := bytes.IndexByte(b, '&')
	bsPos := bytes.IndexByte(b, '\\')
	if ampPos == -1 && bsPos == -1 {
		return w.Write(b)
	}
	written := 0
	n := 0
	// Instead of scanning every byte, jump directly between the cached
	// positions of the next '&' and '\\', re-deriving via IndexByte only the
	// cursor that was actually consumed (or overrun) by the last match.
	for ampPos != -1 || bsPos != -1 {
		var i int
		if bsPos != -1 && (ampPos == -1 || bsPos < ampPos) {
			i = bsPos
		} else {
			i = ampPos
		}
		next := i + 1
		if b[i] == '\\' {
			if i < limit-1 && (util.IsPunct(b[next]) || d.cfg.EscapedSpace && b[next] == ' ') {
				wr, err := w.Write(b[n:i])
				written += wr
				if err != nil {
					return written, err
				}
				if b[next] != ' ' {
					wr, err := w.Write(b[next : next+1])
					written += wr
					if err != nil {
						return written, err
					}
				}
				n = next + 1
				next = n
			}
		} else { // b[i] == '&'
			pos := i
			handled := false
			if next < limit {
				if b[next] == '#' {
					nnext := next + 1
					if nnext < limit {
						nc := b[nnext]
						// code point like #x22;
						if nnext < limit && nc == 'x' || nc == 'X' {
							start := nnext + 1
							j, ok := util.ReadWhile(b, [2]int{start, limit}, util.IsHexDecimal)
							if ok && j < limit && b[j] == ';' && j-start < 7 {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:j]), 16, 32)
								wr, err := w.Write(b[n:pos])
								written += wr
								if err != nil {
									return written, err
								}
								n = j + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								wr, err = w.Write(buf[:runeSize])
								written += wr
								if err != nil {
									return written, err
								}
								next = n
								handled = true
							}
							// code point like #1234;
						} else if nc >= '0' && nc <= '9' {
							start := nnext
							j, ok := util.ReadWhile(b, [2]int{start, limit}, util.IsNumeric)
							if ok && j < limit && j-start < 8 && b[j] == ';' {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:j]), 10, 32)
								wr, err := w.Write(b[n:pos])
								written += wr
								if err != nil {
									return written, err
								}
								n = j + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								wr, err = w.Write(buf[:runeSize])
								written += wr
								if err != nil {
									return written, err
								}
								next = n
								handled = true
							}
						}
					}
				} else {
					start := next
					j, ok := util.ReadWhile(b, [2]int{start, limit}, util.IsAlphaNumeric)
					if ok && j < limit && b[j] == ';' {
						name := util.BytesToReadOnlyString(b[start:j])
						entity, found := util.LookUpHTML5EntityByName(name)
						if found {
							wr, err := w.Write(b[n:pos])
							written += wr
							if err != nil {
								return written, err
							}
							n = j + 1
							wr, err = w.Write(entity.Characters)
							written += wr
							if err != nil {
								return written, err
							}
							next = n
							handled = true
						}
					}
				}
			}
			if !handled {
				next = pos + 1
			}
		}
		if ampPos != -1 && ampPos < next {
			ampPos = indexByteFrom(b, next, '&')
		}
		if bsPos != -1 && bsPos < next {
			bsPos = indexByteFrom(b, next, '\\')
		}
	}
	wr, err := w.Write(b[n:])
	written += wr
	if err != nil {
		return written, err
	}
	return written, nil
}

type identityDecoder struct {
}

func (d *identityDecoder) Decode(b []byte) []byte {
	return b
}

func (d *identityDecoder) DecodeTo(w io.Writer, b []byte) (int, error) {
	return w.Write(b)
}

// IdentityDecoder is a decoder that does not perform any decoding and returns the bytes as is.
var IdentityDecoder Decoder = &identityDecoder{}
