package text

import (
	"bytes"
	"slices"
	"unsafe"

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

type valuer interface {
	Value(source []byte) string
	Bytes(source []byte) []byte
	Str(source []byte) string
}

var _ valuer = (*Value)(nil)

// A Value holds a single-line inline value that is either an owned string
// or a position range within a source byte slice.
// Use Value for data that must not contain newlines per the CommonMark spec
// (e.g. link destinations, fenced code block info strings).
type Value struct {
	s     string
	index Index
}

// ValueInput is a type constraint for types that can be converted to a Value.
type ValueInput interface {
	string | []byte | Index | Value
}

// NewValue returns a Value from the given input, which may be a string, byte slice, or Index.
func NewValue[T ValueInput](v T) Value {
	switch val := any(v).(type) {
	case string:
		return NewStringValue(val)
	case []byte:
		return NewStringValue(util.BytesToReadOnlyString(val))
	case Index:
		return NewIndexValue(val)
	case Value:
		return val
	default:
		panic("unsupported type")
	}
}

// NewIndexValue returns a Value backed by a source position.
func NewIndexValue(i Index) Value {
	return Value{index: i}
}

// NewStringValue returns a Value backed by an owned string.
func NewStringValue(s string) Value {
	return Value{s: s}
}

// Value returns a Value representation of this type.
//
// - HTML entities in the returned string are decoded.
func (v Value) Value(source []byte) string {
	b := v.Bytes(source)
	resolved, changed := resolveEntityReferences(b)
	if changed {
		return util.BytesToReadOnlyString(resolved)
	}
	return util.BytesToReadOnlyString(b)
}

// Bytes returns the source byte slice corresponding to this value.
//
// - HTML entities in the returned byte slice are not decoded.
func (v Value) Bytes(source []byte) []byte {
	if v.IsOwned() {
		return util.StringToReadOnlyBytes(v.s)
	}
	return source[v.index.Start:v.index.Stop]
}

// Str returns the string representation of this value.
//
// - HTML entities in the returned string are not decoded.
func (v Value) Str(source []byte) string {
	if v.IsOwned() {
		return v.s
	}
	return util.BytesToReadOnlyString(source[v.index.Start:v.index.Stop])
}

// IsOwned returns true if this value is an owned string not derived from
// the source byte slice.
func (v Value) IsOwned() bool {
	return v.index.Start == 0 && v.index.Stop == 0
}

// Index returns the source position of this value.
// The result is meaningful only when IsOwned() returns false.
func (v Value) Index() Index {
	return v.index
}

var _ valuer = (*MultilineValue)(nil)

// A MultilineValue holds a potentially multiline inline value that is
// either an owned string or a set of source position ranges.
// Use MultilineValue for data that may span multiple lines per the CommonMark
// spec (e.g. link labels).
//
// When constructing the byte value from source positions, the ranges are
// concatenated and any trailing newline of each range is replaced with a
// single space, matching CommonMark's line-folding rules.
type MultilineValue struct {
	s       string
	index   [1]Index
	indices []Index
}

// MultilineValueInput is a type constraint for types that can be converted to a MultilineValue.
type MultilineValueInput interface {
	string | []byte | Index | []Index | MultilineValue
}

// NewMultilineValue returns a MultilineValue from the given input, which may be a string,
// byte slice, Index, or slice of Index.
func NewMultilineValue[T MultilineValueInput](v T) MultilineValue {
	switch val := any(v).(type) {
	case string:
		return NewStringMultilineValue(val)
	case []byte:
		return NewStringMultilineValue(util.BytesToReadOnlyString(val))
	case Index:
		return NewIndexMultilineValue(val)
	case []Index:
		return NewIndicesMultilineValue(val)
	case MultilineValue:
		return val
	default:
		panic("unsupported type")
	}
}

// NewIndexMultilineValue returns a MultilineValue backed by a single source position.
func NewIndexMultilineValue(i Index) MultilineValue {
	return MultilineValue{index: [1]Index{i}}
}

// NewIndicesMultilineValue returns a MultilineValue backed by source positions.
func NewIndicesMultilineValue(indices []Index) MultilineValue {
	if len(indices) == 1 {
		return MultilineValue{index: [1]Index{indices[0]}}
	}
	return MultilineValue{indices: indices}
}

// NewMultilineValueFromSegments returns a MultilineValue backed by source positions
// derived from the given segments. Segment padding is dropped since MultilineValue
// is used for inline content where padding does not apply.
func NewMultilineValueFromSegments(segs []Segment) MultilineValue {
	if len(segs) == 1 {
		return MultilineValue{index: [1]Index{{Start: segs[0].Start, Stop: segs[0].Stop}}}
	}
	indices := make([]Index, len(segs))
	for i, seg := range segs {
		indices[i] = Index{Start: seg.Start, Stop: seg.Stop}
	}
	return MultilineValue{indices: indices}
}

// NewStringMultilineValue returns a MultilineValue backed by an owned string.
func NewStringMultilineValue(s string) MultilineValue {
	return MultilineValue{s: s}
}

// Value returns a Value representation of this type.
//
// - HTML entities in the returned string are decoded.
func (v MultilineValue) Value(source []byte) string {
	if v.IsOwned() {
		b := util.StringToReadOnlyBytes(v.s)
		resolved, changed := resolveEntityReferences(b)
		if changed {
			return util.BytesToReadOnlyString(resolved)
		}
		return v.s
	}
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		b := source[v.index[0].Start:v.index[0].Stop]
		resolved, changed := resolveEntityReferences(b)
		if changed {
			return util.BytesToReadOnlyString(resolved)
		}
		return util.BytesToReadOnlyString(b)
	}

	if len(v.indices) == 0 {
		return ""
	}

	if len(v.indices) == 1 {
		b := source[v.indices[0].Start:v.indices[0].Stop]
		resolved, changed := resolveEntityReferences(b)
		if changed {
			return util.BytesToReadOnlyString(resolved)
		}
		return util.BytesToReadOnlyString(b)
	}

	b := slices.Clone(source[v.indices[0].Start:v.indices[0].Stop])
	for _, idx := range v.indices[1:] {
		chunk := source[idx.Start:idx.Stop]
		resolved, changed := resolveEntityReferences(chunk)
		if changed {
			b = append(b, resolved...)
		} else {
			b = append(b, chunk...)
		}
	}
	return util.BytesToReadOnlyString(b)
}

