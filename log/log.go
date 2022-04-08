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
	"sync"

	"github.com/chenquan/zap-plus/config"
	"github.com/chenquan/zap-plus/trace"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	stdtrace "go.opentelemetry.io/otel/trace"
)

var (
	mu     = sync.RWMutex{}
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		os.Stdout, zap.InfoLevel))
)

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
		logger.Panic("unknown format")
	}
	logger = zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller())

	trace.StartAgent(&c.Trace)

	return
}

// WithContext release fields to a new logger.
// Plugins can use this method to release plugin name field.
func WithContext(ctx context.Context) *zap.Logger {
	spanId := spanIdFromContext(ctx)
	straceId := traceIdFromContext(ctx)

	if spanId == "" || straceId == "" {
		return logger
	}

	return logger.With(
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

func log() *zap.Logger {
	mu.RLock()
	l := logger
	mu.RUnlock()
	return l
}

func SetLog(log *zap.Logger) {
	mu.Lock()
	logger = log
	mu.Unlock()
}

func Info(msg string, fields ...zap.Field) {
	log().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	log().Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	log().Warn(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	log().Panic(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	log().Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	log().Fatal(msg, fields...)
}

func With(fields ...zap.Field) {
	log().With(fields...)
}

func Check(lvl zapcore.Level, msg string) {
	log().Check(lvl, msg)
}

func Named(s string) *zap.Logger {
	return log().Named(s)
}

func Core() zapcore.Core {
	return log().Core()
}

func Sugar() *zap.SugaredLogger {
	return log().Sugar()
}

func Sync() error {
	return log().Sync()
}

func WithOptions(opts ...zap.Option) *zap.Logger {
	return log().WithOptions(opts...)
}
