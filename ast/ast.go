// Package ast defines AST nodes that represent markdown elements.
package ast

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"strings"

	textm "github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// NodeKind indicates more specific type than NodeType.
type NodeKind int

func (k NodeKind) String() string {
	return kindNames[k]
}

// CurrentKindValue is the current value of NodeKind.
//
// You **MUST NOT** use this value.
// This value is public for the goldmark internal package to use.
var CurrentKindValue NodeKind

var kindNames = []string{""}

// NewNodeKind returns a new Kind value.
func NewNodeKind(name string) NodeKind {
	CurrentKindValue++
	kindNames = append(kindNames, name)
	return CurrentKindValue
}

// An Attribute is an attribute of the Node.
type Attribute struct {
	Name  string
	Value textm.MultilineValue
}

// A Node interface defines basic AST node functionalities.
type Node interface {
	// Kind returns a kind of this node.
	Kind() NodeKind

	// Pos returns a position of this node in a source.
	// If this node position is not defined, Pos returns -1.
	Pos() int

	// SetPos sets a position of this node in a source.
	// Some node may ignore this method. For example, Paragraph node ignores this method because
	// it calculates its position from its lines.
	SetPos(v int)

	// NextSibling returns a next sibling node of this node.
	NextSibling() Node

	// PreviousSibling returns a previous sibling node of this node.
	PreviousSibling() Node

	// Parent returns a parent node of this node.
	Parent() Node

	// SetParent sets a parent node to this node.
	SetParent(Node)

	// SetPreviousSibling sets a previous sibling node to this node.
	SetPreviousSibling(Node)

	// SetNextSibling sets a next sibling node to this node.
	SetNextSibling(Node)

	// HasChildren returns true if this node has any children, otherwise false.
	HasChildren() bool

	// ChildCount returns a total number of children.
	ChildCount() int

	// Children returns an iterator of children of this node.
	Children() iter.Seq[Node]

	// FirstChild returns a first child of this node.
	FirstChild() Node

	// LastChild returns a last child of this node.
	LastChild() Node

	// AppendChild append a node child to the tail of the children.
	AppendChild(child Node)

	// RemoveChild removes a node child from this node.
	// If a node child is not children of this node, RemoveChild nothing to do.
	RemoveChild(child Node)

	// RemoveChildren removes all children from this node.
	RemoveChildren()

	// ReplaceChild replace a node v1 with a node insertee.
	// If v1 is not children of this node, ReplaceChild append a insetee to the
	// tail of the children.
	ReplaceChild(v1, insertee Node)

	// InsertBefore inserts a node insertee before a node v1.
	// If v1 is not children of this node, InsertBefore append a insetee to the
	// tail of the children.
	InsertBefore(v1, insertee Node)

	// InsertAfter inserts a node insertee after a node v1.
	// If v1 is not children of this node, InsertBefore append a insetee to the
	// tail of the children.
	InsertAfter(v1, insertee Node)

	// OwnerDocument returns this node's owner document.
	// If this node is not a child of the Document node, OwnerDocument
	// returns nil.
	OwnerDocument() *Document

	// Dump dumps an AST node.
	Dump(source []byte) *NodeDump

	// SetAttribute sets the given value to the attributes.
	SetAttribute(name string, value textm.MultilineValue)

	// Attribute returns a (attribute value, true) if an attribute
	// associated with the given name is found, otherwise
	// (zero MultilineValue, false)
	Attribute(name string) (textm.MultilineValue, bool)

	// Attributes returns a list of attributes.
	// This may be a nil if there are no attributes.
	Attributes() []Attribute

	// RemoveAttributes removes all attributes from this node.
	RemoveAttributes()
}

