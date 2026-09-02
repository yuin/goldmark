// Package parser contains stuff that are related to parsing a Markdown text.
package parser

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// A LinkDefinition interface represents a link definition in Markdown text.
type LinkDefinition interface {
	// String implements Stringer.
	String() string

	// Label returns a label of the link definition.
	Label() []byte

	// Destination returns a destination(URL) of the link definition.
	Destination() []byte

	// Title returns a title of the link definition.
	Title() []byte
}

type linkDefinition struct {
	node        *ast.LinkReferenceDefinition
	label       []byte
	destination []byte
	title       []byte
}

// NewLinkDefinition returns a new LinkDefinition.
func NewLinkDefinition(label, destination, title []byte) LinkDefinition {
	return &linkDefinition{nil, label, destination, title}
}

func newLinkDefinitionFromNode(v *ast.LinkReferenceDefinition, src []byte) LinkDefinition {
	return &linkDefinition{
		node:        v,
		label:       v.Label.Bytes(src),
		destination: v.Destination.Bytes(src),
		title:       v.Title.Bytes(src),
	}
}

func (r *linkDefinition) Label() []byte {
	return r.label
}

func (r *linkDefinition) Destination() []byte {
	return r.destination
}

func (r *linkDefinition) Title() []byte {
	return r.title
}

func (r *linkDefinition) String() string {
	return fmt.Sprintf("LinkDefinition{Label:%s, Destination:%s, Title:%s}", r.label, r.destination, r.title)
}

// An IDGenerator generates element IDs from a value and node kind.
// Implementations should return a base ID; uniqueness is handled by IDs.
type IDGenerator interface {
	// Generate generates a base element id for the given value and node kind.
	Generate(value []byte, kind ast.NodeKind) []byte
}

// IDs is a collection of element ids, tracking uniqueness across a parse.
type IDs struct {
	values    map[string]bool
	generator IDGenerator
}

type idsConfig struct {
	IDGenerator IDGenerator
}

// An IDsOption is an interface for options that can be passed to NewIDs.
type IDsOption interface {
	SetIDsOption(*idsConfig)
}

// NewIDs returns a new IDs.
// By default, the default IDGenerator is used.
// Use WithIDGenerator as an IDsOption to customize ID generation.
func NewIDs(opts ...IDsOption) *IDs {
	c := &idsConfig{IDGenerator: &defaultIDGenerator{}}
	for _, opt := range opts {
		opt.SetIDsOption(c)
	}
	return &IDs{
		values:    map[string]bool{},
		generator: c.IDGenerator,
	}
}

// Generate generates a unique element id for the given value and node kind.
// If the base id from the generator is already used, a numeric suffix is appended.
func (s *IDs) Generate(value []byte, kind ast.NodeKind) []byte {
	result := s.generator.Generate(value, kind)
	key := util.BytesToReadOnlyString(result)
	if _, ok := s.values[key]; !ok {
		s.values[key] = true
		return result
	}
	for i := 1; ; i++ {
		newResult := fmt.Sprintf("%s-%d", result, i)
		if _, ok := s.values[newResult]; !ok {
			s.values[newResult] = true
			return []byte(newResult)
		}
	}
}

// Put marks the given element id as used.
func (s *IDs) Put(value []byte) {
	s.values[util.BytesToReadOnlyString(value)] = true
}

type defaultIDGenerator struct{}

func (g *defaultIDGenerator) Generate(value []byte, kind ast.NodeKind) []byte {
	value = util.TrimLeftSpace(value)
	value = util.TrimRightSpace(value)
	result := make([]byte, 0, len(value))
	for i := 0; i < len(value); {
		v := value[i]
		l := util.UTF8Len(v)
		i += int(l)
		if l != 1 {
			continue
		}
		if util.IsAlphaNumeric(v) {
			if 'A' <= v && v <= 'Z' {
				v += 'a' - 'A'
			}
			result = append(result, v)
		} else if util.IsSpace(v) || v == '-' || v == '_' {
			result = append(result, '-')
		}
	}
	if len(result) == 0 {
		if kind == ast.KindHeading {
			return []byte("heading")
		}
		return []byte("id")
	}
	return result
}

// ContextKey is a key that is used to set arbitrary values to the context.
type ContextKey int

// ContextKeyMax is a maximum value of the ContextKey.
var ContextKeyMax ContextKey

// NewContextKey return a new ContextKey value.
func NewContextKey() ContextKey {
	ContextKeyMax++
	return ContextKeyMax
}

