package trace

import (
	"context"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/onflow/flow-go/model/flow"
)

const DefaultEntityCacheSize = 1000

const SensitivityCaptureAll = 0
const EntityTypeBlock = "Block"
const EntityTypeCollection = "Collection"
const EntityTypeTransaction = "Transaction"

type SpanName string

func (s SpanName) Child(subOp string) SpanName {
	return SpanName(string(s) + "." + subOp)
}

// Tracer is the implementation of the Tracer interface
type Tracer struct {
	tracer      trace.Tracer
	shutdown    func(context.Context) error
	log         zerolog.Logger
	spanCache   *lru.Cache
	chainID     string
	samplingPct int
}

// NewTracer creates a new tracer.
func NewTracer(
	log zerolog.Logger,
	serviceName string,
	chainID string,
	samplingPct int,
) (*Tracer, error) {
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		//sdktrace.WithSampler(IdentitySampler(samplingPct)), // XXX
		//sdktrace.WithIDGenerator(defaultIDGenerator()),     // XXX
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Debug().Err(err).Msg("tracing error")
	}))

	spanCache, err := lru.New(int(DefaultEntityCacheSize))
	if err != nil {
		return nil, err
	}

	return &Tracer{
		tracer:      tracerProvider.Tracer(""),
		shutdown:    tracerProvider.Shutdown,
		log:         log,
		spanCache:   spanCache,
		samplingPct: samplingPct,
		chainID:     chainID,
	}, nil
}

// Ready returns a channel that will close when the network stack is ready.
func (t *Tracer) Ready() <-chan struct{} {
	ready := make(chan struct{})
	go func() {
		close(ready)
	}()
	return ready
}

// Done returns a channel that will close when shutdown is complete.
func (t *Tracer) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		if err := t.shutdown(ctx); err != nil {
			t.log.Error().Err(err).Msg("failed to shutdown tracer")
		}

		t.spanCache.Purge()
		close(done)
	}()
	return done
}

// entityRootSpan returns the root span for the given entity from the cache
// and if not exist it would construct it and cache it and return it
// This should be used mostly for the very first span created for an entity on the service
func (t *Tracer) entityRootSpan(
	ctx context.Context,
	entityID flow.Identifier,
	entityType string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	if c, ok := t.spanCache.Get(entityID); ok {
		span := c.(trace.Span)
		return trace.ContextWithSpan(ctx, span), span
	}

	traceID := (*trace.TraceID)(entityID[:16])
	spanID := (*trace.SpanID)(entityID[16:])

	spanConfig := trace.SpanContextConfig{
		TraceID:    *traceID,
		SpanID:     *spanID,
		TraceFlags: trace.TraceFlags(0).WithSampled(true),
	}
	spanCtx := trace.NewSpanContext(spanConfig)
	trace.ContextWithSpanContext(ctx, spanCtx)
	ctx, span := t.tracer.Start(ctx, string(entityType), trace.WithNewRoot())

	span.SetAttributes(attribute.String("entity_id", entityID.String()))
	span.SetAttributes(attribute.String("chainID", t.chainID))
	t.spanCache.Add(entityID, span)

	span.End() // finish span right away
	return ctx, span
}

func (t *Tracer) StartBlockSpan(
	ctx context.Context,
	blockID flow.Identifier,
	spanName SpanName,
	opts ...trace.SpanStartOption) (trace.Span, context.Context, bool) {

	//if !blockID.IsSampled(t.samplingPct) {
	//	return &NoopSpan{&NoopTracer{}}, ctx, false
	//}

	ctx, rootSpan := t.entityRootSpan(ctx, blockID, EntityTypeBlock)
	return t.StartSpanFromParent(rootSpan, spanName, opts...), ctx, true
}

func (t *Tracer) StartCollectionSpan(
	ctx context.Context,
	collectionID flow.Identifier,
	spanName SpanName,
	opts ...trace.SpanStartOption) (trace.Span, context.Context, bool) {

	// if !collectionID.IsSampled(t.samplingPct) {
	// 	return &NoopSpan{&NoopTracer{}}, ctx, false
	// }

	ctx, rootSpan := t.entityRootSpan(ctx, collectionID, EntityTypeCollection)
	return t.StartSpanFromParent(rootSpan, spanName, opts...), ctx, true
}

// StartTransactionSpan starts a span that will be aggregated under the given transaction.
// All spans for the same transaction will be aggregated under a root span
func (t *Tracer) StartTransactionSpan(
	ctx context.Context,
	transactionID flow.Identifier,
	spanName SpanName,
	opts ...trace.SpanStartOption) (trace.Span, context.Context, bool) {

	// if !transactionID.IsSampled(uint(t.samplingPct)) {
	// 	return &NoopSpan{&NoopTracer{}}, ctx, false
	// }

	ctx, rootSpan := t.entityRootSpan(ctx, transactionID, EntityTypeTransaction)
	return t.StartSpanFromParent(rootSpan, spanName, opts...), ctx, true
}

func (t *Tracer) StartSpanFromContext(
	ctx context.Context,
	operationName SpanName,
	opts ...trace.SpanStartOption,
) (trace.Span, context.Context) {
	//parentSpan := trace.SpanFromContext(ctx)
	// if parentSpan == nil {
	// 	return &NoopSpan{&NoopTracer{}}, ctx
	// }
	// if _, ok := parentSpan.(*NoopSpan); ok {
	// 	return &NoopSpan{&NoopTracer{}}, ctx
	// }

	ctx, span := t.tracer.Start(ctx, string(operationName), opts...)
	return span, ctx
}

func (t *Tracer) StartSpanFromParent(
	parentSpan trace.Span,
	operationName SpanName,
	opts ...trace.SpanStartOption,
) trace.Span {
	// if _, ok := span.(*NoopSpan); ok {
	// 	return &NoopSpan{&NoopTracer{}}
	// }

	ctx := trace.ContextWithSpanContext(context.Background(), parentSpan.SpanContext())
	_, span := t.tracer.Start(ctx, string(operationName), opts...)
	return span
}

func (t *Tracer) RecordSpanFromParent(
	parentSpan trace.Span,
	operationName SpanName,
	duration time.Duration,
	attributes []attribute.KeyValue,
	opts ...trace.SpanStartOption,
) {
	// if _, ok := span.(*NoopSpan); ok {
	// 	return
	// }

	end := time.Now()
	start := end.Add(-duration)

	ctx := trace.ContextWithSpanContext(context.Background(), parentSpan.SpanContext())
	opts = append(opts, trace.WithAttributes(attributes...), trace.WithTimestamp(start))
	_, span := t.tracer.Start(ctx, string(operationName), opts...)
	span.End(trace.WithTimestamp(end))
}

// WithSpanFromContext encapsulates executing a function within an span, i.e., it starts a span with the specified SpanName from the context,
// executes the function f, and finishes the span once the function returns.
func (t *Tracer) WithSpanFromContext(ctx context.Context,
	operationName SpanName,
	f func(),
	opts ...trace.SpanStartOption) {
	span, _ := t.StartSpanFromContext(ctx, operationName, opts...)
	defer span.End()

	f()
}
