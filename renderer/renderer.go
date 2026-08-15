// Package renderer renders the given AST to certain formats.
package renderer

import (
	"maps"
	"reflect"
	"sync"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/util"
)

// ContextKey is a key that is used to set arbitrary values to the rendering context.
type ContextKey int

// contextKeyMax is a maximum value of the ContextKey.
var contextKeyMax ContextKey

// NewContextKey returns a new ContextKey value.
func NewContextKey() ContextKey {
	contextKeyMax++
	return contextKeyMax
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
	if contextKeyMax > 0 {
		c.store = make([]any, contextKeyMax+1)
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
	Render(w W, source []byte, n ast.Node, opts ...RenderOption) error
	RenderStringSource(w W, source string, n ast.Node, opts ...RenderOption) error
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

// A NodeRendererDecorator is a function type used for decorating a NodeRenderer.
type NodeRendererDecorator[W any] = func(next NodeRenderer[W]) NodeRenderer[W]

// A Config struct holds configuration for Renderer.
type Config[W any, C any] struct {
	nodeRenderers          map[ast.NodeKind]NodeRenderer[W]
	nodeRendererDecorators map[ast.NodeKind]NodeRendererDecorator[W]
	extensions             []Extension[C]
}

// Option is a functional option for configuring a Renderer, typically
// passed to a format-specific constructor (e.g. html.New).
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

// WithNodeRendererDecorators sets the node renderer decorators for the Renderer.
//
// If a decorator is already set for a node kind, the new decorator will be applied to the existing one.
func WithNodeRendererDecorators[W any, C any](
	nodeRendererDecorators map[ast.NodeKind]NodeRendererDecorator[W]) Option[C] {
	return NewOptionFunc(func(c *C) {
		cfg := getConfig[W, C](c)
		if cfg != nil {
			if cfg.nodeRendererDecorators == nil {
				cfg.nodeRendererDecorators = make(map[ast.NodeKind]NodeRendererDecorator[W])
			}
			for kind, decorator := range nodeRendererDecorators {
				existing, ok := cfg.nodeRendererDecorators[kind]
				if ok {
					newDecorator := func(next NodeRenderer[W]) NodeRenderer[W] {
						return decorator(existing(next))
					}
					cfg.nodeRendererDecorators[kind] = newDecorator
				} else {
					cfg.nodeRendererDecorators[kind] = decorator
				}
			}
		}
	})
}

// WithNodeRendererDecorator sets a node renderer decorator for the given node kind.
//
// If a decorator is already set for the node kind, the new decorator will be applied to the existing one.
func WithNodeRendererDecorator[W any, C any](kind ast.NodeKind, decorator NodeRendererDecorator[W]) Option[C] {
	return NewOptionFunc(func(c *C) {
		cfg := getConfig[W, C](c)
		if cfg != nil {
			if cfg.nodeRendererDecorators == nil {
				cfg.nodeRendererDecorators = make(map[ast.NodeKind]NodeRendererDecorator[W])
			}
			existing, ok := cfg.nodeRendererDecorators[kind]
			if ok {
				newDecorator := func(next NodeRenderer[W]) NodeRenderer[W] {
					return decorator(existing(next))
				}
				cfg.nodeRendererDecorators[kind] = newDecorator
			} else {
				cfg.nodeRendererDecorators[kind] = decorator
			}
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

// Helper is a helper struct for implementing Renderer.
type Helper[W any, C any] struct {
	config        C
	options       []Option[C]
	nodeRenderers []NodeRenderer[W]
	initSync      sync.Once
}

// NewHelper returns a new Helper configured with the given Options.
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

type renderConfig struct {
	context Context
}

// RenderOption is an interface that represents an option for rendering.
type RenderOption interface {
	SetRenderOption(*renderConfig)
}

type withContext struct {
	context Context
}

func (o *withContext) SetRenderOption(c *renderConfig) {
	c.context = o.context
}

// WithContext sets the context for rendering.
func WithContext(context Context) RenderOption {
	return &withContext{context: context}
}

// Render renders the given AST node to the given writer with the given Renderer.
func (r *Helper[W, C]) Render(w W, source []byte, n ast.Node, opts ...RenderOption) error {
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
		r.nodeRenderers = make([]NodeRenderer[W], ast.CurrentKindValue+1)
		for kind, nr := range cfg.nodeRenderers {
			decorator, ok := cfg.nodeRendererDecorators[kind]
			if ok {
				r.nodeRenderers[kind] = decorator(nr)
			} else {
				r.nodeRenderers[kind] = nr
			}
		}
	})

	var rc Context
	if len(opts) != 0 {
		var rcfg renderConfig
		for _, opt := range opts {
			opt.SetRenderOption(&rcfg)
		}
		if rcfg.context != nil {
			if c, ok := rcfg.context.(*renderContext); ok {
				if c.renderFn == nil {
					c.renderFn = r.renderFn
				}
			}
			rc = rcfg.context
		}
	}
	if rc == nil {
		rc = NewContext(WithRenderFunc(r.renderFn))
	}
	err := r.renderFn(w, source, n, rc)
	if err != nil {
		return err
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
func (r *Helper[W, C]) RenderStringSource(w W, source string, n ast.Node, opts ...RenderOption) error {
	return r.Render(w, util.StringToReadOnlyBytes(source), n, opts...)
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