// A Context interface holds a information that are necessary to parse
// Markdown text.
type Context interface {
	// String implements Stringer.
	String() string

	// Get returns a value associated with the given key.
	Get(ContextKey) any

	// ComputeIfAbsent computes a value if a value associated with the given key is absent and returns the value.
	ComputeIfAbsent(ContextKey, func() any) any

	// Set sets the given value to the context.
	Set(ContextKey, any)

	// AddLinkDefinition adds the given link definition to this context.
	AddLinkDefinition(LinkDefinition)

	// LinkDefinition returns (a link definition, true) if a link definition associated with
	// the given label exists, otherwise (nil, false).
	LinkDefinition(label string) (LinkDefinition, bool)

	// LinkDefinitions returns a list of link definitions.
	LinkDefinitions() []LinkDefinition

	// IDs returns a collection of the element ids.
	IDs() *IDs

	// BlockOffset returns a first non-space character position on current line.
	// This value is valid only for BlockParser.Open.
	// BlockOffset returns -1 if there is no current line (EOF).
	BlockOffset() int

	// SetBlockOffset sets a first non-space character position on current line.
	// This value is valid only for BlockParser.Open.
	SetBlockOffset(int)

	// BlockIndent returns an indent width on current line.
	// This value is valid only for BlockParser.Open.
	// BlockIndent returns -1 if there is no current line (EOF).
	BlockIndent() int

	// SetBlockIndent sets an indent width on current line.
	// This value is valid only for BlockParser.Open.
	SetBlockIndent(int)

	// FirstDelimiter returns a first delimiter of the current delimiter list.
	FirstDelimiter() *Delimiter

	// LastDelimiter returns a last delimiter of the current delimiter list.
	LastDelimiter() *Delimiter

	// PushDelimiter appends the given delimiter to the tail of the current
	// delimiter list.
	PushDelimiter(delimiter *Delimiter)

	// RemoveDelimiter removes the given delimiter from the current delimiter list.
	RemoveDelimiter(d *Delimiter)

	// ClearDelimiters clears the current delimiter list.
	ClearDelimiters(bottom ast.Node)

	// OpenedBlocks returns a list of nodes that are currently in parsing.
	OpenedBlocks() []Block

	// SetOpenedBlocks sets a list of nodes that are currently in parsing.
	SetOpenedBlocks([]Block)

	// LastOpenedBlock returns a last node that is currently in parsing.
	LastOpenedBlock() Block

	// IsInLinkLabel returns true if current position seems to be in link label.
	IsInLinkLabel() bool
}

// A ContextOption is an interface for options that can be passed to NewContext.
type ContextOption interface {
	SetContextOption(*contextConfig)
}

type contextConfig struct {
	IDGenerator IDGenerator
}

type parseContext struct {
	store          []any
	ids            *IDs
	hasIDGenerator bool
	linkDefs       map[string]LinkDefinition
	blockOffset    int
	blockIndent    int
	delimiters     *Delimiter
	lastDelimiter  *Delimiter
	openedBlocks   []Block
}

// NewContext returns a new Context.
// By default, a new IDs with the default IDGenerator is used.
// Use WithIDGenerator as a ContextOption to customize ID generation.
func NewContext(opts ...ContextOption) Context {
	cc := &contextConfig{IDGenerator: nil}
	for _, opt := range opts {
		opt.SetContextOption(cc)
	}
	idGenerator := cc.IDGenerator
	if idGenerator == nil {
		idGenerator = &defaultIDGenerator{}
	}
	return &parseContext{
		store:          make([]any, ContextKeyMax+1),
		linkDefs:       map[string]LinkDefinition{},
		ids:            NewIDs(WithIDGenerator(idGenerator)),
		hasIDGenerator: cc.IDGenerator != nil,
		blockOffset:    -1,
		blockIndent:    -1,
		openedBlocks:   []Block{},
	}
}

func (p *parseContext) Get(key ContextKey) any {
	return p.store[key]
}

func (p *parseContext) ComputeIfAbsent(key ContextKey, f func() any) any {
	v := p.store[key]
	if v == nil {
		v = f()
		p.store[key] = v
	}
	return v
}

func (p *parseContext) Set(key ContextKey, value any) {
	p.store[key] = value
}

func (p *parseContext) IDs() *IDs {
	return p.ids
}

func (p *parseContext) BlockOffset() int {
	return p.blockOffset
}

func (p *parseContext) SetBlockOffset(v int) {
	p.blockOffset = v
}

func (p *parseContext) BlockIndent() int {
	return p.blockIndent
}

func (p *parseContext) SetBlockIndent(v int) {
	p.blockIndent = v
}

func (p *parseContext) LastDelimiter() *Delimiter {
	return p.lastDelimiter
}

func (p *parseContext) FirstDelimiter() *Delimiter {
	return p.delimiters
}

func (p *parseContext) PushDelimiter(d *Delimiter) {
	if p.delimiters == nil {
		p.delimiters = d
		p.lastDelimiter = d
	} else {
		l := p.lastDelimiter
		p.lastDelimiter = d
		l.NextDelimiter = d
		d.PreviousDelimiter = l
	}
}

func (p *parseContext) RemoveDelimiter(d *Delimiter) {
	if d.PreviousDelimiter == nil {
		p.delimiters = d.NextDelimiter
	} else {
		d.PreviousDelimiter.NextDelimiter = d.NextDelimiter
		if d.NextDelimiter != nil {
			d.NextDelimiter.PreviousDelimiter = d.PreviousDelimiter
		}
	}
	if d.NextDelimiter == nil {
		p.lastDelimiter = d.PreviousDelimiter
	}
	if p.delimiters != nil {
		p.delimiters.PreviousDelimiter = nil
	}
	if p.lastDelimiter != nil {
		p.lastDelimiter.NextDelimiter = nil
	}
	d.NextDelimiter = nil
	d.PreviousDelimiter = nil
	if d.Length != 0 {
		mergeOrReplaceTextSegment(d.Parent(), d, d.value, d.decoder)
	} else {
		d.Parent().RemoveChild(d)
	}
}

func (p *parseContext) ClearDelimiters(bottom ast.Node) {
	if p.lastDelimiter == nil {
		return
	}
	bottomDelim, _ := bottom.(*Delimiter)
	for c := p.lastDelimiter; c != nil && c != bottomDelim; {
		prev := c.PreviousDelimiter
		p.RemoveDelimiter(c)
		c = prev
	}
}

