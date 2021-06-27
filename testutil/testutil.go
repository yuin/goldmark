package testutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

// TestingT is a subset of the functionality provided by testing.T.
type TestingT interface {
	Logf(string, ...interface{})
	Skipf(string, ...interface{})
	Errorf(string, ...interface{})
	FailNow()
}

// MarkdownTestCase represents a test case.
type MarkdownTestCase struct {
	No          int
	Description string
	Markdown    string
	Expected    string
}

const attributeSeparator = "//- - - - - - - - -//"
const caseSeparator = "//= = = = = = = = = = = = = = = = = = = = = = = =//"

// ParseCliCaseArg parses -case command line args.
func ParseCliCaseArg() []int {
	ret := []int{}
	for _, a := range os.Args {
		if strings.HasPrefix(a, "case=") {
			parts := strings.Split(a, "=")
			for _, cas := range strings.Split(parts[1], ",") {
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
func DoTestCaseFile(m goldmark.Markdown, filename string, t TestingT, no ...int) {
	fp, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	c := MarkdownTestCase{
		No:          -1,
		Description: "",
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
			for _, n := range no {
				if n == c.No {
					shouldAdd = true
					break
				}
			}
		}
		if shouldAdd {
			cases = append(cases, c)
		}
	}
	DoTestCases(m, cases, t)
}

// DoTestCases runs a set of test cases.
func DoTestCases(m goldmark.Markdown, cases []MarkdownTestCase, t TestingT, opts ...parser.ParseOption) {
	for _, testCase := range cases {
		DoTestCase(m, testCase, t, opts...)
	}
}

// DoTestCase runs a test case.
func DoTestCase(m goldmark.Markdown, testCase MarkdownTestCase, t TestingT, opts ...parser.ParseOption) {
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
			t.Errorf(format, testCase.No, description, testCase.Markdown, testCase.Expected, err, debug.Stack())
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
			t.Errorf(format, testCase.No, description, testCase.Markdown, testCase.Expected, out.Bytes(),
				DiffPretty([]byte(testCase.Expected), out.Bytes()))
		}
	}()

	if err := m.Convert([]byte(testCase.Markdown), &out, opts...); err != nil {
		panic(err)
	}
	ok = bytes.Equal(bytes.TrimSpace(out.Bytes()), bytes.TrimSpace([]byte(testCase.Expected)))
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
func DiffPretty(v1, v2 []byte) []byte {
	var b bytes.Buffer
	diffs := simpleDiff(v1, v2)
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
				b.WriteString(fmt.Sprintf("%s | %s\n", c, util.VisualizeSpaces(line)))
			} else {
				b.WriteString(fmt.Sprintf("%s | %s\n", c, line))
			}
		}
	}
	return b.Bytes()
}
