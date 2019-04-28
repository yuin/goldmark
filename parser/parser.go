// Package parser contains stuff that are related to parsing a Markdown text.
package parser

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A Reference interface represents a link reference in Markdown text.
type Reference interface {
	// String implements Stringer.
	String() string

	// Label returns a label of the reference.
	Label() []byte

	// Destination returns a destination(URL) of the reference.
	Destination() []byte

	// Title returns a title of the reference.
	Title() []byte
}

type reference struct {
	label       []byte
	destination []byte
	title       []byte
}

// NewReference returns a new Reference.
func NewReference(label, destination, title []byte) Reference {
	return &reference{label, destination, title}
}

func (r *reference) Label() []byte {
	return r.label
}

func (r *reference) Destination() []byte {
	return r.destination
}

func (r *reference) Title() []byte {
	return r.title
}

func (r *reference) String() string {
	return fmt.Sprintf("Reference{Label:%s, Destination:%s, Title:%s}", r.label, r.destination, r.title)
}

// ContextKey is a key that is used to set arbitary values to the context.
type ContextKey int32

// New returns a new ContextKey value.
func (c *ContextKey) New() ContextKey {
	return ContextKey(atomic.AddInt32((*int32)(c), 1))
}

// ContextKeyMax is a maximum value of the ContextKey.
var ContextKeyMax ContextKey

// NewContextKey return a new ContextKey value.
func NewContextKey() ContextKey {
	return ContextKeyMax.New()
}

// A Context interface holds a information that are necessary to parse
// Markdown text.
type Context interface {
	// String implements Stringer.
	String() string

	// Source returns a source of Markdown text.
	Source() []byte

	// Get returns a value associated with given key.
	Get(ContextKey) interface{}

	// Set sets given value to the context.
	Set(ContextKey, interface{})

	// AddReference adds given reference to this context.
	AddReference(Reference)

	// Reference returns (a reference, true) if a reference associated with
	// given label exists, otherwise (nil, false).
	Reference(label string) (Reference, bool)

	// References returns a list of references.
	References() []Reference

	// BlockOffset returns a first non-space character position on current line.
	// This value is valid only for BlockParser.Open.
	BlockOffset() int

	// BlockOffset sets a first non-space character position on current line.
	// This value is valid only for BlockParser.Open.
	SetBlockOffset(int)

	// FirstDelimiter returns a first delimiter of the current delimiter list.
	FirstDelimiter() *Delimiter

	// LastDelimiter returns a last delimiter of the current delimiter list.
	LastDelimiter() *Delimiter

	// PushDelimiter appends given delimiter to the tail of the current
	// delimiter list.
	PushDelimiter(delimiter *Delimiter)

	// RemoveDelimiter removes given delimiter from the current delimiter list.
	RemoveDelimiter(d *Delimiter)

	// ClearDelimiters clears the current delimiter list.
	ClearDelimiters(bottom ast.Node)

	// OpenedBlocks returns a list of nodes that are currently in parsing.
	OpenedBlocks() []Block

	// SetOpenedBlocks sets a list of nodes that are currently in parsing.
	SetOpenedBlocks([]Block)

	// LastOpenedBlock returns a last node that is currently in parsing.
	LastOpenedBlock() Block

	// SetLastOpenedBlock sets a last node that is currently in parsing.
	SetLastOpenedBlock(Block)
}

type parseContext struct {
	store           []interface{}
	source          []byte
	refs            map[string]Reference
	blockOffset     int
	delimiters      *Delimiter
	lastDelimiter   *Delimiter
	openedBlocks    []Block
	lastOpenedBlock Block
}

// NewContext returns a new Context.
func NewContext(source []byte) Context {
	return &parseContext{
		store:           make([]interface{}, ContextKeyMax+1),
		source:          source,
		refs:            map[string]Reference{},
		blockOffset:     0,
		delimiters:      nil,
		lastDelimiter:   nil,
		openedBlocks:    []Block{},
		lastOpenedBlock: Block{},
	}
}

func (p *parseContext) Get(key ContextKey) interface{} {
	return p.store[key]
}

func (p *parseContext) Set(key ContextKey, value interface{}) {
	p.store[key] = value
}

func (p *parseContext) BlockOffset() int {
	return p.blockOffset
}

func (p *parseContext) SetBlockOffset(v int) {
	p.blockOffset = v
}