func (p *parseContext) AddLinkDefinition(ref LinkDefinition) {
	key := util.ToLinkReference(ref.Label())
	if _, ok := p.linkDefs[key]; !ok {
		p.linkDefs[key] = ref
	}
}

func (p *parseContext) LinkDefinition(label string) (LinkDefinition, bool) {
	v, ok := p.linkDefs[label]
	return v, ok
}

func (p *parseContext) LinkDefinitions() []LinkDefinition {
	ret := make([]LinkDefinition, 0, len(p.linkDefs))
	for _, v := range p.linkDefs {
		ret = append(ret, v)
	}
	return ret
}

func (p *parseContext) String() string {
	refs := []string{}
	for _, r := range p.linkDefs {
		refs = append(refs, r.String())
	}

	return fmt.Sprintf("Context{Store:%#v, LinkDefinitions:%s}", p.store, strings.Join(refs, ","))
}

func (p *parseContext) OpenedBlocks() []Block {
	return p.openedBlocks
}

func (p *parseContext) SetOpenedBlocks(v []Block) {
	p.openedBlocks = v
}

func (p *parseContext) LastOpenedBlock() Block {
	if l := len(p.openedBlocks); l != 0 {
		return p.openedBlocks[l-1]
	}
	return Block{}
}

func (p *parseContext) IsInLinkLabel() bool {
	tlist := p.Get(linkLabelStateKey)
	return tlist != nil
}

// State represents parser's state.
// State is designed to use as a bit flag.
type State int

const (
	// None is the zero value of [State], indicating that no flags are set.
	None State = 0

	// Continue indicates parser can continue parsing.
	Continue State = 1 << iota

	// Close indicates parser cannot parse anymore.
	Close

	// HasChildren indicates parser may have child blocks.
	HasChildren

	// NoChildren indicates parser does not have child blocks.
	NoChildren

	// RequireParagraph indicates parser requires that the last node
	// must be a paragraph and is not converted to other nodes by
	// ParagraphTransformers.
	RequireParagraph
)

// A Config struct is a data structure that holds configuration of the Parser.
type Config struct {
	// IDGenerator is a custom IDGenerator for element id generation.
	IDGenerator IDGenerator

	// EscapedSpace indicates that a '\' escaped half-space(0x20) should not trigger parsers.
	// This defaults to false unless enabled via [WithEscapedSpace] passed to [New].
	EscapedSpace bool

	// Attribute indicates that custom attributes are enabled.
	// This defaults to false unless enabled via [WithAttribute] passed to [New].
	Attribute bool

	withoutDefaultParsers bool

	autoHeadingID bool

	blockParsers          util.PrioritizedValues[BlockParser]
	inlineParsers         util.PrioritizedValues[InlineParser]
	paragraphTransformers util.PrioritizedValues[ParagraphTransformer]
	astTransformers       util.PrioritizedValues[ASTTransformer]
	extensions            []Extension
}

// An Option interface is a functional option type for the Parser.
type Option interface {
	SetParserOption(*Config)
}

type withAttribute struct{}

func (o *withAttribute) SetParserOption(c *Config) {
	c.Attribute = true
}

func (o *withAttribute) setHeadingOption(p *HeadingConfig) {
	p.attribute = true
}

// WithAttribute is a functional option that enables custom attributes.
// It can be used as a parser Option and a HeadingOption.
func WithAttribute() interface {
	Option
	HeadingOption
} {
	return &withAttribute{}
}

type withDefaultParsers struct {
	v bool
}

func (o *withDefaultParsers) SetParserOption(c *Config) {
	c.withoutDefaultParsers = !o.v
}

// WithDefaultParsers is a functional option that indicates whether default parsers should be used.
func WithDefaultParsers(v bool) Option {
	return &withDefaultParsers{v}
}

type withExtensions struct {
	value []Extension
}

func (o *withExtensions) SetParserOption(c *Config) {
	c.extensions = append(c.extensions, o.value...)
}

// WithExtensions is a functional option that allows you to add extensions to the parser.
func WithExtensions(ext ...Extension) Option {
	return &withExtensions{ext}
}

// Nil is a special AST node that represents an empty node.
// If a parser returns Nil, the parser is considered as successful but does not add any node to the AST tree.
var Nil = ast.NewText(text.NewSingleLineValueFromString("", nil))

// A Parser interface parses Markdown text into AST nodes.
type Parser interface {
	// Parse parses the given Markdown text into AST nodes.
	Parse(source []byte, opts ...ParseOption) ast.Node

	// ParseStringSource is a helper function that parses a string source into AST nodes using the given parser.
	//
	// This function converts the string source into a read-only byte slice without copying the data, and then
	// calls the Parse method of the provided parser.
	ParseStringSource(source string, opts ...ParseOption) ast.Node
}

