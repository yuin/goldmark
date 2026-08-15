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

// Decode implements the [Decoder] interface.
func (d *DefaultDecoder) Decode(b []byte) []byte {
	if bytes.IndexByte(b, '&') == -1 && bytes.IndexByte(b, '\\') == -1 {
		return b
	}
	cob := util.NewCopyOnWriteBuffer(b)
	limit := len(b)
	var ok bool
	n := 0
	for i := 0; i < limit; i++ {
		c := b[i]
		if i < limit-1 && c == '\\' && (util.IsPunct(b[i+1]) || d.cfg.EscapedSpace && b[i+1] == ' ') {
			cob.Write(b[n:i])
			if b[i+1] != ' ' {
				_ = cob.WriteByte(b[i+1])
			}
			i++
			n = i + 1
			continue
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit {
				if b[next] == '#' {
					nnext := next + 1
					if nnext < limit {
						nc := b[nnext]
						// code point like #x22;
						if nnext < limit && nc == 'x' || nc == 'X' {
							start := nnext + 1
							i, ok = util.ReadWhile(b, [2]int{start, limit}, util.IsHexDecimal)
							if ok && i < limit && b[i] == ';' && i-start < 7 {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:i]), 16, 32)
								cob.Write(b[n:pos])
								n = i + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								cob.Write(buf[:runeSize])
								continue
							}
							// code point like #1234;
						} else if nc >= '0' && nc <= '9' {
							start := nnext
							i, ok = util.ReadWhile(b, [2]int{start, limit}, util.IsNumeric)
							if ok && i < limit && i-start < 8 && b[i] == ';' {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:i]), 10, 32)
								cob.Write(b[n:pos])
								n = i + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								cob.Write(buf[:runeSize])
								continue
							}
						}
					}
				} else {
					start := next
					i, ok = util.ReadWhile(b, [2]int{start, limit}, util.IsAlphaNumeric)
					if ok && i < limit && b[i] == ';' {
						name := util.BytesToReadOnlyString(b[start:i])
						entity, ok := util.LookUpHTML5EntityByName(name)
						if ok {
							cob.Write(b[n:pos])
							n = i + 1
							cob.Write(entity.Characters)
							continue
						}
					}
				}
			}
			i = next - 1
		}
	}
	if cob.IsCopied() {
		cob.Write(b[n:])
	}
	return cob.Bytes()
}

// DecodeTo implements the [Decoder] interface.
func (d *DefaultDecoder) DecodeTo(w io.Writer, b []byte) (int, error) {
	if bytes.IndexByte(b, '&') == -1 && bytes.IndexByte(b, '\\') == -1 {
		return w.Write(b)
	}
	written := 0
	limit := len(b)
	var ok bool
	n := 0
	for i := 0; i < limit; i++ {
		c := b[i]
		if i < limit-1 && c == '\\' && (util.IsPunct(b[i+1]) || d.cfg.EscapedSpace && b[i+1] == ' ') {
			wr, err := w.Write(b[n:i])
			written += wr
			if err != nil {
				return written, err
			}
			if b[i+1] != ' ' {
				wr, err := w.Write(b[i+1 : i+2])
				written += wr
				if err != nil {
					return written, err
				}
			}
			i++
			n = i + 1
			continue
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit {
				if b[next] == '#' {
					nnext := next + 1
					if nnext < limit {
						nc := b[nnext]
						// code point like #x22;
						if nnext < limit && nc == 'x' || nc == 'X' {
							start := nnext + 1
							i, ok = util.ReadWhile(b, [2]int{start, limit}, util.IsHexDecimal)
							if ok && i < limit && b[i] == ';' && i-start < 7 {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:i]), 16, 32)
								wr, err := w.Write(b[n:pos])
								written += wr
								if err != nil {
									return written, err
								}
								n = i + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								wr, err = w.Write(buf[:runeSize])
								written += wr
								if err != nil {
									return written, err
								}
								continue
							}
							// code point like #1234;
						} else if nc >= '0' && nc <= '9' {
							start := nnext
							i, ok = util.ReadWhile(b, [2]int{start, limit}, util.IsNumeric)
							if ok && i < limit && i-start < 8 && b[i] == ';' {
								v, _ := strconv.ParseUint(util.BytesToReadOnlyString(b[start:i]), 10, 32)
								wr, err := w.Write(b[n:pos])
								written += wr
								if err != nil {
									return written, err
								}
								n = i + 1
								buf := make([]byte, 6)
								runeSize := utf8.EncodeRune(buf, util.ToValidRune(rune(v)))
								wr, err = w.Write(buf[:runeSize])
								written += wr
								if err != nil {
									return written, err
								}
								continue
							}
						}
					}
				} else {
					start := next
					i, ok = util.ReadWhile(b, [2]int{start, limit}, util.IsAlphaNumeric)
					if ok && i < limit && b[i] == ';' {
						name := util.BytesToReadOnlyString(b[start:i])
						entity, ok := util.LookUpHTML5EntityByName(name)
						if ok {
							wr, err := w.Write(b[n:pos])
							written += wr
							if err != nil {
								return written, err
							}
							n = i + 1
							wr, err = w.Write(entity.Characters)
							written += wr
							if err != nil {
								return written, err
							}
							continue
						}
					}
				}
			}
			i = next - 1
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

type codeSpanDecoder struct {
}

func (d *codeSpanDecoder) Decode(b []byte) []byte {
	if bytes.IndexByte(b, '\n') >= 0 {
		return bytes.ReplaceAll(b, []byte{'\n'}, []byte{' '})
	}
	return b
}

func (d *codeSpanDecoder) DecodeTo(w io.Writer, b []byte) (int, error) {
	return w.Write(d.Decode(b))
}

// CodeSpanDecoder is a decoder that decodes code spans.
var CodeSpanDecoder Decoder = &codeSpanDecoder{}
