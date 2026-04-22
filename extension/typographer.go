package extension

import (
	"unicode"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var uncloseCounterKey = parser.NewContextKey()

type unclosedCounter struct {
	Single int
	Double int
}

func (u *unclosedCounter) Reset() {
	u.Single = 0
	u.Double = 0
}

func getUnclosedCounter(pc parser.Context) *unclosedCounter {
	v := pc.Get(uncloseCounterKey)
	if v == nil {
		v = &unclosedCounter{}
		pc.Set(uncloseCounterKey, v)
	}
	return v.(*unclosedCounter)
}

// TypographicPunctuation is a key of the punctuations that can be replaced with
// typographic entities.
type TypographicPunctuation int

const (
	// LeftSingleQuote is ' .
	LeftSingleQuote TypographicPunctuation = iota + 1
	// RightSingleQuote is ' .
	RightSingleQuote
	// LeftDoubleQuote is " .
	LeftDoubleQuote
	// RightDoubleQuote is " .
	RightDoubleQuote
	// EnDash is -- .
	EnDash
	// EmDash is --- .
	EmDash
	// Ellipsis is ... .
	Ellipsis
	// LeftAngleQuote is << .
	LeftAngleQuote
	// RightAngleQuote is >> .
	RightAngleQuote
	// Apostrophe is ' .
	Apostrophe

	typographicPunctuationMax
)

type typographerConfig struct {
	Substitutions [][]byte
}

func newDefaultSubstitutions() [][]byte {
	replacements := make([][]byte, typographicPunctuationMax)
	replacements[LeftSingleQuote] = []byte("&lsquo;")
	replacements[RightSingleQuote] = []byte("&rsquo;")
	replacements[LeftDoubleQuote] = []byte("&ldquo;")
	replacements[RightDoubleQuote] = []byte("&rdquo;")
	replacements[EnDash] = []byte("&ndash;")
	replacements[EmDash] = []byte("&mdash;")
	replacements[Ellipsis] = []byte("&hellip;")
	replacements[LeftAngleQuote] = []byte("&laquo;")
	replacements[RightAngleQuote] = []byte("&raquo;")
	replacements[Apostrophe] = []byte("&rsquo;")

	return replacements
}

// TypographerParserOption is a functional option for the Typographer parser.
type TypographerParserOption func(*typographerConfig)

// TypographicSubstitutions is a list of the substitutions for the Typographer extension.
type TypographicSubstitutions map[TypographicPunctuation][]byte

// WithTypographicSubstitutions is a functional option that specify replacement text
// for punctuations.
func WithTypographicSubstitutions[T []byte | string](values map[TypographicPunctuation]T) TypographerParserOption {
	replacements := newDefaultSubstitutions()
	for k, v := range values {
		replacements[k] = []byte(v)
	}
	return func(p *typographerConfig) {
		p.Substitutions = replacements
	}
}

type typographerParser struct {
	typographerConfig
}

func newTypographerParser(opts ...TypographerParserOption) parser.InlineParser {
	p := &typographerParser{
		typographerConfig: typographerConfig{
			Substitutions: newDefaultSubstitutions(),
		},
	}
	for _, o := range opts {
		o(&p.typographerConfig)
	}
	return p
}

func (s *typographerParser) Trigger() []byte {
	return []byte{'\'', '"', '-', '.', ',', '<', '>', '*', '['}
}

func (s *typographerParser) Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
	c := line[0]
	if len(line) > 2 {
		switch c {
		case '-':
			if s.Substitutions[EmDash] != nil && line[1] == '-' && line[2] == '-' { // ---
				node := gast.NewStringText(string(s.Substitutions[EmDash]))
				block.Advance(3)
				return node
			}
		case '.':
			if s.Substitutions[Ellipsis] != nil && line[1] == '.' && line[2] == '.' { // ...
				node := gast.NewStringText(string(s.Substitutions[Ellipsis]))
				block.Advance(3)
				return node
			}
			return nil
		}
	}
	if len(line) > 1 {
		switch c {
		case '<':
			if s.Substitutions[LeftAngleQuote] != nil && line[1] == '<' { // <<
				node := gast.NewStringText(string(s.Substitutions[LeftAngleQuote]))
				block.Advance(2)
				return node
			}
			return nil
		case '>':
			if s.Substitutions[RightAngleQuote] != nil && line[1] == '>' { // >>
				node := gast.NewStringText(string(s.Substitutions[RightAngleQuote]))
				block.Advance(2)
				return node
			}
			return nil
		case '-':
			if s.Substitutions[EnDash] != nil && line[1] == '-' { // --
				node := gast.NewStringText(string(s.Substitutions[EnDash]))
				block.Advance(2)
				return node
			}
		}
	}
	if c == '\'' || c == '"' {
		before := block.PrecedingCharacter()
		after := rune(' ')
		if len(line) > 1 {
			after = util.ToRune(line, 1)
		}
		canOpen := parser.IsLeftFlankingDelimiterRun(before, after)
		canClose := parser.IsRightFlankingDelimiterRun(before, after)
		if !canOpen && !canClose {
			return nil
		}
		counter := getUnclosedCounter(pc)
		if c == '\'' {
			if s.Substitutions[Apostrophe] != nil {
				// Handle decade abbrevations such as '90s
				if canOpen && !canClose && len(line) > 3 &&
					util.IsNumeric(line[1]) && util.IsNumeric(line[2]) && line[3] == 's' {
					after := rune(' ')
					if len(line) > 4 {
						after = util.ToRune(line, 4)
					}
					if len(line) == 3 || util.IsSpaceRune(after) || util.IsPunctRune(after) {
						node := gast.NewStringText(string(s.Substitutions[Apostrophe]))
						block.Advance(1)
						return node
					}
				}
				// special cases: 'twas, 'em, 'net
				if len(line) > 1 && (unicode.IsPunct(before) || unicode.IsSpace(before)) &&
					(line[1] == 't' || line[1] == 'e' || line[1] == 'n' || line[1] == 'l') {
					node := gast.NewStringText(string(s.Substitutions[Apostrophe]))
					block.Advance(1)
					return node
				}
				// Convert normal apostrophes. This is probably more flexible than necessary but
				// converts any apostrophe in between two alphanumerics.
				if len(line) > 1 && (unicode.IsDigit(before) || unicode.IsLetter(before)) &&
					(unicode.IsLetter(util.ToRune(line, 1))) {
					node := gast.NewStringText(string(s.Substitutions[Apostrophe]))
					block.Advance(1)
					return node
				}
			}
			if s.Substitutions[LeftSingleQuote] != nil && canOpen && !canClose {
				nt := LeftSingleQuote
				// special cases: Alice's, I'm, Don't, You'd
				if len(line) > 1 && (line[1] == 's' || line[1] == 'm' || line[1] == 't' || line[1] == 'd') &&
					(len(line) < 3 || util.IsPunct(line[2]) || util.IsSpace(line[2])) {
					nt = RightSingleQuote
				}
				// special cases: I've, I'll, You're
				if len(line) > 2 && ((line[1] == 'v' && line[2] == 'e') ||
					(line[1] == 'l' && line[2] == 'l') || (line[1] == 'r' && line[2] == 'e')) &&
					(len(line) < 4 || util.IsPunct(line[3]) || util.IsSpace(line[3])) {
					nt = RightSingleQuote
				}
				if nt == LeftSingleQuote {
					counter.Single++
				}

				node := gast.NewStringText(string(s.Substitutions[nt]))
				block.Advance(1)
				return node
			}
			if s.Substitutions[RightSingleQuote] != nil {
				// plural possesive and abbreviations: Smiths', doin'
				if len(line) > 1 && unicode.IsSpace(util.ToRune(line, 0)) || unicode.IsPunct(util.ToRune(line, 0)) &&
					(len(line) > 2 && !unicode.IsDigit(util.ToRune(line, 1))) {
					node := gast.NewStringText(string(s.Substitutions[RightSingleQuote]))
					block.Advance(1)
					return node
				}
			}
			if s.Substitutions[RightSingleQuote] != nil && counter.Single > 0 {
				isClose := canClose && !canOpen
				maybeClose := canClose && canOpen && len(line) > 1 && unicode.IsPunct(util.ToRune(line, 1)) &&
					(len(line) == 2 || (len(line) > 2 && util.IsPunct(line[2]) || util.IsSpace(line[2])))
				if isClose || maybeClose {
					node := gast.NewStringText(string(s.Substitutions[RightSingleQuote]))
					block.Advance(1)
					counter.Single--
					return node
				}
			}
		}
		if c == '"' {
			if s.Substitutions[LeftDoubleQuote] != nil && canOpen && !canClose {
				node := gast.NewStringText(string(s.Substitutions[LeftDoubleQuote]))
				block.Advance(1)
				counter.Double++
				return node
			}
			if s.Substitutions[RightDoubleQuote] != nil && counter.Double > 0 {
				isClose := canClose && !canOpen
				maybeClose := canClose && canOpen && len(line) > 1 && (unicode.IsPunct(util.ToRune(line, 1))) &&
					(len(line) == 2 || (len(line) > 2 && util.IsPunct(line[2]) || util.IsSpace(line[2])))
				if isClose || maybeClose {
					// special case: "Monitor 21""
					if len(line) > 1 && line[1] == '"' && unicode.IsDigit(before) {
						return nil
					}
					node := gast.NewStringText(string(s.Substitutions[RightDoubleQuote]))
					block.Advance(1)
					counter.Double--
					return node
				}
			}
		}
	}
	return nil
}

func (s *typographerParser) CloseBlock(_ gast.Node, pc parser.Context) {
	getUnclosedCounter(pc).Reset()
}

type typographerParserExtension struct {
	options []TypographerParserOption
}

// NewTypographerParser returns a new parser.Extension that replaces punctuations with typographic entities.
func NewTypographerParser(opts ...TypographerParserOption) parser.Extension {
	return &typographerParserExtension{
		options: opts,
	}
}

func (e *typographerParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(
			util.Prioritized(newTypographerParser(e.options...), 9999),
		),
	}
}