func (p *parseContext) Source() []byte {
	return p.source
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
		ast.MergeOrReplaceTextSegment(d.Parent(), d, d.Segment)
	} else {
		d.Parent().RemoveChild(d.Parent(), d)
	}
}

func (p *parseContext) ClearDelimiters(bottom ast.Node) {
	if p.lastDelimiter == nil {
		return
	}
	var c ast.Node
	for c = p.lastDelimiter; c != nil && c != bottom; {
		prev := c.PreviousSibling()
		if d, ok := c.(*Delimiter); ok {
			p.RemoveDelimiter(d)
		}
		c = prev
	}
}

func (p *parseContext) AddReference(ref Reference) {
	key := util.ToLinkReference(ref.Label())
	if _, ok := p.refs[key]; !ok {
		p.refs[key] = ref
	}
}

func (p *parseContext) Reference(label string) (Reference, bool) {
	v, ok := p.refs[label]
	return v, ok
}

func (p *parseContext) References() []Reference {
	ret := make([]Reference, 0, len(p.refs))
	for _, v := range p.refs {
		ret = append(ret, v)
	}
	return ret
}

func (p *parseContext) String() string {
	refs := []string{}
	for _, r := range p.refs {
		refs = append(refs, r.String())
	}

	return fmt.Sprintf("Context{Store:%#v, Refs:%s}", p.store, strings.Join(refs, ","))
}

func (p *parseContext) OpenedBlocks() []Block {
	return p.openedBlocks
}

func (p *parseContext) SetOpenedBlocks(v []Block) {
	p.openedBlocks = v
}

func (p *parseContext) LastOpenedBlock() Block {
	return p.lastOpenedBlock
}

func (p *parseContext) SetLastOpenedBlock(v Block) {
	p.lastOpenedBlock = v
}

// State represents parser's state.
// State is designed to use as a bit flag.
type State int

const (
	none State = 1 << iota

	// Continue indicates parser can continue parsing.
	Continue

	// Close indicates parser cannot parse anymore.
	Close

	// HasChildren indicates parser may have child blocks.
	HasChildren

	// NoChildren indicates parser does not have child blocks.
	NoChildren
)

// A Config struct is a data structure that holds configuration of the Parser.
type Config struct {
	Options               map[OptionName]interface{}
	BlockParsers          util.PrioritizedSlice /*<BlockParser>*/
	InlineParsers         util.PrioritizedSlice /*<InlineParser>*/
	ParagraphTransformers util.PrioritizedSlice /*<ParagraphTransformer>*/
	ASTTransformers       util.PrioritizedSlice /*<ASTTransformer>*/
}

// NewConfig returns a new Config.
func NewConfig() *Config {
	return &Config{
		Options:               map[OptionName]interface{}{},
		BlockParsers:          util.PrioritizedSlice{},
		InlineParsers:         util.PrioritizedSlice{},
		ParagraphTransformers: util.PrioritizedSlice{},
		ASTTransformers:       util.PrioritizedSlice{},
	}
}

// An Option interface is a functional option type for the Parser.
type Option interface {
	SetConfig(*Config)
}

// OptionName is a name of parser options.
type OptionName string

// A Parser interface parses Markdown text into AST nodes.
type Parser interface {
	// Parse parses given Markdown text into AST nodes.
	Parse(reader text.Reader, opts ...ParseOption) ast.Node

	// AddOption adds given option to thie parser.
	AddOption(Option)
}

// A SetOptioner interface sets given option to the object.
type SetOptioner interface {
	// SetOption sets given option to the object.
	// Unacceptable options may be passed.
	// Thus implementations must ignore unacceptable options.
	SetOption(name OptionName, value interface{})
}

// A BlockParser interface parses a block level element like Paragraph, List,
// Blockquote etc.
type BlockParser interface {
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
	Close(node ast.Node, pc Context)

	// CanInterruptParagraph returns true if the parser can interrupt pargraphs,
	// otherwise false.
	CanInterruptParagraph() bool

	// CanAcceptIndentedLine returns true if the parser can open new node when
	// given line is being indented more than 3 spaces.
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

	// Parse parse given block into an inline node.
	//
	// Parse can parse beyond the current line.
	// If Parse has been able to parse the current line, it must advance a reader
	// position by consumed byte length.
	Parse(parent ast.Node, block text.Reader, pc Context) ast.Node

	// CloseBlock will be called when a block is closed.
	CloseBlock(parent ast.Node, pc Context)
}

