package text

import (
	"bytes"
	"io"
	"slices"

	"github.com/yuin/goldmark/v2/util"
)

// An Index represents a position range in a source byte slice.
// Unlike Segment, Index does not carry parsing metadata such as
// Padding or ForceNewline.
type Index struct {
	Start int
	Stop  int
}

// NewIndex returns a new Index.
func NewIndex(start, stop int) Index {
	return Index{Start: start, Stop: stop}
}

// NewIndexFromSegment returns a new Index derived from the given Segment,
// dropping any padding and ignoring ForceNewline.
func NewIndexFromSegment(seg Segment) Index {
	return Index{Start: seg.Start, Stop: seg.Stop}
}

// IsEmpty returns true if this index is empty, otherwise false.
func (i Index) IsEmpty() bool {
	return i.Start >= i.Stop
}

// A Value represents an inline value.
type Value interface {
	// Value returns the decoded string representation of this value.
	Value(source []byte) string

	// Bytes returns the source byte slice corresponding to this value.
	Bytes(source []byte) []byte

	// Str returns the string representation of this value.
	Str(source []byte) string

	// IsOwned returns true if this value is an owned string not derived from the source byte slice.
	IsOwned() bool

	// IsEmpty returns true if this value is empty, otherwise false.
	IsEmpty() bool

	// Index returns the source position of this value.
	//
	// The result is meaningful only when IsOwned() returns false.
	// If Value is backed by multiple source positions, Index returns the first position.
	Index() Index

	// Indices returns the slice of Index values in this value.
	//
	// The result is meaningful only when IsOwned() returns false.
	Indices() []Index

	// WriteTo writes the value to the given buffer, using the [Value].Value.
	WriteTo(w io.Writer, source []byte) (int, error)
}

var _ Value = (*SingleLineValue)(nil)

// A SingleLineValue holds a single-line inline value that is either an owned string
// or a position range within a source byte slice.
// Use SingleLineValue for data that must not contain newlines per the CommonMark spec
// (e.g. link destinations, fenced code block info strings).
type SingleLineValue struct {
	decoder Decoder
	s       string
	index   Index
}

// SingleLineValueInput is a type constraint for types that can be converted to a Value.
type SingleLineValueInput interface {
	string | []byte | Index
}

// NewSingleLineValue returns a Value from the given input, bound to the given Decoder.
func NewSingleLineValue[T SingleLineValueInput](v T, decoder Decoder) SingleLineValue {
	switch val := any(v).(type) {
	case string:
		return NewSingleLineValueFromString(val, decoder)
	case []byte:
		return NewSingleLineValueFromString(util.BytesToReadOnlyString(val), decoder)
	case Index:
		return NewSingleLineValueFromIndex(val, decoder)
	default:
		panic("unsupported type")
	}
}

// NewSingleLineValueFromIndex returns a Value backed by a source position, bound to the
// given Decoder.
func NewSingleLineValueFromIndex(i Index, decoder Decoder) SingleLineValue {
	return SingleLineValue{decoder: decoder, index: i}
}

// NewSingleLineValueFromSegment returns a Value backed by a source position derived from the
// given Segment, bound to the given Decoder.
func NewSingleLineValueFromSegment(seg Segment, decoder Decoder) SingleLineValue {
	return NewSingleLineValueFromIndex(NewIndexFromSegment(seg), decoder)
}

// NewSingleLineValueFromString returns a Value backed by an owned string, bound to the
// given Decoder.
// This function does not check whether the string contains newlines; it is
// the caller's responsibility to ensure that the string is single-line.
//
// [ValueBuilder.Build] will automatically choose between [SingleLineValue] and
// [MultiLineValue] based on the presence of newlines in the string.
func NewSingleLineValueFromString(s string, decoder Decoder) SingleLineValue {
	return SingleLineValue{decoder: decoder, s: s}
}

func (v SingleLineValue) decoderOrDefault() Decoder {
	if v.decoder == nil {
		return IdentityDecoder
	}
	return v.decoder
}

// Value implements [Value].Value .
func (v SingleLineValue) Value(source []byte) string {
	b := v.decoderOrDefault().Decode(v.Bytes(source))
	return util.BytesToReadOnlyString(b)
}