// A BlockParser interface parses a block level element like Paragraph, List,
// Blockquote etc.
type BlockParser interface {
	// Trigger returns a list of characters that triggers the Open method of
	// this parser.
	// If Trigger returns a nil, Open will be called with any lines.
	Trigger() []byte

	// Open parses the current line and returns a result of parsing.
	//
	// Open must not parse beyond the current line.
	// If Open has been able to parse the current line, Open must advance a reader
	// position by consumed byte length.
	//
	// If Open has not been able to parse the current line, Open should returns
	// (nil, NoChildren). If Open has been able to parse the current line, Open
	// should returns a new Block node and returns HasChildren or NoChildren.
	Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)

	// Continue parses the current line and returns a result of parsing.
	//
	// Continue must not parse beyond the current line.
	// If Continue has been able to parse the current line, Continue must advance
	// a reader position by consumed byte length.
	//
	// If Continue has not been able to parse the current line, Continue should
	// returns Close. If Continue has been able to parse the current line,
	// Continue should returns (Continue | NoChildren) or
	// (Continue | HasChildren)
	Continue(node ast.Node, reader text.Reader, pc Context) State

	// Close will be called when the parser returns Close.
	Close(node ast.Node, reader text.Reader, pc Context)

	// CanInterruptParagraph returns true if the parser can interrupt paragraphs,
	// otherwise false.
	CanInterruptParagraph() bool

	// CanAcceptIndentedLine returns true if the parser can open new node when
	// the given line is being indented more than 3 spaces.
	CanAcceptIndentedLine() bool
}

// An InlineParser interface parses an inline level element like CodeSpan, Link etc.
type InlineParser interface {
	// Trigger returns a list of characters that triggers Parse method of
	// this parser.
	// Trigger characters must be a punctuation or a halfspace.
	// Halfspaces triggers this parser when character is any spaces characters or
	// a head of line
	Trigger() []byte

	// Parse parse the given block into an inline node.
	//
	// Parse can parse beyond the current line.
	// If Parse has been able to parse the current line, it must advance a reader
	// position by consumed byte length.
	Parse(parent ast.Node, block text.Reader, pc Context) ast.Node
}

// A CloseBlocker interface is a callback function that will be
// called when block is closed in the inline parsing.
type CloseBlocker interface {
	// CloseBlock will be called when a block is closed.
	CloseBlock(parent ast.Node, block text.Reader, pc Context)
}

// A ParagraphTransformer transforms parsed Paragraph nodes.
// For example, link references are searched in parsed Paragraphs.
type ParagraphTransformer interface {
	// Transform transforms the given paragraph.
	Transform(node *ast.Paragraph, reader text.Reader, pc Context)
}

// ASTTransformer transforms entire Markdown document AST tree.
type ASTTransformer interface {
	// Transform transforms the given AST tree.
	Transform(node *ast.Document, reader text.Reader, pc Context)
}

// A Block struct holds a node and correspond parser pair.
type Block struct {
	// Node is a BlockNode.
	Node ast.Node
	// Parser is a BlockParser.
	Parser BlockParser
}

type parser struct {
	blockParsers          [256][]BlockParser
	freeBlockParsers      []BlockParser
	inlineParsers         [256][]InlineParser
	closeBlockers         []CloseBlocker
	paragraphTransformers []ParagraphTransformer
	astTransformers       []ASTTransformer
	escapedSpace          bool
	idGenerator           IDGenerator
	decoder               text.Decoder
	config                *Config
	initSync              sync.Once
	triggerable           [256]bool
}

type withBlockParsers struct {
	value []util.PrioritizedValue[BlockParser]
}

func (o *withBlockParsers) SetParserOption(c *Config) {
	c.blockParsers = append(c.blockParsers, o.value...)
}

// WithBlockParsers is a functional option that allow you to add
// BlockParsers to the parser.
func WithBlockParsers(bs ...util.PrioritizedValue[BlockParser]) Option {
	return &withBlockParsers{bs}
}

type withInlineParsers struct {
	value []util.PrioritizedValue[InlineParser]
}

func (o *withInlineParsers) SetParserOption(c *Config) {
	c.inlineParsers = append(c.inlineParsers, o.value...)
}

// WithInlineParsers is a functional option that allow you to add
// InlineParsers to the parser.
func WithInlineParsers(is ...util.PrioritizedValue[InlineParser]) Option {
	return &withInlineParsers{is}
}

type withParagraphTransformers struct {
	value []util.PrioritizedValue[ParagraphTransformer]
}

func (o *withParagraphTransformers) SetParserOption(c *Config) {
	c.paragraphTransformers = append(c.paragraphTransformers, o.value...)
}

// WithParagraphTransformers is a functional option that allow you to add
// ParagraphTransformers to the parser.
func WithParagraphTransformers(ps ...util.PrioritizedValue[ParagraphTransformer]) Option {
	return &withParagraphTransformers{ps}
}

type withASTTransformers struct {
	value []util.PrioritizedValue[ASTTransformer]
}

func (o *withASTTransformers) SetParserOption(c *Config) {
	c.astTransformers = append(c.astTransformers, o.value...)
}

// WithASTTransformers is a functional option that allow you to add
// ASTTransformers to the parser.
func WithASTTransformers(ps ...util.PrioritizedValue[ASTTransformer]) Option {
	return &withASTTransformers{ps}
}

type withEscapedSpace struct {
}

func (o *withEscapedSpace) SetParserOption(c *Config) {
	c.EscapedSpace = true
}

// WithEscapedSpace is a functional option indicates that a '\' escaped half-space(0x20) should not trigger parsers.
func WithEscapedSpace() Option {
	return &withEscapedSpace{}
}

type withIDGenerator struct {
	gen IDGenerator
}

func (o *withIDGenerator) SetParserOption(c *Config) {
	c.IDGenerator = o.gen
}

