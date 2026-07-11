// Package renderer renders the given AST to certain formats.
package renderer

import (
	"maps"
	"reflect"
	"slices"
	"sync"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/util"
)

// ContextKey is a key that is used to set arbitrary values to the rendering context.
type ContextKey int

// ContextKeyMax is a maximum value of the ContextKey.
var ContextKeyMax ContextKey

// NewContextKey returns a new ContextKey value.
func NewContextKey() ContextKey {
	ContextKeyMax++
	return ContextKeyMax
}

// A Context interface holds information that is necessary to render Markdown text.
type Context interface {
	// Get returns a value associated with the given key.
	Get(ContextKey) any

	// ComputeIfAbsent computes a value if a value associated with the given key is absent and returns the value.
	ComputeIfAbsent(ContextKey, func() any) any

	// Set sets the given value to the context.
	Set(ContextKey, any)

	// Render renders the given node using the renderer associated with this context.
	// If no rendering function has been set, it is a no-op and returns nil.
	Render(w any, source []byte, n ast.Node) error
}

// ContextOption is a functional option for NewContext.
type ContextOption func(*renderContext)

// WithRenderFunc sets the rendering function used by Context.Render.
func WithRenderFunc(f func(any, []byte, ast.Node, Context) error) ContextOption {
	return func(c *renderContext) {
		c.renderFn = f
	}
}

type renderContext struct {
	store    []any
	renderFn func(any, []byte, ast.Node, Context) error
}