// WriteTo implements [Value].WriteTo .
func (v SingleLineValue) WriteTo(w io.Writer, source []byte) (int, error) {
	return v.decoderOrDefault().DecodeTo(w, v.Bytes(source))
}

// Bytes implements [Value].Bytes .
func (v SingleLineValue) Bytes(source []byte) []byte {
	if v.IsOwned() {
		return util.StringToReadOnlyBytes(v.s)
	}
	return source[v.index.Start:v.index.Stop]
}

// Str implements [Value].Str .
func (v SingleLineValue) Str(source []byte) string {
	if v.IsOwned() {
		return v.s
	}
	return util.BytesToReadOnlyString(source[v.index.Start:v.index.Stop])
}

// IsOwned implements [Value].IsOwned .
func (v SingleLineValue) IsOwned() bool {
	return v.index.Start == 0 && v.index.Stop == 0
}

// IsEmpty implements [Value].IsEmpty .
func (v SingleLineValue) IsEmpty() bool {
	if v.IsOwned() {
		return v.s == ""
	}
	return v.index.Start >= v.index.Stop
}

// Index implements [Value].Index .
func (v SingleLineValue) Index() Index {
	return v.index
}

// Indices implements [Value].Indices .
func (v SingleLineValue) Indices() []Index {
	return []Index{v.index}
}

// WithStop returns a new SingleLineValue with the same start index but a different stop index.
// This method panics if the value is owned.
func (v SingleLineValue) WithStop(stop int) SingleLineValue {
	if v.IsOwned() {
		panic("cannot set stop on owned value")
	}
	return SingleLineValue{decoder: v.decoder, index: Index{Start: v.index.Start, Stop: stop}}
}

var _ Value = (*MultiLineValue)(nil)

// A MultiLineValue holds a potentially multiline inline value that is
// either an owned string or a set of source position ranges.
// Use MultiLineValue for data that may span multiple lines per the CommonMark
// spec (e.g. link labels).
//
// When constructing the byte value from source positions, the ranges are
// simply concatenated verbatim; no newline folding or other transformation
// is applied.
type MultiLineValue struct {
	decoder Decoder
	s       string
	index   [1]Index
	indices []Index
}

// MultiLineValueInput is a type constraint for types that can be converted to a MultiLineValue.
type MultiLineValueInput interface {
	string | []byte | Index | []Index
}

// NewMultiLineValue returns a MultiLineValue from the given input, which may be a string,
// byte slice, Index, or slice of Index, bound to the given Decoder.
func NewMultiLineValue[T MultiLineValueInput](v T, decoder Decoder) MultiLineValue {
	switch val := any(v).(type) {
	case string:
		return NewMultiLineValueFromString(val, decoder)
	case []byte:
		return NewMultiLineValueFromString(util.BytesToReadOnlyString(val), decoder)
	case Index:
		return NewMultiLineValueFromIndex(val, decoder)
	case []Index:
		return NewMultiLineValueFromIndices(val, decoder)
	default:
		panic("unsupported type")
	}
}

// NewMultiLineValueFromIndex returns a MultiLineValue backed by a single source position,
// bound to the given Decoder.
func NewMultiLineValueFromIndex(i Index, decoder Decoder) MultiLineValue {
	return MultiLineValue{decoder: decoder, index: [1]Index{i}}
}

// NewMultiLineValueFromIndices returns a MultiLineValue backed by source positions,
// bound to the given Decoder.
func NewMultiLineValueFromIndices(indices []Index, decoder Decoder) MultiLineValue {
	if len(indices) == 1 {
		return MultiLineValue{decoder: decoder, index: [1]Index{indices[0]}}
	}
	return MultiLineValue{decoder: decoder, indices: indices}
}

// NewMultiLineValueFromString returns a MultiLineValue backed by an owned string, bound to
// the given Decoder.
func NewMultiLineValueFromString(s string, decoder Decoder) MultiLineValue {
	return MultiLineValue{decoder: decoder, s: s}
}

