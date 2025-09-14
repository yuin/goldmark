package extension

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
)

func TestEscapedSpace(t *testing.T) {
	markdown := goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	no := 1
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Without spaces around an emphasis started with east asian punctuations, it is not interpreted as an emphasis(as defined in CommonMark spec)",
			Markdown:    "太郎は**「こんにちわ」**と言った\nんです",
			Expected:    "<p>太郎は**「こんにちわ」**と言った\nんです</p>",
		},
		t,
	)

	no = 2
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "With spaces around an emphasis started with east asian punctuations, it is interpreted as an emphasis(but remains unnecessary spaces)",
			Markdown:    "太郎は **「こんにちわ」** と言った\nんです",
			Expected:    "<p>太郎は <strong>「こんにちわ」</strong> と言った\nんです</p>",
		},
		t,
	)

	// Enables EscapedSpace
	markdown = goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(NewCJK(WithEscapedSpace())),
	)

	no = 3
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "With spaces around an emphasis started with east asian punctuations,it is interpreted as an emphasis",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nんです",
			Expected:    "<p>太郎は<strong>「こんにちわ」</strong>と言った\nんです</p>",
		},
		t,
	)

	// ' ' triggers Linkify extension inline parser.
	// Escaped spaces should not trigger the inline parser.

	markdown = goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(
			NewCJK(WithEscapedSpace()),
			Linkify,
		),
	)

	no = 4
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Escaped space and linkfy extension",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nんです",
			Expected:    "<p>太郎は<strong>「こんにちわ」</strong>と言った\nんです</p>",
		},
		t,
	)
}

func TestEastAsianLineBreaks(t *testing.T) {
	markdown := goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	no := 1
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks are rendered as a newline, so some asian users will see it as an unnecessary space",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言った\nんです</p>",
		},
		t,
	)

	// Enables EastAsianLineBreaks

	markdown = goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(NewCJK(WithEastAsianLineBreaks())),
	)

	no = 2
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between east asian wide characters are ignored",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言ったんです</p>",
		},
		t,
	)

	no = 3
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between western characters are rendered as a newline",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言ったa\nbんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言ったa\nbんです</p>",
		},
		t,
	)

	no = 4
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between a western character and an east asian wide character are rendered as a newline",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言ったa\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言ったa\nんです</p>",
		},
		t,
	)

	no = 5
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between an east asian wide character and a western character are rendered as a newline",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nbんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言った\nbんです</p>",
		},
		t,
	)

	// WithHardWraps take precedence over WithEastAsianLineBreaks
	markdown = goldmark.New(goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(NewCJK(WithEastAsianLineBreaks())),
	)
	no = 6
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "WithHardWraps take precedence over WithEastAsianLineBreaks",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言った<br />\nんです</p>",
		},
		t,
	)

	// Tests with EastAsianLineBreaksStyleSimple
	markdown = goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(
			NewCJK(WithEastAsianLineBreaks()),
			Linkify,
		),
	)
	no = 7
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "WithEastAsianLineBreaks and linkfy extension",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\r\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言ったんです</p>",
		},
		t,
	)
	no = 8
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between east asian wide characters or punctuations are ignored",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と、\r\n言った\r\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と、言ったんです</p>",
		},
		t,
	)
	no = 9
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between an east asian wide character and a western character are ignored",
			Markdown:    "私はプログラマーです。\n東京の会社に勤めています。\nGoでWebアプリケーションを開発しています。",
			Expected:    "<p>私はプログラマーです。東京の会社に勤めています。\nGoでWebアプリケーションを開発しています。</p>",
		},
		t,
	)

	// Tests with EastAsianLineBreaksCSS3Draft
	markdown = goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(
			NewCJK(WithEastAsianLineBreaks(EastAsianLineBreaksCSS3Draft)),
		),
	)
	no = 10
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between a western character and an east asian wide character are ignored",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言ったa\nんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言ったaんです</p>",
		},
		t,
	)

	no = 11
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between an east asian wide character and a western character are ignored",
			Markdown:    "太郎は\\ **「こんにちわ」**\\ と言った\nbんです",
			Expected:    "<p>太郎は\\ <strong>「こんにちわ」</strong>\\ と言ったbんです</p>",
		},
		t,
	)

	no = 12
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Soft line breaks between an east asian wide character and a western character are ignored",
			Markdown:    "私はプログラマーです。\n東京の会社に勤めています。\nGoでWebアプリケーションを開発しています。",
			Expected:    "<p>私はプログラマーです。東京の会社に勤めています。GoでWebアプリケーションを開発しています。</p>",
		},
		t,
	)
}

