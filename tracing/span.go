package tracing

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Span struct {
	name    string
	span    trace.Span
	context context.Context
	logger  *zap.Logger
}

// NewSpan returns a new span from the global tracer. Depending on the `cus`
// argument, the span is either a plain one or a customised one. Each resulting
// span must be completed with `defer span.End()` right after the call.
func NewSpan(ctx context.Context, name string, cus SpanCustomiser) (context.Context, Span) {
	ctx, span := otel.Tracer("goweather").Start(ctx, name)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00")
	encoderCfg.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(strings.ToUpper(level.String()))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	logger := zap.New(core)
	if cus == nil {
		return ctx, Span{
			name:    name,
			span:    span,
			context: ctx,
			logger:  logger,
		}
	}

	return ctx, Span{
		name:    name,
		span:    span,
		context: ctx,
		logger:  logger,
	}

}

// SpanFromContext returns the current span from a context. If you wish to avoid
// creating child spans for each operation and just rely on the parent span, use
// this function throughout the application. With such practise you will get
// flatter span tree as opposed to deeper version. You can always mix and match
// both functions.
func SpanFromContext(ctx context.Context) Span {
	return Span{
		context: ctx,
		logger:  nil,
		span:    trace.SpanFromContext(ctx),
	}
}

// AddSpanTags adds a new tags to the span. It will appear under "Tags" section
// of the selected span. Use this if you think the tag and its value could be
// useful while debugging.
func (s Span) AddSpanTags(tags map[string]string) {
	list := make([]attribute.KeyValue, len(tags))

	var i int
	for k, v := range tags {
		list[i] = attribute.Key(k).String(v)
		i++
	}

	s.span.SetAttributes(list...)
}

// AddSpanEvents adds a new events to the span. It will appear under the "Logs"
// section of the selected span. Use this if the event could mean anything
// valuable while debugging.
func (s Span) AddSpanEvents(name string, events map[string]string) {
	list := make([]trace.EventOption, len(events))

	var i int
	for k, v := range events {
		list[i] = trace.WithAttributes(attribute.Key(k).String(v))
		i++
	}

	s.span.AddEvent(name, list...)
}

// AddSpanError adds a new event to the span. It will appear under the "Logs"
// section of the selected span. This is not going to flag the span as "failed".
// Use this if you think you should log any exceptions such as critical, error,
// warning, caution etc. Avoid logging sensitive data!
func (s Span) AddSpanError(err error) {
	s.span.RecordError(err)
}

// FailSpan flags the span as "failed" and adds "error" label on listed trace.
// Use this after calling the `AddSpanError` function so that there is some sort
// of relevant exception logged against it.
func (s Span) FailSpan(msg string) {
	s.span.SetStatus(codes.Error, msg)
}

// SpanCustomiser is used to enforce custom span options. Any custom concrete
// span customiser type must implement this interface.
type SpanCustomiser interface {
	customise() []trace.SpanOption
}

func (s Span) Log(msg string) {
	traceid := s.span.SpanContext().TraceID().String()
	spanid := s.span.SpanContext().SpanID().String()
	s.logger.Info(msg,
		zap.String("span", s.name),
		zap.String("trace_id", traceid),
		zap.String("span_id", spanid))

}

func (s Span) End() {
	s.span.End()
}