func (v MultiLineValue) decoderOrDefault() Decoder {
	if v.decoder == nil {
		return IdentityDecoder
	}
	return v.decoder
}

// Value implements [Value].Value .
func (v MultiLineValue) Value(source []byte) string {
	d := v.decoderOrDefault()
	if v.IsOwned() {
		b := util.StringToReadOnlyBytes(v.s)
		return util.BytesToReadOnlyString(d.Decode(b))
	}
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		b := source[v.index[0].Start:v.index[0].Stop]
		return util.BytesToReadOnlyString(d.Decode(b))
	}

	if len(v.indices) == 0 {
		return ""
	}

	if len(v.indices) == 1 {
		b := source[v.indices[0].Start:v.indices[0].Stop]
		return util.BytesToReadOnlyString(d.Decode(b))
	}

	if contiguous, start, stop := contiguousIndices(v.indices); contiguous {
		return util.BytesToReadOnlyString(d.Decode(source[start:stop]))
	}

	b := slices.Clone(source[v.indices[0].Start:v.indices[0].Stop])
	for _, idx := range v.indices[1:] {
		chunk := source[idx.Start:idx.Stop]
		b = append(b, d.Decode(chunk)...)
	}
	return util.BytesToReadOnlyString(b)
}

func contiguousIndices(indices []Index) (contiguous bool, start, stop int) {
	start, stop = indices[0].Start, indices[0].Stop
	for _, idx := range indices[1:] {
		if idx.Start != stop {
			return false, 0, 0
		}
		stop = idx.Stop
	}
	return true, start, stop
}

// WriteTo implements [Value].WriteTo .
func (v MultiLineValue) WriteTo(w io.Writer, source []byte) (int, error) {
	if v.IsOwned() {
		return v.decoderOrDefault().DecodeTo(w, util.StringToReadOnlyBytes(v.s))
	}
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		return v.decoderOrDefault().DecodeTo(w, source[v.index[0].Start:v.index[0].Stop])
	}
	if len(v.indices) == 0 {
		return 0, nil
	}
	if len(v.indices) == 1 {
		return v.decoderOrDefault().DecodeTo(w, source[v.indices[0].Start:v.indices[0].Stop])
	}
	n := int(0)
	d := v.decoderOrDefault()
	for _, idx := range v.indices {
		written, err := d.DecodeTo(w, source[idx.Start:idx.Stop])
		n += written
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// Bytes implements [Value].Bytes .
func (v MultiLineValue) Bytes(source []byte) []byte {
	if v.IsOwned() {
		return util.StringToReadOnlyBytes(v.s)
	}
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		return source[v.index[0].Start:v.index[0].Stop]
	}
	if len(v.indices) == 0 {
		return nil
	}
	if len(v.indices) == 1 {
		return source[v.indices[0].Start:v.indices[0].Stop]
	}
	buf := slices.Clone(source[v.indices[0].Start:v.indices[0].Stop])
	for _, idx := range v.indices[1:] {
		chunk := source[idx.Start:idx.Stop]
		buf = append(buf, chunk...)
	}
	return buf
}

// Str implements [Value].Str .
func (v MultiLineValue) Str(source []byte) string {
	if v.IsOwned() {
		return v.s
	}
	return util.BytesToReadOnlyString(v.Bytes(source))
}

// IsOwned implements [Value].IsOwned .
func (v MultiLineValue) IsOwned() bool {
	return v.indices == nil && v.index[0].Start == 0 && v.index[0].Stop == 0
}

// IsEmpty implements [Value].IsEmpty .
func (v MultiLineValue) IsEmpty() bool {
	if v.IsOwned() {
		return v.s == ""
	}
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		return v.index[0].Start >= v.index[0].Stop
	}
	for _, idx := range v.indices {
		if idx.Start < idx.Stop {
			return false
		}
	}
	return true
}

// Index implements [Value].Index .
func (v MultiLineValue) Index() Index {
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		return v.index[0]
	}
	if len(v.indices) > 0 {
		return v.indices[0]
	}
	return Index{}
}

// Indices implements [Value].Indices .
func (v MultiLineValue) Indices() []Index {
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		return v.index[:]
	}
	return v.indices
}

