package extension

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
	"github.com/yuin/goldmark/text"
)

func TestASTBlockNodeText(t *testing.T) {
	var cases = []struct {
		Name   string
		Source string
		T1     string
		T2     string
		C      bool
	}{
		{
			Name: "DefinitionList",
			Source: `c1
:   c2
    c3

a

c4
:   c5
    c6`,
			T1: `c1c2
c3`,
			T2: `c4c5
c6`,
		},
		{
			Name: "Table",
			Source: `| h1 | h2 |
| -- | -- |
| c1 | c2 |

a


| h3 | h4 |
| -- | -- |
| c3 | c4 |`,

			T1: `h1h2c1c2`,
			T2: `h3h4c3c4`,
		},
	}

	for _, cs := range cases {
		t.Run(cs.Name, func(t *testing.T) {
			s := []byte(cs.Source)
			md := goldmark.New(
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
				goldmark.WithExtensions(
					DefinitionList,
					Table,
				),
			)
			n := md.Parser().Parse(text.NewReader(s))
			c1 := n.FirstChild()
			c2 := c1.NextSibling().NextSibling()
			if cs.C {
				c1 = c1.FirstChild()
				c2 = c2.FirstChild()
			}
			if !bytes.Equal(c1.Text(s), []byte(cs.T1)) { // nolint: staticcheck

				t.Errorf("%s unmatch:\n%s", cs.Name, testutil.DiffPretty(c1.Text(s), []byte(cs.T1))) // nolint: staticcheck

			}
			if !bytes.Equal(c2.Text(s), []byte(cs.T2)) { // nolint: staticcheck

				t.Errorf("%s(EOF) unmatch: %s", cs.Name, testutil.DiffPretty(c2.Text(s), []byte(cs.T2))) // nolint: staticcheck

			}
		})
	}

}

func TestASTInlineNodeText(t *testing.T) {
	var cases = []struct {
		Name   string
		Source string
		T1     string
	}{
		{
			Name:   "Strikethrough",
			Source: `~c1 *c2*~`,
			T1:     `c1 c2`,
		},
	}

	for _, cs := range cases {
		t.Run(cs.Name, func(t *testing.T) {
			s := []byte(cs.Source)
			md := goldmark.New(
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
				goldmark.WithExtensions(
					Strikethrough,
				),
			)
			n := md.Parser().Parse(text.NewReader(s))
			c1 := n.FirstChild().FirstChild()
			if !bytes.Equal(c1.Text(s), []byte(cs.T1)) { // nolint: staticcheck

				t.Errorf("%s unmatch:\n%s", cs.Name, testutil.DiffPretty(c1.Text(s), []byte(cs.T1))) // nolint: staticcheck

			}
		})
	}

}
