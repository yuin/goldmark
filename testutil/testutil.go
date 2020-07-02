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
func DoTestCases(m goldmark.Markdown, cases []MarkdownTestCase, t TestingT) {
	for _, testCase := range cases {
		DoTestCase(m, testCase, t)
	}
}

// DoTestCase runs a test case.
func DoTestCase(m goldmark.Markdown, testCase MarkdownTestCase, t TestingT) {
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
`
			t.Errorf(format, testCase.No, description, testCase.Markdown, testCase.Expected, out.Bytes())
		}
	}()

	if err := m.Convert([]byte(testCase.Markdown), &out); err != nil {
		panic(err)
	}
	ok = bytes.Equal(bytes.TrimSpace(out.Bytes()), bytes.TrimSpace([]byte(testCase.Expected)))
}