// A BlockNode interface is a Node that represents a block element.
// Block nodes contain source text segments that will be parsed into inline nodes.
type BlockNode interface {
	Node
	blockNode()

	// HasBlankPreviousLines returns true if the row before this node is blank,
	// otherwise false.
	HasBlankPreviousLines() bool

	// SetBlankPreviousLines sets whether the row before this node is blank.
	SetBlankPreviousLines(v bool)

	// Source returns text segments that hold positions in a source.
	// Source holds raw source text that will be parsed into inline nodes.
	Source() []textm.Segment

	// SetSource sets text segments that hold positions in a source.
	SetSource([]textm.Segment)

	// AppendSource appends a text segment to the source.
	AppendSource(textm.Segment)
}

// An InlineNode interface is a Node that represents an inline element.
type InlineNode interface {
	Node
	inlineNode()
}

// A BaseNode struct implements the Node interface partially.
type BaseNode struct {
	self       Node
	firstChild Node
	lastChild  Node
	parent     Node
	next       Node
	prev       Node
	childCount int
	attributes []Attribute
	pos        int
}

func ensureIsolated(v Node) {
	if p := v.Parent(); p != nil {
		p.RemoveChild(v)
	}
}

// Init initializes this node. Init must be called in each node's constructor.
func (n *BaseNode) Init(self Node) {
	n.self = self
	n.pos = -1
}

// Pos implements Node.Pos .
func (n *BaseNode) Pos() int {
	return n.pos
}

// SetPos implements Node.SetPos .
func (n *BaseNode) SetPos(v int) {
	n.pos = v
}

// HasChildren implements Node.HasChildren .
func (n *BaseNode) HasChildren() bool {
	return n.firstChild != nil
}

// SetPreviousSibling implements Node.SetPreviousSibling .
func (n *BaseNode) SetPreviousSibling(v Node) {
	n.prev = v
}

// SetNextSibling implements Node.SetNextSibling .
func (n *BaseNode) SetNextSibling(v Node) {
	n.next = v
}

// PreviousSibling implements Node.PreviousSibling .
func (n *BaseNode) PreviousSibling() Node {
	return n.prev
}

// NextSibling implements Node.NextSibling .
func (n *BaseNode) NextSibling() Node {
	return n.next
}

// RemoveChild implements Node.RemoveChild .
func (n *BaseNode) RemoveChild(v Node) {
	if v.Parent() != n.self {
		return
	}
	n.childCount--
	prev := v.PreviousSibling()
	next := v.NextSibling()
	if prev != nil {
		prev.SetNextSibling(next)
	} else {
		n.firstChild = next
	}
	if next != nil {
		next.SetPreviousSibling(prev)
	} else {
		n.lastChild = prev
	}
	v.SetParent(nil)
	v.SetPreviousSibling(nil)
	v.SetNextSibling(nil)
}

// RemoveChildren implements Node.RemoveChildren .
func (n *BaseNode) RemoveChildren() {
	for c := n.firstChild; c != nil; {
		c.SetParent(nil)
		c.SetPreviousSibling(nil)
		next := c.NextSibling()
		c.SetNextSibling(nil)
		c = next
	}
	n.firstChild = nil
	n.lastChild = nil
	n.childCount = 0
}

// FirstChild implements Node.FirstChild .
func (n *BaseNode) FirstChild() Node {
	return n.firstChild
}

// LastChild implements Node.LastChild .
func (n *BaseNode) LastChild() Node {
	return n.lastChild
}

// ChildCount implements Node.ChildCount .
func (n *BaseNode) ChildCount() int {
	return n.childCount
}

// Children implements Node.Children .
func (n *BaseNode) Children() iter.Seq[Node] {
	return func(yield func(Node) bool) {
		for c := n.firstChild; c != nil; {
			nc := c.NextSibling()
			if !yield(c) {
				return
			}
			c = nc
		}
	}
}

// Parent implements Node.Parent .
func (n *BaseNode) Parent() Node {
	return n.parent
}

// SetParent implements Node.SetParent .
func (n *BaseNode) SetParent(v Node) {
	n.parent = v
}