// NewContext returns a new rendering Context.
func NewContext(opts ...ContextOption) Context {
	c := &renderContext{
		renderFn: func(_ any, _ []byte, _ ast.Node, _ Context) error { return nil },
	}
	if ContextKeyMax > 0 {
		c.store = make([]any, ContextKeyMax+1)
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *renderContext) Get(key ContextKey) any {
	if int(key) >= len(c.store) {
		return nil
	}
	return c.store[key]
}

func (c *renderContext) ComputeIfAbsent(key ContextKey, f func() any) any {
	if int(key) >= len(c.store) {
		return nil
	}
	v := c.store[key]
	if v == nil {
		v = f()
		c.store[key] = v
	}
	return v
}

func (c *renderContext) Set(key ContextKey, value any) {
	if int(key) >= len(c.store) {
		panic("context key is out of range")
	}
	c.store[key] = value
}

func (c *renderContext) Render(w any, source []byte, n ast.Node) error {
	return c.renderFn(w, source, n, c)
}

// A Renderer interface is used for rendering a given AST node to a certain format.
type Renderer[W any] interface {
	Render(w W, source []byte, n ast.Node) error
	RenderStringSource(w W, source string, n ast.Node) error
}

// A NodeRenderer interface is used for rendering a given node.
type NodeRenderer[W any] interface {
	Render(w W, source []byte, n ast.Node, entering bool, rc Context) (ast.WalkStatus, error)
}

// NodeRendererFunc is a function that implements NodeRenderer interface.
func NodeRendererFunc[W any](f func(w W, source []byte,
	n ast.Node, entering bool, rc Context) (ast.WalkStatus, error)) NodeRenderer[W] {
	return &nodeRendererFunc[W]{f: f}
}

type nodeRendererFunc[W any] struct {
	f func(w W, source []byte, n ast.Node, entering bool, rc Context) (ast.WalkStatus, error)
}

func (f *nodeRendererFunc[W]) Render(w W, source []byte,
	n ast.Node, entering bool, rc Context) (ast.WalkStatus, error) {
	return f.f(w, source, n, entering, rc)
}

// A Hook interface is used for hooking into the rendering process.
type Hook[W any] interface {
	PreRender(w W, source []byte, n ast.Node, rc Context) error

	PostRender(w W, source []byte, n ast.Node, rc Context) error
}

// EmptyHook is a Hook that does nothing.
type EmptyHook[W any] struct{}

// PreRender implements Hook.PreRender.
func (h *EmptyHook[W]) PreRender(_ W, _ []byte, _ ast.Node, _ Context) error {
	return nil
}

// PostRender implements Hook.PostRender.
func (h *EmptyHook[W]) PostRender(_ W, _ []byte, _ ast.Node, _ Context) error {
	return nil
}

// A Config struct holds configuration for Renderer.
type Config[W any, C any] struct {
	nodeRenderers map[ast.NodeKind]NodeRenderer[W]
	extensions    []Extension[C]
	hooks         []Hook[W]
}

// Option is a functional option for NewRenderer.
type Option[C any] interface {
	SetFormatOption(*C)
}

type optionFunc[C any] struct {
	f func(*C)
}

func (o *optionFunc[C]) SetFormatOption(c *C) {
	o.f(c)
}

// NewOptionFunc returns a new Option that applies the given function to the configuration.
func NewOptionFunc[C any](f func(*C)) Option[C] {
	return &optionFunc[C]{f: f}
}

// WithNodeRenderers sets the node renderers for the Renderer.
func WithNodeRenderers[W any, C any](nodeRenderers map[ast.NodeKind]NodeRenderer[W]) Option[C] {
	return NewOptionFunc(func(c *C) {
		cfg := getConfig[W, C](c)
		if cfg != nil {
			if cfg.nodeRenderers == nil {
				cfg.nodeRenderers = make(map[ast.NodeKind]NodeRenderer[W])
			}
			maps.Copy(cfg.nodeRenderers, nodeRenderers)
		}
	})
}

// WithNodeRenderer sets a node renderer for the given node kind.
func WithNodeRenderer[W any, C any](kind ast.NodeKind, nodeRenderer NodeRenderer[W]) Option[C] {
	return NewOptionFunc(func(c *C) {
		cfg := getConfig[W, C](c)
		if cfg != nil {
			if cfg.nodeRenderers == nil {
				cfg.nodeRenderers = make(map[ast.NodeKind]NodeRenderer[W])
			}
			cfg.nodeRenderers[kind] = nodeRenderer
		}
	})
}

// WithExtensions sets the extensions for the Renderer.
func WithExtensions[W any, C any](ext ...Extension[C]) Option[C] {
	return NewOptionFunc(func(c *C) {
		cfg := getConfig[W, C](c)
		if cfg != nil {
			cfg.extensions = append(cfg.extensions, ext...)
		}
	})
}

// WithHooks sets the hooks for the Renderer.
func WithHooks[W any, C any](hooks ...Hook[W]) Option[C] {
	return NewOptionFunc(func(c *C) {
		cfg := getConfig[W, C](c)
		if cfg != nil {
			cfg.hooks = append(cfg.hooks, hooks...)
		}
	})
}

// Helper is a helper struct for implementing Renderer.
type Helper[W any, C any] struct {
	config           C
	options          []Option[C]
	nodeRenderersMap map[ast.NodeKind]NodeRenderer[W]
	nodeRenderers    []NodeRenderer[W]
	initSync         sync.Once
	hooks            []Hook[W]
}

// NewHelper returns a new RendererHelper with the given RendererSpec.
func NewHelper[W any, C any](opts ...Option[C]) *Helper[W, C] {
	var c C
	if df, ok := any(c).(interface {
		Default() C
	}); ok {
		c = df.Default()
	}

	h := &Helper[W, C]{
		options: opts,
		config:  c,
	}
	return h
}

// Config returns the configuration of this Helper.
func (r *Helper[W, C]) Config() *C {
	return &r.config
}

// Register registers the given NodeRenderer for the given node kind.
func (r *Helper[W, C]) Register(kind ast.NodeKind, n NodeRenderer[W]) {
	if r.nodeRenderersMap == nil {
		r.nodeRenderersMap = make(map[ast.NodeKind]NodeRenderer[W])
	}
	r.nodeRenderersMap[kind] = n
}

func (r *Helper[W, C]) renderFn(a any, source []byte, n ast.Node, rc Context) error {
	w := a.(W)
	return ast.Walk(n, func(n ast.Node, entering bool) (s ast.WalkStatus, err error) {
		s = ast.WalkStatus(ast.WalkContinue)
		f := r.nodeRenderers[n.Kind()]
		if f != nil {
			s, err = f.Render(w, source, n, entering, rc)
		}
		return
	})
}

// Render renders the given AST node to the given writer with the given Renderer.
func (r *Helper[W, C]) Render(w W, source []byte, n ast.Node) error {
	r.initSync.Do(func() {
		for _, opt := range r.options {
			opt.SetFormatOption(&r.config)
		}
		cfg := getConfig[W, C](&r.config)
		for _, ext := range cfg.extensions {
			for _, opt := range ext.RendererOptions(&r.config) {
				opt.SetFormatOption(&r.config)
			}
		}
		for kind, nr := range cfg.nodeRenderers {
			r.Register(kind, nr)
		}
		r.nodeRenderers = make([]NodeRenderer[W], ast.CurrentKindValue+1)
		for kind, nr := range cfg.nodeRenderers {
			r.nodeRenderers[kind] = nr
		}
		r.hooks = cfg.hooks
	})
	rc := NewContext(WithRenderFunc(r.renderFn))
	for _, hook := range r.hooks {
		if err := hook.PreRender(w, source, n, rc); err != nil {
			return err
		}
	}
	err := r.renderFn(w, source, n, rc)
	if err != nil {
		return err
	}
	for _, hook := range slices.Backward(r.hooks) {
		if err := hook.PostRender(w, source, n, rc); err != nil {
			return err
		}
	}
	a := any(w)
	if fw, ok := a.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			return err
		}
	}
	if errr, ok := a.(interface{ Error() error }); ok {
		if err := errr.Error(); err != nil {
			return err
		}
	}
	return nil
}

// RenderStringSource renders the given AST node to the given writer with the given Renderer and string source.
func (r *Helper[W, C]) RenderStringSource(w W, source string, n ast.Node) error {
	return r.Render(w, util.StringToReadOnlyBytes(source), n)
}

// Extension is an interface that represents an extension for Renderer.
type Extension[C any] interface {
	// RendererOptions returns the options for the Renderer.
	RendererOptions(*C) []Option[C]
}

func getConfig[W any, C any](c any) *Config[W, C] {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		field := v.Elem().FieldByName("Config")
		if field.IsValid() && field.CanAddr() {
			return field.Addr().Interface().(*Config[W, C])
		}
	}
	return nil
}
