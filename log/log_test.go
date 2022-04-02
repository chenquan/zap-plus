package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/chenquan/zap-plus/config"
	"github.com/chenquan/zap-plus/trace"
	"github.com/stretchr/testify/assert"
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
	log := Logger()

	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	assert.NoError(t, err)
	assert.True(t, strings.Contains(buffer.String(), "debug"))
	assert.True(t, strings.Contains(buffer.String(), "info"))
	assert.True(t, strings.Contains(buffer.String(), "warn"))
	assert.True(t, strings.Contains(buffer.String(), "error"))

	buffer.Reset()
	log.WithContext(context.Background()).Info("info")
	assert.True(t, strings.Contains(buffer.String(), "info"))

	buffer.Reset()
	ctx, span := trace.Start(context.Background(), "any")
	log.WithContext(ctx).Info("info")
	assert.True(t, strings.Contains(buffer.String(), "info"))
	span.End()

	buffer.Reset()

	ctx, span = trace.Start(context.Background(), "any")
	log = LoggerModule("Module")
	log.WithContext(ctx).Info("info")
	assert.True(t, strings.Contains(buffer.String(), "info"))
	span.End()
}