// A ParagraphTransformer transforms parsed Paragraph nodes.
// For example, link references are searched in parsed Paragraphs.
type ParagraphTransformer interface {
	// Transform transforms given paragraph.
	Transform(node *ast.Paragraph, pc Context)
}

// ASTTransformer transforms entire Markdown document AST tree.
type ASTTransformer interface {
	// Transform transforms given AST tree.
	Transform(node *ast.Document, pc Context)
}

// DefaultBlockParsers returns a new list of default BlockParsers.
// Priorities of default BlockParsers are:
//
//     SetextHeadingParser, 100
//     ThemanticBreakParser, 200
//     ListParser, 300
//     ListItemParser, 400
//     CodeBlockParser, 500
//     ATXHeadingParser, 600
//     FencedCodeBlockParser, 700
//     BlockquoteParser, 800
//     HTMLBlockParser, 900
//     ParagraphParser, 1000
func DefaultBlockParsers() []util.PrioritizedValue {
	return []util.PrioritizedValue{
		util.Prioritized(NewSetextHeadingParser(), 100),
		util.Prioritized(NewThemanticBreakParser(), 200),
		util.Prioritized(NewListParser(), 300),
		util.Prioritized(NewListItemParser(), 400),
		util.Prioritized(NewCodeBlockParser(), 500),
		util.Prioritized(NewATXHeadingParser(), 600),
		util.Prioritized(NewFencedCodeBlockParser(), 700),
		util.Prioritized(NewBlockquoteParser(), 800),
		util.Prioritized(NewHTMLBlockParser(), 900),
		util.Prioritized(NewParagraphParser(), 1000),
	}
}

// DefaultInlineParsers returns a new list of default InlineParsers.
// Priorities of default InlineParsers are:
//
//     CodeSpanParser, 100
//     LinkParser, 200
//     AutoLinkParser, 300
//     RawHTMLParser, 400
//     EmphasisParser, 500
func DefaultInlineParsers() []util.PrioritizedValue {
	return []util.PrioritizedValue{
		util.Prioritized(NewCodeSpanParser(), 100),
		util.Prioritized(NewLinkParser(), 200),
		util.Prioritized(NewAutoLinkParser(), 300),
		util.Prioritized(NewRawHTMLParser(), 400),
		util.Prioritized(NewEmphasisParser(), 500),
	}
}

// DefaultParagraphTransformers returns a new list of default ParagraphTransformers.
// Priorities of default ParagraphTransformers are:
//
//     LinkReferenceParagraphTransformer, 100
func DefaultParagraphTransformers() []util.PrioritizedValue {
	return []util.PrioritizedValue{
		util.Prioritized(LinkReferenceParagraphTransformer, 100),
	}
}

// A Block struct holds a node and correspond parser pair.
type Block struct {
	// Node is a BlockNode.
	Node ast.Node
	// Parser is a BlockParser.
	Parser BlockParser
}

type parser struct {
	options               map[OptionName]interface{}
	blockParsers          []BlockParser
	inlineParsers         [256][]InlineParser
	inlineParsersList     []InlineParser
	paragraphTransformers []ParagraphTransformer
	astTransformers       []ASTTransformer
	config                *Config
	initSync              sync.Once
}

type withBlockParsers struct {
	value []util.PrioritizedValue
}

func (o *withBlockParsers) SetConfig(c *Config) {
	c.BlockParsers = append(c.BlockParsers, o.value...)
}

// WithBlockParsers is a functional option that allow you to add
// BlockParsers to the parser.
func WithBlockParsers(bs ...util.PrioritizedValue) Option {
	return &withBlockParsers{bs}
}

type withInlineParsers struct {
	value []util.PrioritizedValue
}

func (o *withInlineParsers) SetConfig(c *Config) {
	c.InlineParsers = append(c.InlineParsers, o.value...)
}

// WithInlineParsers is a functional option that allow you to add
// InlineParsers to the parser.
func WithInlineParsers(bs ...util.PrioritizedValue) Option {
	return &withInlineParsers{bs}
}

type withParagraphTransformers struct {
	value []util.PrioritizedValue
}

func (o *withParagraphTransformers) SetConfig(c *Config) {
	c.ParagraphTransformers = append(c.ParagraphTransformers, o.value...)
}

// WithParagraphTransformers is a functional option that allow you to add
// ParagraphTransformers to the parser.
func WithParagraphTransformers(ps ...util.PrioritizedValue) Option {
	return &withParagraphTransformers{ps}
}