func (o *withIDGenerator) SetContextOption(c *contextConfig) {
	c.IDGenerator = o.gen
}

func (o *withIDGenerator) SetIDsOption(c *idsConfig) {
	c.IDGenerator = o.gen
}

// WithIDGenerator is a functional option that sets a custom IDGenerator for element id generation.
// It can be used as a parser Option, a parser ContextOption, and a parser IDsOption.
func WithIDGenerator(gen IDGenerator) interface {
	Option
	ContextOption
	IDsOption
} {
	return &withIDGenerator{gen}
}

// New returns a new Parser with given options.
func New(options ...Option) Parser {
	config := &Config{}
	for _, opt := range options {
		opt.SetParserOption(config)
	}
	if !config.withoutDefaultParsers {
		for _, opt := range CommonMark.ParserOptions(config) {
			opt.SetParserOption(config)
		}
	}

	for _, ext := range config.extensions {
		options := ext.ParserOptions(config)
		for _, opt := range options {
			opt.SetParserOption(config)
		}
	}

	p := &parser{
		config: config,
	}

	return p
}

func (p *parser) AddOptions(opts ...Option) {
	for _, opt := range opts {
		opt.SetParserOption(p.config)
	}
}

func (p *parser) addBlockParser(v util.PrioritizedValue[BlockParser]) {
	bp := v.Value
	tcs := bp.Trigger()
	if tcs == nil {
		p.freeBlockParsers = append(p.freeBlockParsers, bp)
	} else {
		for _, tc := range tcs {
			if p.blockParsers[tc] == nil {
				p.blockParsers[tc] = []BlockParser{}
			}
			p.blockParsers[tc] = append(p.blockParsers[tc], bp)
		}
	}
}

func (p *parser) addInlineParser(v util.PrioritizedValue[InlineParser]) {
	ip := v.Value
	tcs := ip.Trigger()
	if cb, ok := ip.(CloseBlocker); ok {
		p.closeBlockers = append(p.closeBlockers, cb)
	}
	for _, tc := range tcs {
		if p.inlineParsers[tc] == nil {
			p.inlineParsers[tc] = []InlineParser{}
		}
		p.inlineParsers[tc] = append(p.inlineParsers[tc], ip)
	}
}

func (p *parser) addParagraphTransformer(v util.PrioritizedValue[ParagraphTransformer]) {
	pt := v.Value
	p.paragraphTransformers = append(p.paragraphTransformers, pt)
}

func (p *parser) addASTTransformer(v util.PrioritizedValue[ASTTransformer]) {
	at := v.Value
	p.astTransformers = append(p.astTransformers, at)
}

type parseConfig struct {
	context Context

	doPP   bool
	ppOpts []ast.PrettyPrintOption
}

// ParseOption is a functional option type for the Parse method.
type ParseOption func(*parseConfig)

// WithContext is a functional option that sets a custom Context for the Parse method.
func WithContext(ctx Context) ParseOption {
	return func(c *parseConfig) {
		c.context = ctx
	}
}

// WithPrettyPrint is a functional option that prints the AST tree to stdout for debugging purposes.
//
// If an io.Writer is provided, the AST tree will be printed to that writer instead of stdout.
func WithPrettyPrint(opts ...ast.PrettyPrintOption) ParseOption {
	return func(c *parseConfig) {
		c.doPP = true
		c.ppOpts = opts
	}
}

func (p *parser) Parse(source []byte, opts ...ParseOption) ast.Node {
	p.initSync.Do(func() {
		p.config.blockParsers.Sort()
		for _, v := range p.config.blockParsers {
			p.addBlockParser(v)
		}
		for i := range p.blockParsers {
			if p.blockParsers[i] != nil {
				p.blockParsers[i] = append(p.blockParsers[i], p.freeBlockParsers...)
			}
		}

		p.config.inlineParsers.Sort()
		for _, v := range p.config.inlineParsers {
			p.addInlineParser(v)
		}
		hasSpaceParser := p.inlineParsers[' '] != nil
		for c := range p.triggerable {
			flags := charFlags[c]
			if flags&charFlagPunct != 0 && p.inlineParsers[c] != nil {
				p.triggerable[c] = true
			} else if flags&charFlagSpace != 0 && hasSpaceParser {
				p.triggerable[c] = true
			}
		}
		p.config.paragraphTransformers.Sort()
		for _, v := range p.config.paragraphTransformers {
			p.addParagraphTransformer(v)
		}
		p.config.astTransformers.Sort()
		for _, v := range p.config.astTransformers {
			p.addASTTransformer(v)
		}
		p.escapedSpace = p.config.EscapedSpace
		p.idGenerator = p.config.IDGenerator
		if p.idGenerator == nil {
			p.idGenerator = &defaultIDGenerator{}
		}
		if p.config.EscapedSpace {
			p.decoder = text.NewDecoder(text.WithEscapedSpace())
		} else {
			p.decoder = text.NewDecoder()
		}
		p.config = nil
	})
	reader := text.NewReader(source, p.decoder)
	var pc Context
	var cfg parseConfig
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&cfg)
		}
		if cfg.context != nil {
			if c, ok := cfg.context.(*parseContext); ok {
				if !c.hasIDGenerator {
					c.ids = NewIDs(WithIDGenerator(p.idGenerator))
				}
				pc = cfg.context
			}
		}
	}
	if pc == nil {
		pc = NewContext(WithIDGenerator(p.idGenerator))
	}
	root := ast.NewDocument()
	p.parseBlocks(root, reader, pc)

	blockReader := text.NewBlockReader(reader.Source(), nil, p.decoder)
	p.walkBlock(root, func(node ast.Node) {
		p.parseBlock(blockReader, node, pc)
	})
	for _, at := range p.astTransformers {
		at.Transform(root, reader, pc)
	}
	if cfg.doPP {
		dump := root.Dump(reader.Source())
		_ = dump.PrettyPrint(os.Stdout, reader.Source(), cfg.ppOpts...)
	}

	return root
}