// AppendChild implements Node.AppendChild .
func (n *BaseNode) AppendChild(v Node) {
	ensureIsolated(v)
	if n.firstChild == nil {
		n.firstChild = v
		v.SetNextSibling(nil)
		v.SetPreviousSibling(nil)
	} else {
		last := n.lastChild
		last.SetNextSibling(v)
		v.SetPreviousSibling(last)
	}
	v.SetParent(n.self)
	n.lastChild = v
	n.childCount++
}

// ReplaceChild implements Node.ReplaceChild .
func (n *BaseNode) ReplaceChild(v1, insertee Node) {
	n.InsertBefore(v1, insertee)
	n.RemoveChild(v1)
}

// InsertAfter implements Node.InsertAfter .
func (n *BaseNode) InsertAfter(v1, insertee Node) {
	n.InsertBefore(v1.NextSibling(), insertee)
}

// InsertBefore implements Node.InsertBefore .
func (n *BaseNode) InsertBefore(v1, insertee Node) {
	n.childCount++
	if v1 == nil {
		n.AppendChild(insertee)
		return
	}
	ensureIsolated(insertee)
	if v1.Parent() == n.self {
		c := v1
		prev := c.PreviousSibling()
		if prev != nil {
			prev.SetNextSibling(insertee)
			insertee.SetPreviousSibling(prev)
		} else {
			n.firstChild = insertee
			insertee.SetPreviousSibling(nil)
		}
		insertee.SetNextSibling(c)
		c.SetPreviousSibling(insertee)
		insertee.SetParent(n.self)
	}
}

// OwnerDocument implements Node.OwnerDocument.
func (n *BaseNode) OwnerDocument() *Document {
	d := n.Parent()
	for {
		p := d.Parent()
		if p == nil {
			if v, ok := d.(*Document); ok {
				return v
			}
			break
		}
		d = p
	}
	return nil
}

// SetAttribute implements Node.SetAttribute.
func (n *BaseNode) SetAttribute(name string, value textm.MultilineValue) {
	if n.attributes == nil {
		n.attributes = make([]Attribute, 0, 10)
	} else {
		for i, a := range n.attributes {
			if a.Name == name {
				n.attributes[i].Value = value
				return
			}
		}
	}
	n.attributes = append(n.attributes, Attribute{
		Name:  name,
		Value: value,
	})
}

// Attribute implements Node.Attribute.
func (n *BaseNode) Attribute(name string) (textm.MultilineValue, bool) {
	if n.attributes == nil {
		return textm.MultilineValue{}, false
	}
	for _, a := range n.attributes {
		if a.Name == name {
			return a.Value, true
		}
	}
	return textm.MultilineValue{}, false
}

// Attributes implements Node.Attributes.
func (n *BaseNode) Attributes() []Attribute {
	return n.attributes
}

// RemoveAttributes implements Node.RemoveAttributes.
func (n *BaseNode) RemoveAttributes() {
	n.attributes = nil
}

// NodeDump is a struct that holds information for dumping a node.
type NodeDump struct {
	// Node is the node to be dumped.
	Node Node

	// Properties is a map of additional properties to be dumped.
	// Property names should be PascalCase.
	Properties map[string]any
}

// NewNodeDump returns a new NodeDump.
func NewNodeDump(node Node, properties map[string]any) *NodeDump {
	return &NodeDump{
		Node:       node,
		Properties: properties,
	}
}

// Children returns an iterator of children of this node dump.
func (d *NodeDump) Children(source []byte) iter.Seq[*NodeDump] {
	return func(yield func(*NodeDump) bool) {
		for c := d.Node.FirstChild(); c != nil; c = c.NextSibling() {
			if !yield(c.Dump(source)) {
				return
			}
		}
	}
}

