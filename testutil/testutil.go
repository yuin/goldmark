// Package testutil provides utilities for unit tests.
package testutil

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/util"
)

// TestingT is a subset of the functionality provided by testing.T.
type TestingT interface {
	Logf(string, ...any)
	Skipf(string, ...any)
	Errorf(string, ...any)
	FailNow()
}

// MarkdownToStringFunc is a function type that converts markdown to HTML.
type MarkdownToStringFunc func(source string) (string, error)

// NewMarkdownToStringFunc returns a MarkdownToStringFunc that uses the given parser and renderer.
func NewMarkdownToStringFunc(p parser.Parser, r renderer.Renderer[io.Writer]) MarkdownToStringFunc {
	return func(source string) (string, error) {
		var buf bytes.Buffer
		b := util.StringToReadOnlyBytes(source)
		doc := p.Parse(b)
		if err := r.Render(&buf, b, doc); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}

// MarkdownTestCase represents a test case.
type MarkdownTestCase struct {
	No          int
	Description string
	Options     MarkdownTestCaseOptions
	Markdown    string
	Expected    string
}

func source(t *MarkdownTestCase) string {
	ret := t.Markdown
	if t.Options.Trim {
		ret = strings.TrimSpace(ret)
	}
	if t.Options.EnableEscape {
		return string(applyEscapeSequence([]byte(ret)))
	}
	return ret
}

func expected(t *MarkdownTestCase) string {
	ret := t.Expected
	if t.Options.Trim {
		ret = strings.TrimSpace(ret)
	}
	if t.Options.EnableEscape {
		return string(applyEscapeSequence([]byte(ret)))
	}
	return ret
}

// MarkdownTestCaseOptions represents options for each test case.
type MarkdownTestCaseOptions struct {
	EnableEscape bool
	Trim         bool
}

const attributeSeparator = "//- - - - - - - - -//"
const caseSeparator = "//= = = = = = = = = = = = = = = = = = = = = = = =//"

var optionsRegexp = regexp.MustCompile(`(?i)\s*options:(.*)`)

// ParseCliCaseArg parses -case command line args.
func ParseCliCaseArg() []int {
	ret := []int{}
	for _, a := range os.Args {
		if strings.HasPrefix(a, "case=") {
			parts := strings.Split(a, "=")
			for cas := range strings.SplitSeq(parts[1], ",") {
				value, err := strconv.Atoi(strings.TrimSpace(cas))
				if err == nil {
					ret = append(ret, value)
				}
			}
		}
	}
	return ret
}

// DoTestCaseFile runs test cases in a given file.
func DoTestCaseFile(m MarkdownToStringFunc, filename string, t TestingT, no ...int) {
	fp, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = fp.Close()
	}()

	scanner := bufio.NewScanner(fp)
	c := MarkdownTestCase{
		No:          -1,
		Description: "",
		Options:     MarkdownTestCaseOptions{},
		Markdown:    "",
		Expected:    "",
	}
	cases := []MarkdownTestCase{}
	line := 0
	for scanner.Scan() {
		line++
		if util.IsBlank([]byte(scanner.Text())) {
			continue
		}
		header := scanner.Text()
		c.Description = ""
		if strings.Contains(header, ":") {
			parts := strings.Split(header, ":")
			c.No, err = strconv.Atoi(strings.TrimSpace(parts[0]))
			c.Description = strings.Join(parts[1:], ":")
		} else {
			c.No, err = strconv.Atoi(scanner.Text())
		}
		if err != nil {
			panic(fmt.Sprintf("%s: invalid case No at line %d", filename, line))
		}
		if !scanner.Scan() {
			panic(fmt.Sprintf("%s: invalid case at line %d", filename, line))
		}
		line++
		matches := optionsRegexp.FindAllStringSubmatch(scanner.Text(), -1)
		if len(matches) != 0 {
			err = json.Unmarshal([]byte(matches[0][1]), &c.Options)
			if err != nil {
				panic(fmt.Sprintf("%s: invalid options at line %d", filename, line))
			}
			scanner.Scan()
			line++
		}
		if scanner.Text() != attributeSeparator {
			panic(fmt.Sprintf("%s: invalid separator '%s' at line %d", filename, scanner.Text(), line))
		}
		buf := []string{}
		for scanner.Scan() {
			line++
			text := scanner.Text()
			if text == attributeSeparator {
				break
			}
			buf = append(buf, text)
		}
		c.Markdown = strings.Join(buf, "\n")
		buf = []string{}
		for scanner.Scan() {
			line++
			text := scanner.Text()
			if text == caseSeparator {
				break
			}
			buf = append(buf, text)
		}
		c.Expected = strings.Join(buf, "\n")
		if len(c.Expected) != 0 {
			c.Expected = c.Expected + "\n"
		}
		shouldAdd := len(no) == 0
		if !shouldAdd {
			shouldAdd = slices.Contains(no, c.No)
		}
		if shouldAdd {
			cases = append(cases, c)
		}
	}
	DoTestCases(m, cases, t)
}

// DoTestCases runs a set of test cases.
func DoTestCases(m MarkdownToStringFunc, cases []MarkdownTestCase, t TestingT) {
	for _, testCase := range cases {
		DoTestCase(m, testCase, t)
	}
}

