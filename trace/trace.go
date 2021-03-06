/*
 *    Copyright 2022 chenquan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package trace

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/chenquan/zap-plus/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	Jaeger = "jaeger"
	Zipkin = "zipkin"
)

var (
	ErrUnknownExporter = errors.New("unknown exporter error")
	tracer             = otel.Tracer("zap-plus")
	agents             = make(map[string]struct{})
	lock               sync.Mutex
	// GetTracerProvider is alis of otel.GetTracerProvider
	GetTracerProvider = otel.GetTracerProvider
)

// StartAgent starts a opentelemetry agent.
func StartAgent(c *config.Trace) {
	lock.Lock()
	defer lock.Unlock()

	_, ok := agents[c.Endpoint]
	if ok {
		return
	}

	// if error happens, let later calls run.
	if err := startAgent(c); err != nil {
		return
	}

	tracer = otel.Tracer(c.Name)

	agents[c.Endpoint] = struct{}{}
}

func startAgent(c *config.Trace) error {
	opts := []sdktrace.TracerProviderOption{
		// Set the sampling rate based on the parent span to 100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(c.Name))),
	}

	if len(c.Endpoint) > 0 {
		exp, err := createExporter(c)
		if err != nil {
			log.Println("opentelemetry exporter err", err)
			return err
		}

		// Always be sure to batch in production.
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Println("opentelemetry error", err)
	}))

	return nil
}

func createExporter(c *config.Trace) (sdktrace.SpanExporter, error) {
	// Just support jaeger and zipkin now, more for later
	switch c.Batcher {
	case Jaeger:
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.Endpoint)))
	case Zipkin:
		return zipkin.New(c.Endpoint)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownExporter, c.Batcher)
	}
}

// Start creates a span and a context.Context containing the newly-created span.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}
