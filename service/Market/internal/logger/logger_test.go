package logger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger_DebugLevel(t *testing.T) {
	logger := InitLogger("debug")
	require.NotNil(t, logger)
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel())
}

func TestInitLogger_InfoLevel(t *testing.T) {
	logger := InitLogger("info")
	require.NotNil(t, logger)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

func TestInitLogger_WarnLevel(t *testing.T) {
	logger := InitLogger("warn")
	require.NotNil(t, logger)
	assert.Equal(t, logrus.WarnLevel, logger.GetLevel())
}

func TestInitLogger_ErrorLevel(t *testing.T) {
	logger := InitLogger("error")
	require.NotNil(t, logger)
	assert.Equal(t, logrus.ErrorLevel, logger.GetLevel())
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	logger := InitLogger("invalid")
	require.NotNil(t, logger)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

func TestGetLogger_Singleton(t *testing.T) {
	Log = nil
	logger1 := GetLogger()
	logger2 := GetLogger()
	assert.Same(t, logger1, logger2)
}

func TestGetLogger_InitializesIfNil(t *testing.T) {
	Log = nil
	logger := GetLogger()
	require.NotNil(t, logger)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

func TestLogger_JSONFormatter(t *testing.T) {
	logger := InitLogger("info")
	require.NotNil(t, logger)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "level")
	assert.Contains(t, output, "msg")
}

func TestLogger_WithField(t *testing.T) {
	logger := InitLogger("info")
	require.NotNil(t, logger)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.WithField("key", "value").Info("test with field")

	output := buf.String()
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

func TestLogger_WithFields(t *testing.T) {
	logger := InitLogger("info")
	require.NotNil(t, logger)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.WithFields(logrus.Fields{
		"user_id":    123,
		"request_id": "abc-123",
	}).Info("test with fields")

	output := buf.String()
	assert.Contains(t, output, "user_id")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "request_id")
}

func TestLogger_ErrorLogging(t *testing.T) {
	logger := InitLogger("error")
	require.NotNil(t, logger)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Error("error message")

	output := buf.String()
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "error message")
}

func TestLogger_DebugNotLoggedAtInfoLevel(t *testing.T) {
	logger := InitLogger("info")
	require.NotNil(t, logger)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Debug("debug message")

	output := buf.String()
	assert.Empty(t, output)
}

func TestLogger_AllLevels(t *testing.T) {
	levels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger := InitLogger(level)
			require.NotNil(t, logger)
		})
	}
}
