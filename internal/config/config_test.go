package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Port != 9393 {
		t.Errorf("expected port 9393, got %d", cfg.Server.Port)
	}
	if cfg.Database.Path != "formpilot.db" {
		t.Errorf("expected formpilot.db, got %s", cfg.Database.Path)
	}
	if cfg.ICP.MinQualificationScore != 60 {
		t.Errorf("expected min score 60, got %d", cfg.ICP.MinQualificationScore)
	}
}

func TestExpandEnvVars(t *testing.T) {
	_ = os.Setenv("TEST_FORMPILOT_KEY", "secret123")
	defer func() { _ = os.Unsetenv("TEST_FORMPILOT_KEY") }()

	input := "api_key: ${TEST_FORMPILOT_KEY}\nport: ${TEST_UNSET:-8080}"
	output := expandEnvVars(input)

	expected := "api_key: secret123\nport: 8080"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
