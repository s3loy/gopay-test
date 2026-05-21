package logger

import (
	"testing"

	"github.com/s3loy/gopay/internal/pkg/config"
	"go.uber.org/zap/zapcore"
)

func TestInit_ConsoleFormat(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
	log, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if log == nil {
		t.Fatal("Init() returned nil logger")
	}
	if L() == nil {
		t.Error("L() should not return nil after Init")
	}
}

func TestInit_JSONFormat(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	log, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if log == nil {
		t.Fatal("Init() returned nil logger")
	}
}

func TestInit_FileOutput(t *testing.T) {
	cfg := config.LogConfig{
		Level:      "warn",
		Format:     "json",
		Output:     "file",
		FilePath:   t.TempDir() + "/test.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
	log, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if log == nil {
		t.Fatal("Init() returned nil logger")
	}
}

func TestInit_InvalidDir(t *testing.T) {
	cfg := config.LogConfig{
		Level:    "info",
		Format:   "json",
		Output:   "file",
		FilePath: "/nonexistent/readonly/test.log",
	}
	_, err := Init(cfg)
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"fatal", zapcore.FatalLevel},
		{"unknown", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseLevel(tt.input); got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildEncoder(t *testing.T) {
	consoleEncoder := buildEncoder("console")
	if consoleEncoder == nil {
		t.Error("buildEncoder(console) should not return nil")
	}

	jsonEncoder := buildEncoder("json")
	if jsonEncoder == nil {
		t.Error("buildEncoder(json) should not return nil")
	}

	defaultEncoder := buildEncoder("unknown")
	if defaultEncoder == nil {
		t.Error("buildEncoder(unknown) should default to json and not return nil")
	}
}

func TestL_BeforeInit(t *testing.T) {
	// Reset global logger
	oldGlobal := global
	global = nil
	defer func() { global = oldGlobal }()

	log := L()
	if log == nil {
		t.Error("L() should return nop logger, not nil")
	}
}
