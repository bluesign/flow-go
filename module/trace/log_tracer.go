package trace

// import (
// 	"context"
// 	"math/rand"
// 	"time"

// 	"github.com/opentracing/opentracing-go/log"
// 	"github.com/rs/zerolog"
// 	"go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/trace"

// 	"github.com/onflow/flow-go/model/flow"
// )

// type spanKey string

// const activeSpan spanKey = "activeSpan"

// // LogTracer is the implementation of the Tracer interface which passes
// // all the traces back to the passed logger and print them
// // this is mostly useful for debugging and testing
// type LogTracer struct {
// 	log zerolog.Logger
// }

// // LogTracer creates a new tracer.
// func NewLogTracer(log zerolog.Logger) *LogTracer {
// 	trace.Stdo
// 	return &LogTracer{log: log}
// }

// func (t *LogTracer) Ready() <-chan struct{} {
// 	ready := make(chan struct{})
// 	close(ready)
// 	return ready
// }

// func (t *LogTracer) Done() <-chan struct{} {
// 	done := make(chan struct{})
// 	close(done)
// 	return done
// }

// func (t *LogTracer) StartBlockSpan(
// 	ctx context.Context,
// 	blockID flow.Identifier,
// 	spanName SpanName,
// 	opts ...trace.StartSpanOption) (trace.Span, context.Context, bool) {
// 	sp := NewLogSpan(t, spanName)
// 	ctx = context.WithValue(ctx, activeSpan, sp.spanID)
// 	return sp, ctx, true
// }

// func (t *LogTracer) StartCollectionSpan(
// 	ctx context.Context,
// 	collectionID flow.Identifier,
// 	spanName SpanName,
// 	opts ...trace.StartSpanOption) (trace.Span, context.Context, bool) {
// 	sp := NewLogSpan(t, spanName)
// 	ctx = context.WithValue(ctx, activeSpan, sp.spanID)
// 	return sp, ctx, true
// }

// // StartTransactionSpan starts a span that will be aggregated under the given transaction.
// // All spans for the same transaction will be aggregated under a root span
// func (t *LogTracer) StartTransactionSpan(
// 	ctx context.Context,
// 	transactionID flow.Identifier,
// 	spanName SpanName,
// 	opts ...trace.StartSpanOption) (trace.Span, context.Context, bool) {
// 	sp := NewLogSpan(t, spanName)
// 	ctx = context.WithValue(ctx, activeSpan, sp.spanID)
// 	return sp, ctx, true
// }

// func (t *LogTracer) StartSpanFromContext(
// 	ctx context.Context,
// 	operationName SpanName,
// 	opts ...trace.StartSpanOption,
// ) (trace.Span, context.Context) {
// 	parentSpanID := ctx.Value(activeSpan).(uint64)
// 	sp := NewLogSpanWithParent(t, operationName, parentSpanID)
// 	ctx = context.WithValue(ctx, activeSpan, sp.spanID)
// 	return sp, trace.ContextWithSpan(ctx, sp)
// }

// func (t *LogTracer) StartSpanFromParent(
// 	span trace.Span,
// 	operationName SpanName,
// 	opts ...trace.StartSpanOption,
// ) trace.Span {
// 	parentSpan := span.(*LogSpan)
// 	return NewLogSpanWithParent(t, operationName, parentSpan.spanID)
// }

// func (t *LogTracer) RecordSpanFromParent(
// 	span trace.Span,
// 	operationName SpanName,
// 	duration time.Duration,
// 	attributes []attribute.KeyValue,
// 	opts ...trace.StartSpanOption,
// ) {
// 	parentSpan := span.(*LogSpan)
// 	sp := NewLogSpanWithParent(t, operationName, parentSpan.spanID)
// 	sp.start = time.Now().Add(-duration)
// 	span.End()
// }

// // WithSpanFromContext encapsulates executing a function within an span, i.e., it starts a span with the specified SpanName from the context,
// // executes the function f, and finishes the span once the function returns.
// func (t *LogTracer) WithSpanFromContext(ctx context.Context,
// 	operationName SpanName,
// 	f func(),
// 	opts ...trace.StartSpanOption) {
// 	span, _ := t.StartSpanFromContext(ctx, operationName, opts...)
// 	defer span.End()

// 	f()
// }

// type LogSpan struct {
// 	tracer        *LogTracer
// 	spanID        uint64
// 	parentID      uint64
// 	operationName SpanName
// 	start         time.Time
// 	end           time.Time
// 	tags          map[string]interface{}
// }

// func NewLogSpan(tracer *LogTracer, operationName SpanName) *LogSpan {
// 	return &LogSpan{
// 		tracer:        tracer,
// 		spanID:        rand.Uint64(),
// 		operationName: operationName,
// 		start:         time.Now(),
// 		tags:          make(map[string]interface{}),
// 	}
// }

// func NewLogSpanWithParent(tracer *LogTracer, operationName SpanName, parentSpanID uint64) *LogSpan {
// 	sp := NewLogSpan(tracer, operationName)
// 	sp.parentID = parentSpanID
// 	return sp
// }

// func (s *LogSpan) ProduceLog() {
// 	s.tracer.log.Info().
// 		Uint64("spanID", s.spanID).
// 		Uint64("parent", s.parentID).
// 		Time("start", s.start).
// 		Time("end", s.end).
// 		Msgf("Span %s (duration %d ms)", s.operationName, s.end.Sub(s.start).Milliseconds())
// }

// func (s *LogSpan) Finish() {
// 	s.end = time.Now()
// 	s.ProduceLog()
// }
// func (s *LogSpan) FinishWithOptions(opts trace.FinishOptions) {
// 	// TODO support finish options
// 	s.Finish()
// }
// func (s *LogSpan) Context() trace.SpanContext {
// 	return &NoopSpanContext{}
// }
// func (s *LogSpan) SetOperationName(operationName string) trace.Span {
// 	s.operationName = SpanName(operationName)
// 	return s
// }
// func (s *LogSpan) SetTag(key string, value interface{}) trace.Span {
// 	s.tags[key] = value
// 	return s
// }
// func (s *LogSpan) LogFields(fields ...log.Field) {
// 	for _, f := range fields {
// 		s.tags[f.Key()] = f.Value()
// 	}
// }
// func (s *LogSpan) LogKV(alternatingKeyValues ...interface{})             {}
// func (s *LogSpan) SetBaggageItem(restrictedKey, value string) trace.Span { return s }
// func (s *LogSpan) BaggageItem(restrictedKey string) string               { return "" }
// func (s *LogSpan) Tracer() trace.Tracer                                  { return nil }
// func (s *LogSpan) LogEvent(event string)                                 {}
// func (s *LogSpan) LogEventWithPayload(event string, payload interface{}) {}
// func (s *LogSpan) Log(data trace.LogData)                                {}
