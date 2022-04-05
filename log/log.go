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

package log

import (
	"context"
	"io"
	"os"

	"github.com/chenquan/zap-plus/config"
	"github.com/chenquan/zap-plus/trace"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	stdtrace "go.opentelemetry.io/otel/trace"
)

var (
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		os.Stdout, zap.InfoLevel))
	Info        = logger.Info
	Panic       = logger.Panic
	Error       = logger.Error
	Warn        = logger.Warn
	Debug       = logger.Debug
	Fatal       = logger.Fatal
	With        = logger.With
	Check       = logger.Check
	Named       = logger.Named
	Core        = logger.Core
	Sugar       = logger.Sugar
	Sync        = logger.Sync
	WithOptions = logger.WithOptions
)

type Log struct {
	*zap.Logger
}

type options struct {
	w io.Writer
}

type Option func(*options)

func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.w = w
	}
}

func NewLogger(c *config.Config, opts ...Option) (err error) {
	validate := validator.New()
	err = validate.Struct(c)
	if err != nil {
		return
	}

	var logLevel zapcore.Level
	err = logLevel.UnmarshalText([]byte(c.Level))
	if err != nil {
		return
	}

	var w []zapcore.WriteSyncer
	switch c.Mode {
	case "file":
		w = append(w, zapcore.AddSync(&c.Logger))
	case "console":
		w = append(w, zapcore.AddSync(os.Stdout))
	default:
		w = append(w, zapcore.AddSync(&c.Logger), zapcore.AddSync(os.Stdout))
	}

	o := new(options)
	for _, opt := range opts {
		opt(o)
	}
	if o.w != nil {
		w = append(w, zapcore.AddSync(o.w))
	}

	var core zapcore.Core
	switch c.Format {
	case "json":
		core = zapcore.NewCore(zapcore.NewJSONEncoder(
			zap.NewProductionEncoderConfig()),
			zapcore.NewMultiWriteSyncer(w...),
			logLevel)
	case "text":
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.NewMultiWriteSyncer(w...), logLevel)
	default:
		core = zapcore.NewNopCore()
	}
	logger = zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller())

	trace.StartAgent(&c.Trace)

	return
}

// LoggerModule release fields to a new logger.
// Plugins can use this method to release plugin name field.
func LoggerModule(moduleName string) *Log {
	return &Log{
		Logger: logger.With(zap.String("moduleName", moduleName)),
	}
}

// Logger release fields to a new logger.
func Logger() *Log {
	return &Log{
		Logger: logger,
	}
}

// WithContext release fields to a new logger.
// Plugins can use this method to release plugin name field.
func (l *Log) WithContext(ctx context.Context) *zap.Logger {
	spanId := spanIdFromContext(ctx)
	straceId := traceIdFromContext(ctx)

	if spanId != "" || straceId != "" {
		return l.Logger
	}

	return l.With(
		zap.String("traceId", straceId),
		zap.String("spanId", spanId),
	)
}

func spanIdFromContext(ctx context.Context) string {
	spanCtx := stdtrace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func traceIdFromContext(ctx context.Context) string {
	spanCtx := stdtrace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}
