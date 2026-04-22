package extension

import (
	"regexp"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func TestLinkify(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewLinkifyParser()),
		),
		html.New(
			html.WithUnsafe(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/linkify.txt", t, testutil.ParseCliCaseArg()...)
}

func TestLinkifyWithAllowedProtocols(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewLinkifyParser(
				WithAllowedProtocols([]string{
					"ssh:",
				}),
				WithURLRegexp(
					regexp.MustCompile(`\w+://[^\s]+`),
				),
			)),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:       1,
			Markdown: `hoge ssh://user@hoge.com. http://example.com/`,
			Expected: `<p>hoge <a href="ssh://user@hoge.com">ssh://user@hoge.com</a>. http://example.com/</p>`,
		},
		t,
	)
}

func TestLinkifyWithWWWRegexp(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewLinkifyParser(
				WithWWWRegexp(
					regexp.MustCompile(`www\.example\.com`),
				),
			)),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:       1,
			Markdown: `www.google.com www.example.com`,
			Expected: `<p>www.google.com <a href="http://www.example.com">www.example.com</a></p>`,
		},
		t,
	)
}

func TestLinkifyWithEmailRegexp(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewLinkifyParser(
				WithEmailRegexp(
					regexp.MustCompile(`user@example\.com`),
				),
			)),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:       1,
			Markdown: `hoge@example.com user@example.com`,
			Expected: `<p>hoge@example.com <a href="mailto:user@example.com">user@example.com</a></p>`,
		},
		t,
	)
}