func TestCJKFriendlyEmphasis(t *testing.T) {
	markdown := goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(NewCJK(WithCJKFriendlyEmphasis())),
	)
	no := 1
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Only preceding character of the opening mark is a normal CJK character",
			Markdown:    "この**`code`**",
			Expected:    "<p>この<strong><code>code</code></strong></p>",
		},
		t,
	)
	no = 2
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Only following character of the opening mark is a normal CJK character",
			Markdown:    "John**「ハロー」**",
			Expected:    "<p>John<strong>「ハロー」</strong></p>",
		},
		t,
	)
	no = 3
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Only following character of the closing mark is a normal CJK character",
			Markdown:    "**`code`**を",
			Expected:    "<p><strong><code>code</code></strong>を</p>",
		},
		t,
	)
	no = 4
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Only preceding character of the closing mark is a normal CJK character",
			Markdown:    "Git **（ギット）**Hub",
			Expected:    "<p>Git <strong>（ギット）</strong>Hub</p>",
		},
		t,
	)
	no = 5
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes preceding IVS",
			Markdown:    "禰󠄀**(ね)**豆子",
			Expected:    "<p>禰󠄀<strong>(ね)</strong>豆子</p>",
		},
		t,
	)
	no = 6
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the surrounding normal hangul character",
			Markdown:    "**스크립트(script)**라고",
			Expected:    "<p><strong>스크립트(script)</strong>라고</p>",
		},
		t,
	)
	no = 7
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the preceding hangul character that cannot be determined by East Asian Width",
			Markdown:    "ᅡ**(a)**",
			Expected:    "<p>ᅡ<strong>(a)</strong></p>",
		},
		t,
	)
	no = 8
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the following hangul character that cannot be determined by East Asian Width",
			Markdown:    "**(k)**ᄏ",
			Expected:    "<p><strong>(k)</strong>ᄏ</p>",
		},
		t,
	)
	no = 9
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the preceding han SVS",
			Markdown:    "大塚︀**(U+585A U+FE00)** 大塚**(U+FA10)**",
			Expected:    "<p>大塚︀<strong>(U+585A U+FE00)</strong> 大塚<strong>(U+FA10)</strong></p>",
		},
		t,
	)
	no = 10
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the preceding pseudo-emoji CJK symbol",
			Markdown:    "〽︎**(庵点)**は、",
			Expected:    "<p>〽︎<strong>(庵点)</strong>は、</p>",
		},
		t,
	)
	no = 11
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the preceding CJK ambiguous punctuation sequence (regression test)",
			Markdown:    "**“︁Git”︁**Hub",
			Expected:    "<p><strong>“︁Git”︁</strong>Hub</p>",
		},
		t,
	)
	no = 12
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the preceding CJK ambiguous punctuation sequence (underscore)",
			Markdown:    "“︁Git”︁__Hub__",
			Expected:    "<p>“︁Git”︁<strong>Hub</strong></p>",
		},
		t,
	)
}

func TestCJKFriendlyStrikethrough(t *testing.T) {
	markdown := goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(Strikethrough, NewCJK(WithCJKFriendlyEmphasis())),
	)
	no := 1
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Strikethrough is enabled",
			Markdown:    "~~No~~Yes",
			Expected:    "<p><del>No</del>Yes</p>",
		},
		t,
	)
	no = 2
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the preceding supplementary han character",
			Markdown:    "𩸽~~()a~~a",
			Expected:    "<p>𩸽<del>()a</del>a</p>",
		},
		t,
	)
	no = 3
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the following supplementary han character",
			Markdown:    "a~~a()~~𩸽",
			Expected:    "<p>a<del>a()</del>𩸽</p>",
		},
		t,
	)
}

func TestCJKFriendlyEmphasisWithEscapedSpace(t *testing.T) {
	markdown := goldmark.New(goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
		goldmark.WithExtensions(NewCJK(WithEscapedSpace(), WithCJKFriendlyEmphasis())),
	)
	no := 1
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "Recognizes the following supplementary han character",
			Markdown:    "a\\ **()**\\ a𩸽**()**𩸽",
			Expected:    "<p>a<strong>()</strong>a𩸽<strong>()</strong>𩸽</p>",
		},
		t,
	)
}