// ValueBuilder is a helper for building a [Value].
type ValueBuilder struct {
	decoder Decoder
	s       string
	index   Index
	indices []Index
}

// AddIndex adds an Index to the builder.
func (b *ValueBuilder) AddIndex(i Index) *ValueBuilder {
	if b.index.Start == 0 && b.index.Stop == 0 && b.indices == nil {
		b.index = i
		return b
	}

	if b.indices == nil {
		b.indices = make([]Index, 0, 8)
		b.indices = append(b.indices, b.index)
		b.index = Index{}
	}
	b.indices = append(b.indices, i)
	return b
}

// Decoder sets the Decoder to be bound to the Value produced by Build, BuildSingleLine,
// or BuildMultiLine. If not set, the produced Value is bound to [IdentityDecoder].
func (b *ValueBuilder) Decoder(d Decoder) *ValueBuilder {
	b.decoder = d
	return b
}

func (b *ValueBuilder) decoderOrDefault() Decoder {
	if b.decoder == nil {
		return IdentityDecoder
	}
	return b.decoder
}

// AddSegment adds an Index derived from the given Segment to the builder.
func (b *ValueBuilder) AddSegment(seg Segment) *ValueBuilder {
	return b.AddIndex(NewIndexFromSegment(seg))
}

// OwnedString sets an owned string value. When Build is called, the resulting
// Value will be backed by this string rather than any accumulated indices.
func (b *ValueBuilder) OwnedString(s string) *ValueBuilder {
	b.s = s
	return b
}

// OwnedBytes sets an owned string value from the given byte slice. When Build is called, the resulting
// Value will be backed by this string rather than any accumulated indices.
func (b *ValueBuilder) OwnedBytes(bts []byte) *ValueBuilder {
	b.s = util.BytesToReadOnlyString(bts)
	return b
}

func (b *ValueBuilder) isSingle() bool {
	return b.indices == nil && b.s == ""
}

// BuildMultiLine returns a MultiLineValue from the accumulated state.
// If OwnedString or OwnedBytes was called, the result is backed by that string.
// Otherwise, the result is backed by the accumulated indices.
func (b *ValueBuilder) BuildMultiLine() MultiLineValue {
	d := b.decoderOrDefault()
	if b.s != "" {
		return NewMultiLineValueFromString(b.s, d)
	}
	if b.isSingle() {
		return NewMultiLineValueFromIndex(b.index, d)
	}
	return NewMultiLineValueFromIndices(b.indices, d)
}

// BuildSingleLine returns a SingleLineValue from the accumulated state.
// If OwnedString or OwnedBytes was called, the result is backed by that string.
// Otherwise, the result is backed by the first accumulated index.
func (b *ValueBuilder) BuildSingleLine() SingleLineValue {
	d := b.decoderOrDefault()
	if b.s != "" {
		return NewSingleLineValueFromString(b.s, d)
	}
	return NewSingleLineValueFromIndex(b.index, d)
}

// Build returns a Value from the accumulated state.
// If OwnedString or OwnedBytes was called, the result is backed by that string.
// Otherwise, the result is backed by the accumulated indices.
func (b *ValueBuilder) Build() Value {
	if b.s != "" {
		bs := util.StringToReadOnlyBytes(b.s)
		if bytes.IndexByte(bs, '\n') == -1 {
			return b.BuildSingleLine()
		}
		return b.BuildMultiLine()
	}
	if b.isSingle() {
		return b.BuildSingleLine()
	}
	return b.BuildMultiLine()
}

// A Lines value holds the raw content of a block node.  It is either an
// owned string or a sequence of Segment values copied from the parser's
// source-segment list.
//
// Lines is used for block nodes whose rendered content is taken directly from the
// source (e.g. CodeBlock, FencedCodeBlock, HTMLBlock).
//
// Lines is always 'raw'; it does not perform any decoding of the source content.
type Lines struct {
	s    string
	segs []Segment
}

// LinesInput is a type constraint for types that can be converted to a Lines.
type LinesInput interface {
	string | []byte | []Segment | Lines
}