func (p *parser) ParseStringSource(source string, opts ...ParseOption) ast.Node {
	return p.Parse(util.StringToReadOnlyBytes(source), opts...)
}

func (p *parser) transformParagraph(node *ast.Paragraph, reader text.Reader, pc Context) bool {
	for _, pt := range p.paragraphTransformers {
		pt.Transform(node, reader, pc)
		if node.Parent() == nil {
			return true
		}
	}
	return false
}

func (p *parser) closeBlocks(from, to int, reader text.Reader, pc Context) {
	blocks := pc.OpenedBlocks()
	for i := from; i >= to; i-- {
		node := blocks[i].Node
		blocks[i].Parser.Close(blocks[i].Node, reader, pc)
		paragraph, ok := node.(*ast.Paragraph)
		if ok && node.Parent() != nil {
			p.transformParagraph(paragraph, reader, pc)
			continue
		}
	}
	if from == len(blocks)-1 {
		blocks = blocks[0:to]
	} else {
		blocks = append(blocks[0:to], blocks[from+1:]...)
	}
	pc.SetOpenedBlocks(blocks)
}

type blockOpenResult int

const (
	paragraphContinuation blockOpenResult = iota + 1
	newBlocksOpened
	noBlocksOpened
)

func (p *parser) openBlocks(parent ast.Node, blankLine bool, reader text.Reader, pc Context) blockOpenResult {
	result := blockOpenResult(noBlocksOpened)
	continuable := false
	lastBlock := pc.LastOpenedBlock()
	if lastBlock.Node != nil {
		continuable = ast.IsParagraph(lastBlock.Node)
	}
retry:
	var bps []BlockParser
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if len(line) == 0 {
		pc.SetBlockOffset(-1)
		pc.SetBlockIndent(-1)
	} else {
		pc.SetBlockOffset(pos)
		pc.SetBlockIndent(w)
	}

	if line == nil || line[0] == '\n' {
		goto continuable
	}
	bps = p.freeBlockParsers
	if pos < len(line) {
		bps = p.blockParsers[line[pos]]
		if bps == nil {
			bps = p.freeBlockParsers
		}
	}
	if bps == nil {
		goto continuable
	}

	for _, bp := range bps {
		if continuable && result == noBlocksOpened && !bp.CanInterruptParagraph() {
			continue
		}

		if w > 3 && !bp.CanAcceptIndentedLine() {
			continue
		}
		lastBlock = pc.LastOpenedBlock()
		last := lastBlock.Node
		_, blockPos := reader.Position()
		node, state := bp.Open(parent, reader, pc)
		if node != nil {
			node.SetPos(blockPos.Start + max(pc.BlockOffset(), 0))

			// Parser requires last node to be a paragraph.
			// With table extension:
			//
			//     0
			//     -:
			//     -
			//
			// '-' on 3rd line seems a Setext heading because 1st and 2nd lines
			// are being paragraph when the Setext heading parser tries to parse the 3rd
			// line.
			// But 1st line and 2nd line are a table. Thus this paragraph will be transformed
			// by a paragraph transformer. So this text should be converted to a table and
			// an empty list.
			if state&RequireParagraph != 0 {
				if last == parent.LastChild() {
					// Opened paragraph may be transformed by ParagraphTransformers in
					// closeBlocks().
					lastBlock.Parser.Close(last, reader, pc)
					blocks := pc.OpenedBlocks()
					pc.SetOpenedBlocks(blocks[0 : len(blocks)-1])
					if p.transformParagraph(last.(*ast.Paragraph), reader, pc) {
						// Paragraph has been transformed.
						// So this parser is considered as failing.
						continuable = false
						goto retry
					}
				}
			}
			node.(ast.BlockNode).SetBlankPreviousLines(blankLine)
			if last != nil && last.Parent() == nil {
				lastPos := len(pc.OpenedBlocks()) - 1
				p.closeBlocks(lastPos, lastPos, reader, pc)
			}
			parent.AppendChild(node)
			result = newBlocksOpened
			be := Block{node, bp}
			pc.SetOpenedBlocks(append(pc.OpenedBlocks(), be))
			if state&HasChildren != 0 {
				parent = node
				goto retry // try child block
			}
			break // no children, can not open more blocks on this line
		}
	}

continuable:
	if result == noBlocksOpened && continuable {
		state := lastBlock.Parser.Continue(lastBlock.Node, reader, pc)
		if state&Continue != 0 {
			result = paragraphContinuation
		}
	}
	return result
}

func isBlankLine(level int, prevLineBlank []bool) bool {
	if len(prevLineBlank) == 0 {
		return false
	}
	if level >= len(prevLineBlank) {
		level = len(prevLineBlank) - 1
	}
	return prevLineBlank[level]
}