// Bytes returns the source byte slice corresponding to this value.
//
// - HTML entities in the returned byte slice are not decoded.
func (v MultilineValue) Bytes(source []byte) []byte {
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

// Str returns the string representation of this value.
//
// - HTML entities in the returned string are not decoded.
func (v MultilineValue) Str(source []byte) string {
	if v.IsOwned() {
		return v.s
	}
	return util.BytesToReadOnlyString(v.Bytes(source))
}

// IsOwned returns true if this value is an owned string not derived from
// the source byte slice.
func (v MultilineValue) IsOwned() bool {
	return v.indices == nil && v.index[0].Start == 0 && v.index[0].Stop == 0
}

// Indices returns the slice of Index values in this value. The result is
// meaningful only when IsOwned() returns false.
func (v MultilineValue) Indices() []Index {
	if v.index[0].Start != 0 || v.index[0].Stop != 0 {
		return v.index[:]
	}
	return v.indices
}

// MultilineValueBuilder is a helper for building a MultilineValue. It optimises
// for the common case of a single Index by storing it directly in the struct
// and only allocating a slice if multiple indices are added.
type MultilineValueBuilder struct {
	s       string
	index   Index
	indices []Index
}

// AddIndex adds an Index to the builder.
func (b *MultilineValueBuilder) AddIndex(i Index) {
	if b.index.Start == 0 && b.index.Stop == 0 && b.indices == nil {
		b.index = i
		return
	}

	if b.indices == nil {
		b.indices = make([]Index, 0, 8)
		b.indices = append(b.indices, b.index)
		b.index = Index{}
	}
	b.indices = append(b.indices, i)
}

// AddSegment adds an Index derived from the given Segment to the builder.
func (b *MultilineValueBuilder) AddSegment(seg Segment) {
	b.AddIndex(NewIndexFromSegment(seg))
}

// SetString sets an owned string value. When Build is called, the resulting
// MultilineValue will be backed by this string rather than any accumulated indices.
func (b *MultilineValueBuilder) SetString(s string) {
	b.s = s
}

func (b *MultilineValueBuilder) isSingle() bool {
	return b.indices == nil && b.s == ""
}

// IsCollection returns true if the builder contains multiple Indices.
func (b *MultilineValueBuilder) IsCollection() bool {
	return !b.isSingle() && !b.IsOwned()
}

// IsOwned returns true if the builder contains an owned string value.
func (b *MultilineValueBuilder) IsOwned() bool {
	return b.s != ""
}

// Collection returns the slice of Index values in the builder. The result is
// meaningful only when IsCollection() returns true.
func (b *MultilineValueBuilder) Collection() []Index {
	return b.indices
}

// Build returns a MultilineValue from the accumulated state.
// If SetString was called, the result is backed by that string.
// Otherwise, the result is backed by the accumulated indices.
func (b *MultilineValueBuilder) Build() MultilineValue {
	if b.s != "" {
		return NewStringMultilineValue(b.s)
	}
	if b.isSingle() {
		return NewIndexMultilineValue(b.index)
	}
	return NewIndicesMultilineValue(b.indices)
}

// A Lines value holds the raw content of a block node.  It is either an
// owned string or a sequence of Segment values copied from the parser's
// source-segment list.
//
// Lines is used for block nodes whose rendered content is taken directly from the
// source (CodeBlock, FencedCodeBlock, HTMLBlock) so that the
// content is self-contained and can be serialised without access to the
// original source.
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
		return NewStringLines(val)
	case []byte:
		return NewStringLines(util.BytesToReadOnlyString(val))
	case []Segment:
		return NewSegmentsLines(val)
	case Lines:
		return val
	default:
		panic("unsupported type")
	}
}

