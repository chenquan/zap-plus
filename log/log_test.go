package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/chenquan/zap-plus/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestNewLogger(t *testing.T) {
	buffer := &bytes.Buffer{}
	err := NewLogger(&config.Config{
		Trace:  config.Trace{},
		Level:  "debug",
		Format: "text",
		Mode:   "console",
		Logger: lumberjack.Logger{},
	}, WithWriter(buffer))
	assert.NoError(t, err)

	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	assert.NoError(t, err)
	assert.True(t, strings.Contains(buffer.String(), "debug"))
	assert.True(t, strings.Contains(buffer.String(), "info"))
	assert.True(t, strings.Contains(buffer.String(), "warn"))
	assert.True(t, strings.Contains(buffer.String(), "error"))

	buffer.Reset()
	WithContext(context.Background()).Info("info")
	assert.True(t, strings.Contains(buffer.String(), "info"))

	buffer.Reset()
	ctx, span := otel.Tracer("11").Start(context.Background(), "any")
	WithContext(ctx).Info("info")
	assert.True(t, strings.Contains(buffer.String(), "info"))
	span.End()
}