func (p *parser) parseBlocks(parent ast.Node, reader text.Reader, pc Context) {
	pc.SetOpenedBlocks(nil)
	var prevLineBlank, curLineBlank []bool
	for { // process blocks separated by blank lines
		_, _, ok := reader.SkipBlankLines()
		if !ok {
			return
		}
		// first, we try to open blocks
		if p.openBlocks(parent, true, reader, pc) != newBlocksOpened {
			return
		}
		reader.AdvanceLine()
		prevLineBlank = prevLineBlank[:0]
		for { // process opened blocks line by line
			openedBlocks := pc.OpenedBlocks()
			l := len(openedBlocks)
			if l == 0 {
				break
			}
			lastIndex := l - 1
			curLineBlank = curLineBlank[:0]
			for i := range l {
				be := openedBlocks[i]
				line, _ := reader.PeekLine()
				if line == nil {
					p.closeBlocks(lastIndex, 0, reader, pc)
					reader.AdvanceLine()
					return
				}
				curLineBlank = append(curLineBlank, util.IsBlank(line))
				// If node is a paragraph, p.openBlocks determines whether it is continuable.
				// So we do not process paragraphs here.
				if !ast.IsParagraph(be.Node) {
					state := be.Parser.Continue(be.Node, reader, pc)
					if state&Continue != 0 {
						// When current node is a container block and has no children,
						// we try to open new child nodes
						if state&HasChildren != 0 && i == lastIndex {
							isBlank := isBlankLine(i+1, prevLineBlank)
							p.openBlocks(be.Node, isBlank, reader, pc)
							break
						}
						continue
					}
				}
				// current node may be closed or lazy continuation
				isBlank := isBlankLine(i, prevLineBlank)
				thisParent := parent
				if i != 0 {
					thisParent = openedBlocks[i-1].Node
				}
				lastNode := openedBlocks[lastIndex].Node
				result := p.openBlocks(thisParent, isBlank, reader, pc)
				if result != paragraphContinuation {
					// lastNode is a paragraph and was transformed by the paragraph
					// transformers.
					if openedBlocks[lastIndex].Node != lastNode {
						lastIndex--
					}
					p.closeBlocks(lastIndex, i, reader, pc)
				}
				break
			}

			reader.AdvanceLine()
			prevLineBlank, curLineBlank = curLineBlank, prevLineBlank
		}
	}
}

func (p *parser) walkBlock(block ast.Node, cb func(node ast.Node)) {
	for c := block.FirstChild(); c != nil; c = c.NextSibling() {
		p.walkBlock(c, cb)
	}
	cb(block)
}

const (
	lineBreakHard uint8 = 1 << iota
	lineBreakSoft
	lineBreakVisible
)

const (
	charFlagPunct uint8 = 1 << iota
	charFlagSpace
)

var charFlags = func() [256]uint8 {
	var t [256]uint8
	for i := range t {
		c := byte(i)
		if util.IsPunct(c) {
			t[i] |= charFlagPunct
		}
		if util.IsSpace(c) && c != '\r' && c != '\n' {
			t[i] |= charFlagSpace
		}
	}
	return t
}()

func (p *parser) parseBlock(block text.BlockReader, parent ast.Node, pc Context) {
	parentSource := parent.(ast.BlockNode).Source()
	if len(parentSource) == 0 {
		return
	}
	escaped := false
	source := block.Source()
	block.Reset(parentSource)
	for {
	retry:
		line, _ := block.PeekLine()
		if len(line) == 0 {
			break
		}
		lineLength := len(line)
		var lineBreakFlags uint8
		if line[lineLength-1] == '\n' {
			// end is the length of the line's content, excluding the
			// trailing newline sequence ("\n" or "\r\n").
			end := lineLength - 1
			if end > 0 && line[end-1] == '\r' {
				end--
			}
			switch {
			case end > 0 && line[end-1] == '\\' && (end < 2 || line[end-2] != '\\'):
				// ends with an unescaped backslash: a hard, visible line break.
				lineLength = end - 1
				lineBreakFlags = lineBreakHard | lineBreakVisible
			case end > 1 && line[end-1] == ' ' && line[end-2] == ' ':
				// ends with two or more trailing spaces: a hard line break.
				lineLength = end - 2
				lineBreakFlags = lineBreakHard
			default:
				// See https://spec.commonmark.org/0.30/#soft-line-breaks
				lineBreakFlags = lineBreakSoft
			}
		}

		l, startPosition := block.Position()
		n := 0
		for i := range lineLength {
			c := line[i]
			if c == '\n' {
				break
			}
			if i == 0 || p.triggerable[c] {
				flags := charFlags[c]
				isSpace := flags&charFlagSpace != 0
				isPunct := flags&charFlagPunct != 0
				if (isPunct && !escaped) || isSpace && (!escaped || !p.escapedSpace) || i == 0 {
					parserChar := c
					if isSpace || (i == 0 && !isPunct) {
						parserChar = ' '
					}
					ips := p.inlineParsers[parserChar]
					if ips != nil {
						block.Advance(n)
						n = 0
						savedLine, savedPosition := block.Position()
						if i != 0 {
							_, currentPosition := block.Position()
							mergeOrAppendTextSegment(parent, startPosition.Between(currentPosition), block.Decoder())
							_, startPosition = block.Position()
						}
						var inlineNode ast.Node
						for _, ip := range ips {
							inlineNode = ip.Parse(parent, block, pc)
							if inlineNode != nil {
								if inlineNode.Pos() < 0 {
									inlineNode.(interface{ SetPos(int) }).SetPos(startPosition.Start)
								}
								break
							}
							block.SetPosition(savedLine, savedPosition)
						}
						if inlineNode != nil {
							if inlineNode != Nil {
								parent.AppendChild(inlineNode)
							}
							goto retry
						}
					}
				}
			}
			if escaped {
				escaped = false
				n++
				continue
			}

			if c == '\\' {
				escaped = true
				n++
				continue
			}

			escaped = false
			n++
		}
		if n != 0 {
			block.Advance(n)
		}
		currentL, currentPosition := block.Position()
		if l != currentL {
			continue
		}
		diff := startPosition.Between(currentPosition)
		var t *ast.Text
		if lineBreakFlags&(lineBreakHard|lineBreakVisible) == lineBreakHard|lineBreakVisible {
			t = ast.NewText(text.NewSingleLineValueFromSegment(diff, block.Decoder()))
		} else {
			t = ast.NewText(text.NewSingleLineValueFromSegment(diff.TrimRightSpace(source), block.Decoder()))
		}
		t.SetSoftLineBreak(lineBreakFlags&lineBreakSoft != 0)
		t.SetHardLineBreak(lineBreakFlags&lineBreakHard != 0)
		parent.AppendChild(t)
		block.AdvanceLine()
	}

	ProcessDelimiters(nil, pc)
	for _, ip := range p.closeBlockers {
		ip.CloseBlock(parent, block, pc)
	}

}

