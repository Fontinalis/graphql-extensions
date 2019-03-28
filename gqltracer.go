package extensions

import (
	"context"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

// NewGQLTracer returns a fresh, still hot GQLTracer
func NewGQLTracer() graphql.Extension {
	return &GQLTracer{
		result: &TracingResult{
			Execution: &ExecutionResult{
				Resolvers: make([]*ResolverResult, 0),
			},
		},
	}
}

// GQLTracer is an extension for graphql-go/graphql that adds support for opentracing
type GQLTracer struct {
	result *TracingResult
}

// Init implements the graphql-go/graphql.Extension.Init function
func (t *GQLTracer) Init(ctx context.Context, p *graphql.Params) context.Context {
	t.result.StartTime = time.Now()
	return ctx
}

// Name is "opentracing"
func (t *GQLTracer) Name() string {
	return "tracing"
}

// HasResult is false since this extension does not return any
func (t *GQLTracer) HasResult() bool {
	return true
}

// GetResult returns nil
func (t *GQLTracer) GetResult(ctx context.Context) interface{} {
	return t.result
}

// ParseDidStart implements the graphql-go/graphql.Extension.ParseDidStart function
func (t *GQLTracer) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	pr := &ParsingResult{}
	pr.StartOffset = time.Since(t.result.StartTime)
	return ctx, func(err error) {
		pr.Duration = time.Since(t.result.StartTime.Add(pr.StartOffset))
		t.result.Parsing = pr
	}
}

// ValidationDidStart implements the graphql-go/graphql.Extension.ValidationDidStart function
func (t *GQLTracer) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	vr := &ValidationResult{}
	vr.StartOffset = time.Since(t.result.StartTime)
	return ctx, func(errs []gqlerrors.FormattedError) {
		vr.Duration = time.Since(t.result.StartTime.Add(vr.StartOffset))
		t.result.Validation = vr
	}
}

// ExecutionDidStart implements the graphql-go/graphql.Extension.ExecutionDidStart function
func (t *GQLTracer) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	return ctx, func(*graphql.Result) {
		t.result.EndTime = time.Now()
		t.result.Duration = t.result.EndTime.Sub(t.result.StartTime)
	}
}

// ResolveFieldDidStart implements the graphql-go/graphql.Extension.ResolveFieldDidStart function
func (t *GQLTracer) ResolveFieldDidStart(ctx context.Context, i *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	r := &ResolverResult{
		StartOffset: time.Since(t.result.StartTime),
		Path:        i.Path.AsArray(),
		ParentType:  i.ParentType.String(),
		ReturnType:  i.ReturnType.String(),
	}
	return ctx, func(v interface{}, err error) {
		r.Duration = time.Since(t.result.StartTime.Add(r.StartOffset))
		t.result.Execution.AddResolverResult(r)
	}
}

// TracingResult is the structure that's added to the GraphQL call result
type TracingResult struct {
	Version    int               `json:"version"`
	StartTime  time.Time         `json:"startTime"`
	EndTime    time.Time         `json:"endTime"`
	Duration   time.Duration     `json:"duration"`
	Parsing    *ParsingResult    `json:"parsing"`
	Validation *ValidationResult `json:"validation"`
	Execution  *ExecutionResult  `json:"execution"`
}

// ParsingResult is the duration info about the parsing process
type ParsingResult struct {
	StartOffset time.Duration `json:"startOffset"`
	Duration    time.Duration `json:"duration"`
}

// ValidationResult is the duration info about the validation process
type ValidationResult struct {
	StartOffset time.Duration `json:"startOffset"`
	Duration    time.Duration `json:"duration"`
}

// ExecutionResult contains the tracing data about the resolvers
type ExecutionResult struct {
	Resolvers []*ResolverResult `json:"resolvers"`
	mu        *sync.Mutex
}

// AddResolverResult helps to add resolverResults in a thread safe way
func (er *ExecutionResult) AddResolverResult(r *ResolverResult) {
	er.mu.Lock()
	er.Resolvers = append(er.Resolvers, r)
	er.mu.Unlock()
}

// ResolverResult is the tracing info about the fieldResolve process
type ResolverResult struct {
	Path        []interface{} `json:"path"`
	ParentType  string        `json:"parentType"`
	FieldName   string        `json:"fieldName"`
	ReturnType  string        `json:"returnType"`
	StartOffset time.Duration `json:"startOffset"`
	Duration    time.Duration `json:"duration"`
}