// NewLines returns a Lines from the given input, which may be a string, byte slice,
// or slice of Segment.
func NewLines[T LinesInput](v T) Lines {
	switch val := any(v).(type) {
	case string:
		return NewLinesFromString(val)
	case []byte:
		return NewLinesFromString(util.BytesToReadOnlyString(val))
	case []Segment:
		return NewLinesFromSegments(val)
	case Lines:
		return val
	default:
		panic("unsupported type")
	}
}

// NewLinesFromSegments returns a Lines backed by source segments.
func NewLinesFromSegments(segs []Segment) Lines {
	return Lines{segs: segs}
}

// NewLinesFromString returns a Lines backed by an owned string.
func NewLinesFromString(s string) Lines {
	return Lines{s: s}
}

// Segments returns the internal segment slice, allowing callers to
// read and otherwise inspect the segment list.
func (l Lines) Segments() []Segment {
	return l.segs
}

// AppendSegment appends a Segment to the internal segment list.
func (l *Lines) AppendSegment(seg Segment) {
	if l.segs == nil {
		l.segs = make([]Segment, 0, 4)
	}
	l.segs = append(l.segs, seg)
}

// Bytes returns the concatenated byte content of all segments, or the
// owned string.
// The returned byte slice is read-only and should not be modified.
func (l Lines) Bytes(source []byte) []byte {
	if l.IsOwned() {
		return util.StringToReadOnlyBytes(l.s)
	}
	segs := l.segs
	if len(segs) == 0 {
		return nil
	}
	if len(segs) == 1 {
		return segs[0].Bytes(source)
	}
	// If all segments are contiguous in the source with no padding or
	// ForceNewline, return a single sub-slice to avoid allocation.
	first := segs[0]
	if first.Padding == 0 && !first.ForceNewline {
		contiguous := true
		prev := first
		for _, seg := range segs[1:] {
			if seg.Padding != 0 || seg.ForceNewline || seg.Start != prev.Stop {
				contiguous = false
				break
			}
			prev = seg
		}
		if contiguous {
			return source[first.Start:prev.Stop]
		}
	}
	sizeAdvice := 0
	for _, seg := range segs {
		sizeAdvice += seg.Len()
	}
	result := make([]byte, 0, sizeAdvice)

	for _, seg := range segs {
		result = append(result, seg.Bytes(source)...)
	}
	return result
}

// Str returns the string representation of this value.
// The returned string is read-only and should not be modified.
// See Bytes() for details on how the string is constructed from source segments.
func (l Lines) Str(source []byte) string {
	if l.IsOwned() {
		return l.s
	}
	return util.BytesToReadOnlyString(l.Bytes(source))
}

// WriteTo writes the value to the given buffer, using the [Lines].Bytes.
func (l Lines) WriteTo(w io.Writer, source []byte) (int, error) {
	if l.IsOwned() {
		return w.Write(util.StringToReadOnlyBytes(l.s))
	}
	if len(l.segs) == 0 {
		return 0, nil
	}
	if len(l.segs) == 1 {
		return w.Write(l.segs[0].Bytes(source))
	}
	n := 0
	for _, seg := range l.segs {
		b := seg.Bytes(source)
		written, err := w.Write(b)
		n += written
		if err != nil {
			return n, err
		}
		if seg.ForceNewline && (len(b) == 0 || b[len(b)-1] != '\n') {
			written, err := w.Write([]byte{'\n'})
			n += written
			if err != nil {
				return n, err
			}
		}
	}
	return n, nil
}

// IsOwned returns true if this value is an owned string not derived from
// the source byte slice.
func (l Lines) IsOwned() bool {
	return l.segs == nil
}

var space = []byte(" ")

// A Segment struct holds information about source positions.
type Segment struct {
	// Start is a start position of the segment.
	Start int

	// Stop is a stop position of the segment.
	// This value should be excluded.
	Stop int

	// Padding is a padding length of the segment.
	Padding int

	// ForceNewline is true if the segment should be ended with a newline.
	// Some elements(i.e. CodeBlock, FencedCodeBlock) does not trim trailing
	// newlines. Spec defines that EOF is treated as a newline, so we need to
	// add a newline to the end of the segment if it is not empty.
	//
	// i.e.:
	//
	//     ```go
	//     const test = "test"
	//
	// This code does not close the code block and ends with EOF. In this case,
	// we need to add a newline to the end of the last line like `const test = "test"\n`.
	ForceNewline bool
}

