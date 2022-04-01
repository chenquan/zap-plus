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
	Info  = logger.Info
	Panic = logger.Panic
	Error = logger.Error
	Warn  = logger.Warn
	Debug = logger.Debug
	Fatal = logger.Fatal
)

type Log struct {
	*zap.Logger
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

func NewLogger(c *config.Config) (err error) {
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
	if c.Mode == "file" {
		w = append(w, zapcore.AddSync(&c.Logger))
	} else if c.Mode == "console" {
		w = append(w, zapcore.AddSync(os.Stdout))
	} else {
		w = append(w, zapcore.AddSync(&c.Logger), zapcore.AddSync(os.Stdout))
	}

	var core zapcore.Core
	if c.Format == "json" {
		core = zapcore.NewCore(zapcore.NewJSONEncoder(
			zap.NewProductionEncoderConfig()),
			zapcore.NewMultiWriteSyncer(w...),
			logLevel)
	} else if c.Format == "text" {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.NewMultiWriteSyncer(w...), logLevel)
	} else {
		core = zapcore.NewNopCore()
	}
	logger = zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller())

	trace.StartAgent(&c.Trace)

	return
}
