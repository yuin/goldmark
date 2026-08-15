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
	Substitutions []string
}

func newDefaultSubstitutions() []string {
	replacements := make([]string, typographicPunctuationMax)
	replacements[LeftSingleQuote] = "&lsquo;"
	replacements[RightSingleQuote] = "&rsquo;"
	replacements[LeftDoubleQuote] = "&ldquo;"
	replacements[RightDoubleQuote] = "&rdquo;"
	replacements[EnDash] = "&ndash;"
	replacements[EmDash] = "&mdash;"
	replacements[Ellipsis] = "&hellip;"
	replacements[LeftAngleQuote] = "&laquo;"
	replacements[RightAngleQuote] = "&raquo;"
	replacements[Apostrophe] = "&rsquo;"

	return replacements
}

// TypographerParserOption is a functional option for the Typographer parser.
type TypographerParserOption func(*typographerConfig)

// TypographicSubstitutions is a list of the substitutions for the Typographer extension.
type TypographicSubstitutions map[TypographicPunctuation]string

// WithTypographicSubstitutions is a functional option that specify replacement text
// for punctuations.
func WithTypographicSubstitutions[T []byte | string](values map[TypographicPunctuation]T) TypographerParserOption {
	replacements := newDefaultSubstitutions()
	for k, v := range values {
		if len(v) == 0 {
			replacements[k] = ""
		} else {
			replacements[k] = string(v)
		}
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
			if len(s.Substitutions[EmDash]) != 0 && line[1] == '-' && line[2] == '-' { // ---
				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[EmDash], block.Decoder()))
				block.Advance(3)
				return node
			}
		case '.':
			if len(s.Substitutions[Ellipsis]) != 0 && line[1] == '.' && line[2] == '.' { // ...
				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[Ellipsis], block.Decoder()))
				block.Advance(3)
				return node
			}
			return nil
		}
	}
	if len(line) > 1 {
		switch c {
		case '<':
			if len(s.Substitutions[LeftAngleQuote]) != 0 && line[1] == '<' { // <<
				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[LeftAngleQuote], block.Decoder()))
				block.Advance(2)
				return node
			}
			return nil
		case '>':
			if len(s.Substitutions[RightAngleQuote]) != 0 && line[1] == '>' { // >>
				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[RightAngleQuote], block.Decoder()))
				block.Advance(2)
				return node
			}
			return nil
		case '-':
			if len(s.Substitutions[EnDash]) != 0 && line[1] == '-' { // --
				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[EnDash], block.Decoder()))
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
			if len(s.Substitutions[Apostrophe]) != 0 {
				// Handle decade abbrevations such as '90s
				if canOpen && !canClose && len(line) > 3 &&
					util.IsNumeric(line[1]) && util.IsNumeric(line[2]) && line[3] == 's' {
					after := rune(' ')
					if len(line) > 4 {
						after = util.ToRune(line, 4)
					}
					if len(line) == 3 || util.IsSpaceRune(after) || util.IsPunctRune(after) {
						node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[Apostrophe], block.Decoder()))
						block.Advance(1)
						return node
					}
				}
				// special cases: 'twas, 'em, 'net
				if len(line) > 1 && (unicode.IsPunct(before) || unicode.IsSpace(before)) &&
					(line[1] == 't' || line[1] == 'e' || line[1] == 'n' || line[1] == 'l') {
					node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[Apostrophe], block.Decoder()))
					block.Advance(1)
					return node
				}
				// Convert normal apostrophes. This is probably more flexible than necessary but
				// converts any apostrophe in between two alphanumerics.
				if len(line) > 1 && (unicode.IsDigit(before) || unicode.IsLetter(before)) &&
					(unicode.IsLetter(util.ToRune(line, 1))) {
					node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[Apostrophe], block.Decoder()))
					block.Advance(1)
					return node
				}
			}
			if len(s.Substitutions[LeftSingleQuote]) != 0 && canOpen && !canClose {
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

				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[nt], block.Decoder()))
				block.Advance(1)
				return node
			}
			if len(s.Substitutions[RightSingleQuote]) != 0 {
				// plural possesive and abbreviations: Smiths', doin'
				if len(line) > 1 && unicode.IsSpace(util.ToRune(line, 0)) || unicode.IsPunct(util.ToRune(line, 0)) &&
					(len(line) > 2 && !unicode.IsDigit(util.ToRune(line, 1))) {
					node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[RightSingleQuote], block.Decoder()))
					block.Advance(1)
					return node
				}
			}
			if len(s.Substitutions[RightSingleQuote]) != 0 && counter.Single > 0 {
				isClose := canClose && !canOpen
				maybeClose := canClose && canOpen && len(line) > 1 && unicode.IsPunct(util.ToRune(line, 1)) &&
					(len(line) == 2 || (len(line) > 2 && util.IsPunct(line[2]) || util.IsSpace(line[2])))
				if isClose || maybeClose {
					node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[RightSingleQuote], block.Decoder()))
					block.Advance(1)
					counter.Single--
					return node
				}
			}
		}
		if c == '"' {
			if len(s.Substitutions[LeftDoubleQuote]) != 0 && canOpen && !canClose {
				node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[LeftDoubleQuote], block.Decoder()))
				block.Advance(1)
				counter.Double++
				return node
			}
			if len(s.Substitutions[RightDoubleQuote]) != 0 && counter.Double > 0 {
				isClose := canClose && !canOpen
				maybeClose := canClose && canOpen && len(line) > 1 && (unicode.IsPunct(util.ToRune(line, 1))) &&
					(len(line) == 2 || (len(line) > 2 && util.IsPunct(line[2]) || util.IsSpace(line[2])))
				if isClose || maybeClose {
					// special case: "Monitor 21""
					if len(line) > 1 && line[1] == '"' && unicode.IsDigit(before) {
						return nil
					}
					node := gast.NewText(text.NewSingleLineValueFromString(s.Substitutions[RightDoubleQuote], block.Decoder()))
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