type withASTTransformers struct {
	value []util.PrioritizedValue
}

func (o *withASTTransformers) SetConfig(c *Config) {
	c.ASTTransformers = append(c.ASTTransformers, o.value...)
}

// WithASTTransformers is a functional option that allow you to add
// ASTTransformers to the parser.
func WithASTTransformers(ps ...util.PrioritizedValue) Option {
	return &withASTTransformers{ps}
}

type withOption struct {
	name  OptionName
	value interface{}
}

func (o *withOption) SetConfig(c *Config) {
	c.Options[o.name] = o.value
}

// WithOption is a functional option that allow you to set
// an arbitary option to the parser.
func WithOption(name OptionName, value interface{}) Option {
	return &withOption{name, value}
}

// NewParser returns a new Parser with given options.
func NewParser(options ...Option) Parser {
	config := NewConfig()
	for _, opt := range options {
		opt.SetConfig(config)
	}

	p := &parser{
		options: map[OptionName]interface{}{},
		config:  config,
	}

	return p
}

func (p *parser) AddOption(o Option) {
	o.SetConfig(p.config)
}

func (p *parser) addBlockParser(v util.PrioritizedValue, options map[OptionName]interface{}) {
	bp, ok := v.Value.(BlockParser)
	if !ok {
		panic(fmt.Sprintf("%v is not a BlockParser", v.Value))
	}
	so, ok := v.Value.(SetOptioner)
	if ok {
		for oname, ovalue := range options {
			so.SetOption(oname, ovalue)
		}
	}
	p.blockParsers = append(p.blockParsers, bp)
}

func (p *parser) addInlineParser(v util.PrioritizedValue, options map[OptionName]interface{}) {
	ip, ok := v.Value.(InlineParser)
	if !ok {
		panic(fmt.Sprintf("%v is not a InlineParser", v.Value))
	}
	tcs := ip.Trigger()
	so, ok := v.Value.(SetOptioner)
	if ok {
		for oname, ovalue := range options {
			so.SetOption(oname, ovalue)
		}
	}
	p.inlineParsersList = append(p.inlineParsersList, ip)
	for _, tc := range tcs {
		if p.inlineParsers[tc] == nil {
			p.inlineParsers[tc] = []InlineParser{}
		}
		p.inlineParsers[tc] = append(p.inlineParsers[tc], ip)
	}
}

func (p *parser) addParagraphTransformer(v util.PrioritizedValue, options map[OptionName]interface{}) {
	pt, ok := v.Value.(ParagraphTransformer)
	if !ok {
		panic(fmt.Sprintf("%v is not a ParagraphTransformer", v.Value))
	}
	so, ok := v.Value.(SetOptioner)
	if ok {
		for oname, ovalue := range options {
			so.SetOption(oname, ovalue)
		}
	}
	p.paragraphTransformers = append(p.paragraphTransformers, pt)
}

func (p *parser) addASTTransformer(v util.PrioritizedValue, options map[OptionName]interface{}) {
	at, ok := v.Value.(ASTTransformer)
	if !ok {
		panic(fmt.Sprintf("%v is not a ASTTransformer", v.Value))
	}
	so, ok := v.Value.(SetOptioner)
	if ok {
		for oname, ovalue := range options {
			so.SetOption(oname, ovalue)
		}
	}
	p.astTransformers = append(p.astTransformers, at)
}

// A ParseConfig struct is a data structure that holds configuration of the Parser.Parse.
type ParseConfig struct {
	Context Context
}

// A ParseOption is a functional option type for the Parser.Parse.
type ParseOption func(c *ParseConfig)

// WithContext is a functional option that allow you to override
// a default context.
func WithContext(context Context) ParseOption {
	return func(c *ParseConfig) {
		c.Context = context
	}
}