// NewSegment return a new Segment.
func NewSegment(start, stop int) Segment {
	return Segment{
		Start:   start,
		Stop:    stop,
		Padding: 0,
	}
}

// NewSegmentPadding returns a new Segment with the given padding.
func NewSegmentPadding(start, stop, padding int) Segment {
	return Segment{
		Start:   start,
		Stop:    stop,
		Padding: padding,
	}
}

// Bytes returns bytes of the segment.
func (t Segment) Bytes(source []byte) []byte {
	var result []byte
	if t.Padding == 0 {
		result = source[t.Start:t.Stop]
	} else {
		// Fill the padding directly instead of allocating a throwaway slice
		// via bytes.Repeat and copying it in with append.
		result = make([]byte, t.Padding, t.Padding+t.Stop-t.Start+1)
		for i := range result {
			result[i] = ' '
		}
		result = append(result, source[t.Start:t.Stop]...)
	}
	if t.ForceNewline && len(result) > 0 && result[len(result)-1] != '\n' {
		result = append(result, '\n')
	}
	return result
}

// Str returns a string of the segment.
func (t Segment) Str(source []byte) string {
	return util.BytesToReadOnlyString(t.Bytes(source))
}

// Len returns a length of the segment.
func (t Segment) Len() int {
	return t.Stop - t.Start + t.Padding
}

// Between returns a segment between this segment and the given segment.
func (t Segment) Between(other Segment) Segment {
	if t.Stop != other.Stop {
		panic("invalid state")
	}
	return NewSegmentPadding(
		t.Start,
		other.Start,
		t.Padding-other.Padding,
	)
}

// IsEmpty returns true if this segment is empty, otherwise false.
func (t Segment) IsEmpty() bool {
	return t.Start >= t.Stop && t.Padding == 0
}

// TrimRightSpace returns a new segment by slicing off all trailing
// space characters.
func (t Segment) TrimRightSpace(source []byte) Segment {
	v := source[t.Start:t.Stop]
	l := util.TrimRightSpaceLength(v)
	if l == len(v) {
		return NewSegment(t.Start, t.Start)
	}
	return NewSegmentPadding(t.Start, t.Stop-l, t.Padding)
}

// TrimLeftSpace returns a new segment by slicing off all leading
// space characters including padding.
func (t Segment) TrimLeftSpace(source []byte) Segment {
	v := source[t.Start:t.Stop]
	l := util.TrimLeftSpaceLength(v)
	return NewSegment(t.Start+l, t.Stop)
}

// TrimLeftSpaceWidth returns a new segment by slicing off leading space
// characters until the given width.
func (t Segment) TrimLeftSpaceWidth(width int, source []byte) Segment {
	padding := t.Padding
	for ; width > 0; width-- {
		if padding == 0 {
			break
		}
		padding--
	}
	if width == 0 {
		return NewSegmentPadding(t.Start, t.Stop, padding)
	}
	text := source[t.Start:t.Stop]
	start := t.Start
loop:
	for _, c := range text {
		if start >= t.Stop-1 || width <= 0 {
			break
		}
		switch c {
		case ' ':
			width--
		case '\t':
			width -= 4
		default:
			break loop
		}
		start++
	}
	if width < 0 {
		padding = width * -1
	}
	return NewSegmentPadding(start, t.Stop, padding)
}

// WithStart returns a new Segment with the given Start and the same Stop.
// Padding and ForceNewline are reset to their zero values, since Padding
// describes space that precedes the original Start position.
func (t Segment) WithStart(v int) Segment {
	return NewSegmentPadding(v, t.Stop, 0)
}

// WithStop returns a new Segment with the given Stop and the same Start and
// Padding. ForceNewline is reset to false.
func (t Segment) WithStop(v int) Segment {
	return NewSegmentPadding(t.Start, v, t.Padding)
}
