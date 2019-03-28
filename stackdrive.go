package extensions

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"go.opencensus.io/trace"
)

// SDTracer is an extension for graphql-go/graphql that adds support for stackdrive tracing
type SDTracer struct {
	rootLabel string
	root      *trace.Span
}

// NewStackdriveTracer returns a new Stackdriver tracer extension
func NewStackdriveTracer(rootLabel string) graphql.Extension {
	return &SDTracer{
		rootLabel: rootLabel,
	}
}

// Init implements the graphql-go/graphql.Extension.Init function
func (t *SDTracer) Init(ctx context.Context, p *graphql.Params) context.Context {
	_, root := trace.StartSpan(ctx, t.rootLabel)
	t.root = root
	return ctx
}

// Name is "stackdriver"
func (t *SDTracer) Name() string {
	return "stackdriver"
}

// HasResult is false since this extension does not return any
func (t *SDTracer) HasResult() bool {
	return false
}

// GetResult returns nil
func (t *SDTracer) GetResult(ctx context.Context) interface{} {
	return nil
}

// ParseDidStart implements the graphql-go/graphql.Extension.ParseDidStart function
func (t *SDTracer) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	label := fmt.Sprint("parse")
	_, span := trace.StartSpanWithRemoteParent(ctx, label, t.root.SpanContext())
	return ctx, func(err error) {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
	}
}

// ValidationDidStart implements the graphql-go/graphql.Extension.ValidationDidStart function
func (t *SDTracer) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	label := fmt.Sprint("validation")
	_, span := trace.StartSpanWithRemoteParent(ctx, label, t.root.SpanContext())
	return ctx, func(errs []gqlerrors.FormattedError) {
		span.AddAttributes(trace.Int64Attribute("num_of_errors", int64(len(errs))))
		span.End()
	}
}

// ExecutionDidStart implements the graphql-go/graphql.Extension.ExecutionDidStart function
func (t *SDTracer) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	label := fmt.Sprint("execution")
	_, span := trace.StartSpanWithRemoteParent(ctx, label, t.root.SpanContext())
	ctx = context.WithValue(ctx, getCtxKey("stackdriver_root", nil), span.SpanContext())
	return ctx, func(r *graphql.Result) {
		span.End()
		t.root.End()
	}
}

// ResolveFieldDidStart implements the graphql-go/graphql.Extension.ResolveFieldDidStart function
func (t *SDTracer) ResolveFieldDidStart(ctx context.Context, i *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	label := fmt.Sprint(i.Path.AsArray())
	parentCtx, _ := ctx.Value(getParentCtxKey("stackdriver_root", i.Path.AsArray())).(trace.SpanContext)
	_, span := trace.StartSpanWithRemoteParent(ctx, label, parentCtx)
	ctx = context.WithValue(ctx, getCtxKey("stackdriver_root", i.Path.AsArray()), span.SpanContext())
	return ctx, func(v interface{}, err error) {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.AddAttributes(
			trace.StringAttribute("fieldName", i.FieldName),
			trace.StringAttribute("parentType", i.ParentType.String()),
			trace.StringAttribute("returnType", i.ReturnType.String()),
			trace.StringAttribute("value", fmt.Sprint(v)),
		)
		span.End()
	}
}