func (p *parser) Parse(reader text.Reader, opts ...ParseOption) ast.Node {
	p.initSync.Do(func() {
		p.config.BlockParsers.Sort()
		for _, v := range p.config.BlockParsers {
			p.addBlockParser(v, p.config.Options)
		}
		p.config.InlineParsers.Sort()
		for _, v := range p.config.InlineParsers {
			p.addInlineParser(v, p.config.Options)
		}
		p.config.ParagraphTransformers.Sort()
		for _, v := range p.config.ParagraphTransformers {
			p.addParagraphTransformer(v, p.config.Options)
		}
		p.config.ASTTransformers.Sort()
		for _, v := range p.config.ASTTransformers {
			p.addASTTransformer(v, p.config.Options)
		}
		p.config = nil
	})
	c := &ParseConfig{}
	for _, opt := range opts {
		opt(c)
	}
	if c.Context == nil {
		c.Context = NewContext(reader.Source())
	}
	pc := c.Context
	root := ast.NewDocument()
	p.parseBlocks(root, reader, pc)
	blockReader := text.NewBlockReader(reader.Source(), nil)
	p.walkBlock(root, func(node ast.Node) {
		p.parseBlock(blockReader, node, pc)
	})
	for _, at := range p.astTransformers {
		at.Transform(root, pc)
	}
	//root.Dump(reader.Source(), 0)
	return root
}

func (p *parser) transformParagraph(node *ast.Paragraph, pc Context) {
	for _, pt := range p.paragraphTransformers {
		pt.Transform(node, pc)
		if node.Parent() == nil {
			break
		}
	}
}

func (p *parser) closeBlocks(from, to int, pc Context) {
	blocks := pc.OpenedBlocks()
	last := pc.LastOpenedBlock()
	for i := from; i >= to; i-- {
		node := blocks[i].Node
		if node.Parent() != nil {
			blocks[i].Parser.Close(blocks[i].Node, pc)
			paragraph, ok := node.(*ast.Paragraph)
			if ok && node.Parent() != nil {
				p.transformParagraph(paragraph, pc)
			}
		}
	}
	if from == len(blocks)-1 {
		blocks = blocks[0:to]
	} else {
		blocks = append(blocks[0:to], blocks[from+1:]...)
	}
	l := len(blocks)
	if l == 0 {
		last.Node = nil
	} else {
		last = blocks[l-1]
	}
	pc.SetOpenedBlocks(blocks)
	pc.SetLastOpenedBlock(last)
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
	shouldPeek := true
	var currentLineNum int
	var w int
	var pos int
	var line []byte
	for _, bp := range p.blockParsers {
		if shouldPeek {
			currentLineNum, _ = reader.Position()
			line, _ = reader.PeekLine()
			w, pos = util.IndentWidth(line, 0)
			pc.SetBlockOffset(pos)
			shouldPeek = false
			if line == nil || line[0] == '\n' {
				break
			}
		}
		if continuable && result == noBlocksOpened && !bp.CanInterruptParagraph() {
			continue
		}
		if w > 3 && !bp.CanAcceptIndentedLine() {
			continue
		}
		last := pc.LastOpenedBlock().Node
		node, state := bp.Open(parent, reader, pc)
		if l, _ := reader.Position(); l != currentLineNum {
			panic("BlockParser.Open must not advance position beyond the current line")
		}
		if node != nil {
			shouldPeek = true
			node.SetBlankPreviousLines(blankLine)
			if last != nil && last.Parent() == nil {
				lastPos := len(pc.OpenedBlocks()) - 1
				p.closeBlocks(lastPos, lastPos, pc)
			}
			parent.AppendChild(parent, node)
			result = newBlocksOpened
			be := Block{node, bp}
			pc.SetOpenedBlocks(append(pc.OpenedBlocks(), be))
			pc.SetLastOpenedBlock(be)
			if state == HasChildren {
				parent = node
				goto retry // try child block
			}
			break // no children, can not open more blocks on this line
		}
	}
	if result == noBlocksOpened && continuable {
		state := lastBlock.Parser.Continue(lastBlock.Node, reader, pc)
		if state&Continue != 0 {
			result = paragraphContinuation
		}
	}
	return result
}

type lineStat struct {
	lineNum int
	level   int
	isBlank bool
}

func isBlankLine(lineNum, level int, stats []lineStat) ([]lineStat, bool) {
	ret := false
	for i := len(stats) - 1 - level; i >= 0; i-- {
		s := stats[i]
		if s.lineNum == lineNum && s.level == level {
			ret = s.isBlank
			continue
		}
		if s.lineNum < lineNum {
			return stats[i:], ret
		}
	}
	return stats[0:0], ret
}

