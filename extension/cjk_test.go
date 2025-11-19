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

	// Tests with EastAsianLineBreaksStyleSimple (for Chinese)
	markdown = goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			NewCJK(WithEastAsianLineBreaks(EastAsianLineBreaksSimple)),
		),
	)
	no = 20
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "中文汉字之间的软回车应被忽略",
			Markdown:    "被分开成两行\n写的一句话。",
			Expected:    "<p>被分开成两行写的一句话。</p>",
		},
		t,
	)
	no = 21
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "中文常用标点符号之间的软回车应被忽略（换行符前）",
			Markdown:    "一，\n二。\n三？\n四！\n五：\n六；\n七。\n八【\n九】\n十『\n九』\n八‘\n七’\n六“\n五”\n四……\n三、\n二",
			Expected:    "<p>一，二。三？四！五：六；七。八【九】十『九』八‘七’六“五”四……三、二</p>",
		},
		t,
	)

	// 注意：按照中文标点符号规范，中文标点符号与其后的英文字母之间不应该有空格。
	// 但是目前的实现不支持这种判断，所以还是会有空格。
	no = 22
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          no,
			Description: "中文与英文混合",
			Markdown:    "一，\na\n二。\nb",
			Expected:    "<p>一，\na\n二。\nb</p>",
		},
		t,
	)
}