// PrettyPrint pretty prints this node dump to the given writer.
//
// PrettyPrint format may change in the future. Do not rely on the format.
// This method is intended for debugging purpose only.
// The returned error is from the writer. If the writer never returns an error, the returned error is always nil.
func (d *NodeDump) PrettyPrint(w io.Writer, source []byte, level ...int) error {
	ww, ok := w.(util.ErrorBufWriter)
	if !ok {
		ww = util.NewErrorBufWriter(w)
	}

	l := 0
	if len(level) > 0 {
		l = level[0]
	}
	n := d.Node
	name := n.Kind().String()
	indent := strings.Repeat("    ", l)
	_, _ = fmt.Fprintf(ww, "%s%s {\n", indent, name)
	indent2 := strings.Repeat("    ", l+1)
	p := map[string]any{}
	maps.Copy(p, d.Properties)
	p["Pos"] = n.Pos()
	if b, ok := n.(BlockNode); ok {
		if len(b.Source()) != 0 {
			p["Source"] = textm.NewLines(b.Source()).Str(source)
		}
		p["HasBlankPreviousLines"] = b.HasBlankPreviousLines()
	}
	for k, v := range p {
		_, _ = fmt.Fprintf(ww, "%s%s: %v\n", indent2, k, v)
	}
	attrs := n.Attributes()
	if len(attrs) > 0 {
		var ats []any
		for _, attr := range attrs {
			ats = append(ats, attr.Value.Str(source))
		}
		dumpValue(ww, map[string]any{"Attributes": ats}, l+1)
	}
	if n.HasChildren() {
		_, _ = fmt.Fprintf(ww, "%sChildren: [\n", indent2)
		for cd := range d.Children(source) {
			_ = cd.PrettyPrint(ww, source, l+2)
		}
		_, _ = fmt.Fprintf(ww, "%s]\n", indent2)
	}
	_, _ = fmt.Fprintf(ww, "%s}\n", indent)
	_ = ww.Flush()
	return ww.Error()
}

func dumpValue(ww util.BufWriter, v any, level int) {
	indent := strings.Repeat("    ", level)
	indent2 := strings.Repeat("    ", level+1)
	switch v := v.(type) {
	case map[string]any:
		_, _ = ww.WriteString("{\n")
		for k, v := range v {
			_, _ = fmt.Fprintf(ww, "%s%s: ", indent2, k)
			dumpValue(ww, v, level+1)
		}
		_, _ = fmt.Fprintf(ww, "%s}\n", indent)
	case []any:
		_, _ = ww.WriteString("[\n")
		for _, v := range v {
			dumpValue(ww, v, level+1)
		}
		_, _ = fmt.Fprintf(ww, "%s]\n", indent)
	default:
		_, _ = fmt.Fprintf(ww, "%v\n", v)
	}
}

// WalkStatus represents a current status of the Walk function.
type WalkStatus int

const (
	// WalkStop indicates no more walking needed.
	WalkStop WalkStatus = iota + 1

	// WalkSkipChildren indicates that Walk will not walk on children of the current
	// node.
	WalkSkipChildren

	// WalkContinue indicates that Walk can continue to walk.
	WalkContinue
)

// Walker is a function that will be called when Walk find a
// new node.
// entering is set true before walks children, false after walked children.
// If Walker returns error, Walk function immediately stop walking.
type Walker func(n Node, entering bool) (WalkStatus, error)

// Walk walks a AST tree by the depth first search algorithm.
func Walk(n Node, walker Walker) error {
	_, err := walkHelper(n, walker)
	return err
}

func walkHelper(n Node, walker Walker) (WalkStatus, error) {
	status, err := walker(n, true)
	if err != nil || status == WalkStop {
		return status, err
	}
	if status != WalkSkipChildren {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if st, err := walkHelper(c, walker); err != nil || st == WalkStop {
				return WalkStop, err
			}
		}
	}
	status, err = walker(n, false)
	if err != nil || status == WalkStop {
		return WalkStop, err
	}
	return WalkContinue, nil
}

// N is a helper function to create a new node with children.
// Children must be a Node or a string. If a child is a string,
// N creates a Text node with the string and append it to the children.
func N(node Node, children ...any) Node {
	for _, c := range children {
		switch v := c.(type) {
		case Node:
			node.AppendChild(v)
		case string:
			text := NewStringText(v)
			node.AppendChild(text)
		default:
			panic(fmt.Sprintf("unexpected child type: %T", c))
		}
	}
	return node
}