func (p *parser) parseBlocks(parent ast.Node, reader text.Reader, pc Context) {
	pc.SetLastOpenedBlock(Block{})
	pc.SetOpenedBlocks([]Block{})
	blankLines := make([]lineStat, 0, 64)
	isBlank := false
	for { // process blocks separated by blank lines
		_, lines, ok := reader.SkipBlankLines()
		if !ok {
			return
		}
		// first, we try to open blocks
		if p.openBlocks(parent, lines != 0, reader, pc) != newBlocksOpened {
			return
		}
		lineNum, _ := reader.Position()
		for i := 0; i < len(pc.OpenedBlocks()); i++ {
			blankLines = append(blankLines, lineStat{lineNum - 1, i, lines != 0})
		}
		reader.AdvanceLine()
		for len(pc.OpenedBlocks()) != 0 { // process opened blocks line by line
			lastIndex := len(pc.OpenedBlocks()) - 1
			for i := 0; i < len(pc.OpenedBlocks()); i++ {
				be := pc.OpenedBlocks()[i]
				line, _ := reader.PeekLine()
				if line == nil {
					p.closeBlocks(lastIndex, 0, pc)
					reader.AdvanceLine()
					return
				}
				lineNum, _ := reader.Position()
				blankLines = append(blankLines, lineStat{lineNum, i, util.IsBlank(line)})
				// If node is a paragraph, p.openBlocks determines whether it is continuable.
				// So we do not process paragraphs here.
				if !ast.IsParagraph(be.Node) {
					state := be.Parser.Continue(be.Node, reader, pc)
					if state&Continue != 0 {
						// When current node is a container block and has no children,
						// we try to open new child nodes
						if state&HasChildren != 0 && i == lastIndex {
							blankLines, isBlank = isBlankLine(lineNum-1, i, blankLines)
							p.openBlocks(be.Node, isBlank, reader, pc)
							break
						}
						continue
					}
				}
				// current node may be closed or lazy continuation
				blankLines, isBlank = isBlankLine(lineNum-1, i, blankLines)
				thisParent := parent
				if i != 0 {
					thisParent = pc.OpenedBlocks()[i-1].Node
				}
				result := p.openBlocks(thisParent, isBlank, reader, pc)
				if result != paragraphContinuation {
					p.closeBlocks(lastIndex, i, pc)
				}
				break
			}

			reader.AdvanceLine()
		}
	}
}

func (p *parser) walkBlock(block ast.Node, cb func(node ast.Node)) {
	for c := block.FirstChild(); c != nil; c = c.NextSibling() {
		p.walkBlock(c, cb)
	}
	cb(block)
}

func (p *parser) parseBlock(block text.BlockReader, parent ast.Node, pc Context) {
	if parent.IsRaw() {
		return
	}
	escaped := false
	source := block.Source()
	block.Reset(parent.Lines())
	for {
	retry:
		line, _ := block.PeekLine()
		if line == nil {
			break
		}
		lineLength := len(line)
		l, startPosition := block.Position()
		n := 0
		softLinebreak := false
		for i := 0; i < lineLength; i++ {
			c := line[i]
			if c == '\n' {
				softLinebreak = true
				break
			}
			isSpace := util.IsSpace(c)
			isPunct := util.IsPunct(c)
			if (isPunct && !escaped) || isSpace || i == 0 {
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
						ast.MergeOrAppendTextSegment(parent, startPosition.Between(currentPosition))
						_, startPosition = block.Position()
					}
					var inlineNode ast.Node
					for _, ip := range ips {
						inlineNode = ip.Parse(parent, block, pc)
						if inlineNode != nil {
							break
						}
						block.SetPosition(savedLine, savedPosition)
					}
					if inlineNode != nil {
						parent.AppendChild(parent, inlineNode)
						goto retry
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
		stop := diff.Stop
		hardlineBreak := false
		if lineLength > 2 && line[lineLength-2] == '\\' && softLinebreak { // ends with \\n
			stop--
			hardlineBreak = true
		} else if lineLength > 3 && line[lineLength-3] == ' ' && line[lineLength-2] == ' ' && softLinebreak { // ends with [space][space]\n
			hardlineBreak = true
		}
		rest := diff.WithStop(stop)
		text := ast.NewTextSegment(rest.TrimRightSpace(source))
		text.SetSoftLineBreak(softLinebreak)
		text.SetHardLineBreak(hardlineBreak)
		parent.AppendChild(parent, text)
		block.AdvanceLine()
	}

	ProcessDelimiters(nil, pc)
	for _, ip := range p.inlineParsersList {
		ip.CloseBlock(parent, pc)
	}

}
