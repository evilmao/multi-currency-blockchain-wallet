package monitor

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// SpanTags is tags of span.
type SpanTags map[string]interface{}

func setSpanTags(span tracer.Span, tags SpanTags) {
	if len(tags) > 0 {
		for k, v := range tags {
			span.SetTag(k, v)
		}
	}
}

// StartDDSpan starts a datadog span.
func StartDDSpan(operationName string, parentSpan tracer.Span, spanType string, tags SpanTags) tracer.Span {
	var span tracer.Span
	if parentSpan != nil {
		span = tracer.StartSpan(operationName, tracer.ChildOf(parentSpan.Context()))
	} else {
		span = tracer.StartSpan(operationName)
	}

	if len(spanType) > 0 {
		tags[ext.SpanType] = spanType
	}

	setSpanTags(span, tags)
	return span
}

// DDTraceResult provides result.
type DDTraceResult func() (SpanTags, error)

// DeferFinishDDSpan finishes a datadog span.
func DeferFinishDDSpan(span tracer.Span, result DDTraceResult) func() {
	return func() {
		tags, err := result()
		setSpanTags(span, tags)
		span.Finish(tracer.WithError(err))
	}
}

// DDTraceTarget is the func to be traced by datadog.
type DDTraceTarget func(tracer.Span) (SpanTags, error)

// WithDDTracer traces span info by DataDog.
func WithDDTracer(operationName string, parentSpan tracer.Span, spanType string, tags SpanTags, f DDTraceTarget) {
	var (
		extTags SpanTags
		err     error
	)
	span := StartDDSpan(operationName, parentSpan, spanType, tags)
	defer DeferFinishDDSpan(span, func() (SpanTags, error) {
		return extTags, err
	})()

	extTags, err = f(span)
}

// DDTraceSimpleTarget is the func to be traced by datadog.
type DDTraceSimpleTarget func() (SpanTags, error)

// WithSimpleDDTracer traces span info by DataDog.
func WithSimpleDDTracer(operationName string, tags SpanTags, f DDTraceSimpleTarget) {
	WithDDTracer(operationName, nil, "", tags, func(tracer.Span) (SpanTags, error) {
		return f()
	})
}
