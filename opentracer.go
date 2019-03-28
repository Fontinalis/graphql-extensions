package extensions

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	opentracing "github.com/opentracing/opentracing-go"
)

// OpenTracer is an extension for graphql-go/graphql that adds support for opentracing
type OpenTracer struct {
	tracer    opentracing.Tracer
	rootSpan  opentracing.Span
	rootLabel string
}

// NewOpenTracer returns a new OpenTracer graphql-go/graphql extension
func NewOpenTracer(tracer opentracing.Tracer, rootLabel string) graphql.Extension {
	return &OpenTracer{
		tracer:    tracer,
		rootLabel: rootLabel,
	}
}

// Init implements the graphql-go/graphql.Extension.Init function
func (t *OpenTracer) Init(ctx context.Context, p *graphql.Params) context.Context {
	t.rootSpan = t.tracer.StartSpan(t.rootLabel)
	return ctx
}

// Name is "opentracing"
func (t *OpenTracer) Name() string {
	return "opentracing"
}

// HasResult is false since this extension does not return any
func (t *OpenTracer) HasResult() bool {
	return false
}

// GetResult returns nil
func (t *OpenTracer) GetResult(ctx context.Context) interface{} {
	return nil
}

// ParseDidStart implements the graphql-go/graphql.Extension.ParseDidStart function
func (t *OpenTracer) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	label := fmt.Sprint("parse")
	span := t.tracer.StartSpan(label, opentracing.ChildOf(t.rootSpan.Context()))
	return ctx, func(err error) {
		if err != nil {
			span.SetTag("error", err)
		}
		span.Finish()
	}
}

// ValidationDidStart implements the graphql-go/graphql.Extension.ValidationDidStart function
func (t *OpenTracer) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	label := fmt.Sprint("validation")
	span := t.tracer.StartSpan(label, opentracing.ChildOf(t.rootSpan.Context()))
	return ctx, func(errs []gqlerrors.FormattedError) {
		span.SetTag("errors", errs)
		span.Finish()
	}
}

// ExecutionDidStart implements the graphql-go/graphql.Extension.ExecutionDidStart function
func (t *OpenTracer) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	label := fmt.Sprint("execution")
	span := t.tracer.StartSpan(label, opentracing.ChildOf(t.rootSpan.Context()))
	ctx = context.WithValue(ctx, getCtxKey(nil), span.Context())
	return ctx, func(r *graphql.Result) {
		span.Finish()
		t.rootSpan.Finish()
	}
}

// ResolveFieldDidStart implements the graphql-go/graphql.Extension.ResolveFieldDidStart function
func (t *OpenTracer) ResolveFieldDidStart(ctx context.Context, i *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	label := fmt.Sprint(i.Path.AsArray())
	parentCtx, _ := ctx.Value(getParentCtxKey(i.Path.AsArray())).(opentracing.SpanContext)
	span := opentracing.StartSpan(label, opentracing.ChildOf(parentCtx))
	ctx = context.WithValue(ctx, getCtxKey(i.Path.AsArray()), span.Context())
	return ctx, func(v interface{}, err error) {
		if err != nil {
			span.SetTag("error", err)
		}
		span.SetTag("fieldName", i.FieldName)
		span.SetTag("parentType", i.ParentType.String())
		span.SetTag("returnType", i.ReturnType.String())
		span.SetTag("value", v)
		span.Finish()
	}
}

type ctxKey string

func getCtxKey(args []interface{}) ctxKey {
	key := "opentracing_root"
	if args == nil {
		return ctxKey(key)
	}
	for _, arg := range args {
		key += fmt.Sprintf("_%v", arg)
	}
	return ctxKey(key)
}

func getParentCtxKey(args []interface{}) ctxKey {
	key := "opentracing_root"
	if len(args) <= 1 {
		return ctxKey(key)
	}
	if _, ok := args[len(args)-2].(int); ok {
		for i := 0; i < len(args)-2; i++ {
			key += fmt.Sprintf("_%v", args[i])
		}
		return ctxKey(key)
	}
	for i := 0; i < len(args)-1; i++ {
		key += fmt.Sprintf("_%v", args[i])
	}
	return ctxKey(key)
}