// DoTestCase runs a test case.
func DoTestCase(m MarkdownToStringFunc, testCase MarkdownTestCase, t TestingT) {
	var ok bool
	var out bytes.Buffer
	defer func() {
		description := ""
		if len(testCase.Description) != 0 {
			description = ": " + testCase.Description
		}
		if err := recover(); err != nil {
			format := `============= case %d%s ================
Markdown:
-----------
%s

Expected:
----------
%s

Actual
---------
%v
%s
`
			t.Errorf(format, testCase.No, description, source(&testCase), expected(&testCase), err, debug.Stack())
		} else if !ok {
			format := `============= case %d%s ================
Markdown:
-----------
%s

Expected:
----------
%s

Actual
---------
%s

Diff
---------
%s
`
			t.Errorf(format, testCase.No, description, source(&testCase), expected(&testCase), out.Bytes(),
				DiffPretty(expected(&testCase), out.Bytes()))
		}
	}()

	result, err := m(source(&testCase))
	if err != nil {
		panic(err)
	}
	out.WriteString(result)
	ok = bytes.Equal(bytes.TrimSpace(out.Bytes()), bytes.TrimSpace([]byte(expected(&testCase))))
}

type diffType int

const (
	diffRemoved diffType = iota
	diffAdded
	diffNone
)

type diff struct {
	Type  diffType
	Lines [][]byte
}

func simpleDiff(v1, v2 []byte) []diff {
	return simpleDiffAux(
		bytes.Split(v1, []byte("\n")),
		bytes.Split(v2, []byte("\n")))
}

func simpleDiffAux(v1lines, v2lines [][]byte) []diff {
	v1index := map[string][]int{}
	for i, line := range v1lines {
		key := util.BytesToReadOnlyString(line)
		if _, ok := v1index[key]; !ok {
			v1index[key] = []int{}
		}
		v1index[key] = append(v1index[key], i)
	}
	overlap := map[int]int{}
	v1start := 0
	v2start := 0
	length := 0
	for v2pos, line := range v2lines {
		newOverlap := map[int]int{}
		key := util.BytesToReadOnlyString(line)
		if _, ok := v1index[key]; !ok {
			v1index[key] = []int{}
		}
		for _, v1pos := range v1index[key] {
			value := 0
			if v1pos != 0 {
				if v, ok := overlap[v1pos-1]; ok {
					value = v
				}
			}
			newOverlap[v1pos] = value + 1
			if newOverlap[v1pos] > length {
				length = newOverlap[v1pos]
				v1start = v1pos - length + 1
				v2start = v2pos - length + 1
			}
		}
		overlap = newOverlap
	}
	if length == 0 {
		diffs := []diff{}
		if len(v1lines) != 0 {
			diffs = append(diffs, diff{diffRemoved, v1lines})
		}
		if len(v2lines) != 0 {
			diffs = append(diffs, diff{diffAdded, v2lines})
		}
		return diffs
	}
	diffs := simpleDiffAux(v1lines[:v1start], v2lines[:v2start])
	diffs = append(diffs, diff{diffNone, v2lines[v2start : v2start+length]})
	diffs = append(diffs, simpleDiffAux(v1lines[v1start+length:],
		v2lines[v2start+length:])...)
	return diffs
}

// DiffPretty returns pretty formatted diff between given bytes.
func DiffPretty[T1 []byte | string, T2 []byte | string](v1 T1, v2 T2) []byte {
	var b bytes.Buffer
	diffs := simpleDiff([]byte(v1), []byte(v2))
	for _, diff := range diffs {
		c := " "
		switch diff.Type {
		case diffAdded:
			c = "+"
		case diffRemoved:
			c = "-"
		case diffNone:
			c = " "
		}
		for _, line := range diff.Lines {
			if c != " " {
				fmt.Fprintf(&b, "%s | %s\n", c, util.VisualizeSpaces(line))
			} else {
				fmt.Fprintf(&b, "%s | %s\n", c, line)
			}
		}
	}
	return b.Bytes()
}

func applyEscapeSequence(b []byte) []byte {
	result := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\\' && i != len(b)-1 {
			switch b[i+1] {
			case 'a':
				result = append(result, '\a')
				i++
				continue
			case 'b':
				result = append(result, '\b')
				i++
				continue
			case 'f':
				result = append(result, '\f')
				i++
				continue
			case 'n':
				result = append(result, '\n')
				i++
				continue
			case 'r':
				result = append(result, '\r')
				i++
				continue
			case 't':
				result = append(result, '\t')
				i++
				continue
			case 'v':
				result = append(result, '\v')
				i++
				continue
			case '\\':
				result = append(result, '\\')
				i++
				continue
			case 'x':
				if len(b) >= i+3 && util.IsHexDecimal(b[i+2]) && util.IsHexDecimal(b[i+3]) {
					v, _ := hex.DecodeString(string(b[i+2 : i+4]))
					result = append(result, v[0])
					i += 3
					continue
				}
			case 'u', 'U':
				if len(b) > i+2 {
					num := []byte{}
					for j := i + 2; j < len(b); j++ {
						if util.IsHexDecimal(b[j]) {
							num = append(num, b[j])
							continue
						}
						break
					}
					if len(num) >= 4 && len(num) < 8 {
						v, _ := strconv.ParseInt(string(num[:4]), 16, 32)
						result = append(result, []byte(string(rune(v)))...)
						i += 5
						continue
					}
					if len(num) >= 8 {
						v, _ := strconv.ParseInt(string(num[:8]), 16, 32)
						result = append(result, []byte(string(rune(v)))...)
						i += 9
						continue
					}
				}
			}
		}
		result = append(result, b[i])
	}
	return result
}