// NewSegmentsLines returns a Lines backed by source segments.
func NewSegmentsLines(segs []Segment) Lines {
	return Lines{segs: segs}
}

// NewStringLines returns a Lines backed by an owned string.
func NewStringLines(s string) Lines {
	return Lines{s: s}
}

// Segments returns the internal segment slice, allowing callers to
// read and otherwise inspect the segment list.
func (l Lines) Segments() []Segment {
	return l.segs
}

// AppendSegment appends a Segment to the internal segment list.
func (l *Lines) AppendSegment(seg Segment) {
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
func NewSegmentPadding(start, stop, n int) Segment {
	return Segment{
		Start:   start,
		Stop:    stop,
		Padding: n,
	}
}

// Bytes returns bytes of the segment.
func (t Segment) Bytes(buffer []byte) []byte {
	var result []byte
	if t.Padding == 0 {
		result = buffer[t.Start:t.Stop]
	} else {
		result = make([]byte, 0, t.Padding+t.Stop-t.Start+1)
		result = append(result, bytes.Repeat(space, t.Padding)...)
		result = append(result, buffer[t.Start:t.Stop]...)
	}
	if t.ForceNewline && len(result) > 0 && result[len(result)-1] != '\n' {
		result = append(result, '\n')
	}
	return result
}

// Str returns a string of the segment.
func (t Segment) Str(buffer []byte) string {
	return util.BytesToReadOnlyString(t.Bytes(buffer))
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
func (t Segment) TrimRightSpace(buffer []byte) Segment {
	v := buffer[t.Start:t.Stop]
	l := util.TrimRightSpaceLength(v)
	if l == len(v) {
		return NewSegment(t.Start, t.Start)
	}
	return NewSegmentPadding(t.Start, t.Stop-l, t.Padding)
}

// TrimLeftSpace returns a new segment by slicing off all leading
// space characters including padding.
func (t Segment) TrimLeftSpace(buffer []byte) Segment {
	v := buffer[t.Start:t.Stop]
	l := util.TrimLeftSpaceLength(v)
	return NewSegment(t.Start+l, t.Stop)
}

// TrimLeftSpaceWidth returns a new segment by slicing off leading space
// characters until the given width.
func (t Segment) TrimLeftSpaceWidth(width int, buffer []byte) Segment {
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
	text := buffer[t.Start:t.Stop]
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

// WithStart returns a new Segment with same value except Start.
func (t Segment) WithStart(v int) Segment {
	return NewSegmentPadding(v, t.Stop, 0)
}

// WithStop returns a new Segment with same value except Stop.
func (t Segment) WithStop(v int) Segment {
	return NewSegmentPadding(t.Start, v, t.Padding)
}

func resolveEntityReferences(v []byte) ([]byte, bool) {
	if bytes.IndexByte(v, '&') == -1 {
		return v, false
	}
	addr := uintptr(unsafe.Pointer(&v[0]))
	v = util.ResolveNumericReferences(v)
	v = util.ResolveEntityNames(v)
	return v, addr != uintptr(unsafe.Pointer(&v[0]))
}
