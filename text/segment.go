package text

import (
	"bytes"
	"github.com/yuin/goldmark/util"
)

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

	PaddingChars []byte

	RenderPaddingTabs bool
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
func NewSegmentPadding(start, stop, n int, chars []byte) Segment {
	return Segment{
		Start:        start,
		Stop:         stop,
		Padding:      n,
		PaddingChars: chars,
	}
}

func (t Segment) WithRenderPaddingTabs() Segment {
	t.RenderPaddingTabs = true
	return t
}

// Value returns a value of the segment.
func (t *Segment) Value(buffer []byte) []byte {
	if t.RenderPaddingTabs {
		return t.ValueKeepTabs(buffer)
	}
	if t.Padding == 0 {
		return buffer[t.Start:t.Stop]
	}
	result := make([]byte, 0, t.Padding+t.Stop-t.Start+1)
	result = append(result, bytes.Repeat(space, t.Padding)...)
	return append(result, buffer[t.Start:t.Stop]...)
}

func (t *Segment) ValueKeepTabs(buffer []byte) []byte {
	if t.Padding == 0 {
		return buffer[t.Start:t.Stop]
	}
	result := make([]byte, 0, t.Padding+t.Stop-t.Start+1)
	result = append(result, t.PaddingChars...)
	return append(result, buffer[t.Start:t.Stop]...)
}

// Len returns a length of the segment.
func (t *Segment) Len() int {
	return t.Stop - t.Start + t.Padding
}

// Between returns a segment between this segment and the given segment.
func (t *Segment) Between(other Segment) Segment {
	if t.Stop != other.Stop {
		panic("invalid state")
	}
	return NewSegmentPadding(
		t.Start,
		other.Start,
		t.Padding-other.Padding,
		// ???? no idea what here, just put spaces there
		bytes.Repeat([]byte{' '}, t.Padding-other.Padding),
	)
}

// IsEmpty returns true if this segment is empty, otherwise false.
func (t *Segment) IsEmpty() bool {
	return t.Start >= t.Stop && t.Padding == 0
}

// TrimRightSpace returns a new segment by slicing off all trailing
// space characters.
func (t *Segment) TrimRightSpace(buffer []byte) Segment {
	v := buffer[t.Start:t.Stop]
	l := util.TrimRightSpaceLength(v)
	if l == len(v) {
		return NewSegment(t.Start, t.Start)
	}
	return NewSegmentPadding(t.Start, t.Stop-l, t.Padding, t.PaddingChars)
}

// TrimLeftSpace returns a new segment by slicing off all leading
// space characters including padding.
func (t *Segment) TrimLeftSpace(buffer []byte) Segment {
	v := buffer[t.Start:t.Stop]
	l := util.TrimLeftSpaceLength(v)
	return NewSegment(t.Start+l, t.Stop)
}

func trimWidthPaddingChars(origStartPos int, cut, goal int, chars []byte) []byte {
	bytesPos := origStartPos - len(chars)
	var i = 0
	for i < cut {
		if len(chars) == 0 {
			// ???
			return nil
		}
		b := chars[0]
		if b == ' ' {
			chars = chars[1:]
			i++
			bytesPos++
		} else {
			tw := util.TabWidth(bytesPos)
			chars = chars[1:]
			i += tw
			bytesPos++
		}
	}
	// if I can cut exactly, return the cut chars, otherwise just give up and put spaces
	if i == cut {
		return chars
	} else {
		return bytes.Repeat([]byte{' '}, goal)
	}
}

// TrimLeftSpaceWidth returns a new segment by slicing off leading space
// characters until the given width.
func (t *Segment) TrimLeftSpaceWidth(width int, buffer []byte) Segment {
	padding := t.Padding
	origWidth := width
	for ; width > 0; width-- {
		if padding == 0 {
			break
		}
		padding--
	}
	if width == 0 {
		paddingChars := trimWidthPaddingChars(t.Start, origWidth, padding, t.PaddingChars)
		return NewSegmentPadding(t.Start, t.Stop, padding, paddingChars)
	}
	newPaddingChars := []byte{}
	text := buffer[t.Start:t.Stop]
	start := t.Start
	for _, c := range text {
		if start >= t.Stop-1 || width <= 0 {
			break
		}
		if c == ' ' {
			width--
			if width < 0 {
				newPaddingChars = append(newPaddingChars, ' ')
			}
		} else if c == '\t' {
			width -= 4
			if width < 0 {
				newPaddingChars = append(newPaddingChars, '\t')
			}
		} else {
			break
		}
		start++
	}
	if width < 0 {
		padding = width * -1
		return NewSegmentPadding(start, t.Stop, padding, newPaddingChars)
	}
	paddingChars := trimWidthPaddingChars(t.Start, origWidth, padding, t.PaddingChars)
	return NewSegmentPadding(start, t.Stop, padding, paddingChars)
}

// WithStart returns a new Segment with same value except Start.
func (t *Segment) WithStart(v int) Segment {
	return NewSegmentPadding(v, t.Stop, t.Padding, t.PaddingChars)
}

// WithStop returns a new Segment with same value except Stop.
func (t *Segment) WithStop(v int) Segment {
	return NewSegmentPadding(t.Start, v, t.Padding, t.PaddingChars)
}

// ConcatPadding concats the padding to the given slice.
func (t *Segment) ConcatPadding(v []byte) []byte {
	if t.Padding > 0 {
		return append(v, bytes.Repeat(space, t.Padding)...)
	}
	return v
}

// Segments is a collection of the Segment.
type Segments struct {
	values []Segment
}

// NewSegments return a new Segments.
func NewSegments() *Segments {
	return &Segments{
		values: nil,
	}
}

// Append appends the given segment after the tail of the collection.
func (s *Segments) Append(t Segment) {
	if s.values == nil {
		s.values = make([]Segment, 0, 20)
	}
	s.values = append(s.values, t)
}

// AppendAll appends all elements of given segments after the tail of the collection.
func (s *Segments) AppendAll(t []Segment) {
	if s.values == nil {
		s.values = make([]Segment, 0, 20)
	}
	s.values = append(s.values, t...)
}

// Len returns the length of the collection.
func (s *Segments) Len() int {
	if s.values == nil {
		return 0
	}
	return len(s.values)
}

// At returns a segment at the given index.
func (s *Segments) At(i int) Segment {
	return s.values[i]
}

// Set sets the given Segment.
func (s *Segments) Set(i int, v Segment) {
	s.values[i] = v
}

// SetSliced replace the collection with a subsliced value.
func (s *Segments) SetSliced(lo, hi int) {
	s.values = s.values[lo:hi]
}

// Sliced returns a subslice of the collection.
func (s *Segments) Sliced(lo, hi int) []Segment {
	return s.values[lo:hi]
}

// Clear delete all element of the collection.
func (s *Segments) Clear() {
	s.values = nil
}

// Unshift insert the given Segment to head of the collection.
func (s *Segments) Unshift(v Segment) {
	s.values = append(s.values[0:1], s.values[0:]...)
	s.values[0] = v
}