// Extension is an interface that represents an extension for the parser.
type Extension interface {
	// ParserOptions returns a list of parser options to be applied to the parser.
	ParserOptions(*Config) []Option
}

type commonMark struct {
	opts []Option
}

// NewCommonMark returns a new CommonMark extension.
func NewCommonMark(opts ...Option) Extension {
	return &commonMark{opts}
}

func (e *commonMark) ParserOptions(cfg *Config) []Option {
	if len(e.opts) != 0 {
		thisConfig := *cfg
		for _, opt := range e.opts {
			opt.SetParserOption(&thisConfig)
		}
		cfg = &thisConfig
	}

	var hopts []HeadingOption
	if cfg.Attribute {
		hopts = append(hopts, WithAttribute())
	}
	if cfg.autoHeadingID {
		hopts = append(hopts, WithAutoHeadingID())
	}
	return []Option{
		WithBlockParsers(
			util.Prioritized(NewSetextHeadingParser(hopts...), 100),
			util.Prioritized(NewThematicBreakParser(), 200),
			util.Prioritized(NewListParser(), 300),
			util.Prioritized(NewListItemParser(), 400),
			util.Prioritized(NewCodeBlockParser(), 500),
			util.Prioritized(NewATXHeadingParser(hopts...), 600),
			util.Prioritized(NewFencedCodeBlockParser(), 700),
			util.Prioritized(NewBlockquoteParser(), 800),
			util.Prioritized(NewHTMLBlockParser(), 900),
			util.Prioritized(NewParagraphParser(), 1000),
		),
		WithInlineParsers(
			util.Prioritized(NewCodeSpanParser(), 100),
			util.Prioritized(NewLinkParser(), 200),
			util.Prioritized(NewAutoLinkParser(), 300),
			util.Prioritized(NewRawHTMLParser(), 400),
			util.Prioritized(NewEmphasisParser(), 500),
		),
		WithParagraphTransformers(
			util.Prioritized(LinkReferenceParagraphTransformer, 100),
		),
	}
}

func mergeOrReplaceTextSegment(parent ast.Node, n ast.Node, s text.Segment, decoder text.Decoder) {
	prev := n.PreviousSibling()
	if t, ok := prev.(*ast.Text); ok && !t.Value.IsOwned() && t.Value.Index().Stop == s.Start &&
		!t.SoftLineBreak() {
		t.Value = t.Value.WithStop(s.Stop)
		parent.RemoveChild(n)
	} else {
		parent.ReplaceChild(n, ast.NewText(text.NewSingleLineValueFromSegment(s, decoder)))
	}
}

func mergeOrAppendTextSegment(parent ast.Node, s text.Segment, decoder text.Decoder) {
	last := parent.LastChild()
	t, ok := last.(*ast.Text)
	if ok && !t.Value.IsOwned() && t.Value.Index().Stop == s.Start && !t.SoftLineBreak() {
		t.Value = t.Value.WithStop(s.Stop)
	} else {
		parent.AppendChild(ast.NewText(text.NewSingleLineValueFromSegment(s, decoder)))
	}
}

// CommonMark is a commonmark compliant extension.
// This extension adds default block parsers, inline parsers, and paragraph transformers to the parser.
//
// Block parsers:
//
//   - SetextHeadingParser, 100
//   - ThematicBreakParser, 200
//   - ListParser, 300
//   - ListItemParser, 400
//   - CodeBlockParser, 500
//   - ATXHeadingParser, 600
//   - FencedCodeBlockParser, 700
//   - BlockquoteParser, 800
//   - HTMLBlockParser, 900
//   - ParagraphParser, 1000
//
// Inline parsers:
//
//   - CodeSpanParser, 100
//   - LinkParser, 200
//   - AutoLinkParser, 300
//   - RawHTMLParser, 400
//   - EmphasisParser, 500
//
// Paragraph transformers:
//
//   - LinkReferenceParagraphTransformer, 100
var CommonMark = &commonMark{}
