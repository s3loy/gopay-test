package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "gopay" {
		t.Errorf("App.Name = %v, want gopay", cfg.App.Name)
	}
	if cfg.App.Version != "1.0.0" {
		t.Errorf("App.Version = %v, want 1.0.0", cfg.App.Version)
	}
	if cfg.App.Env != "development" {
		t.Errorf("App.Env = %v, want development", cfg.App.Env)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %v, want disable", cfg.Database.SSLMode)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("Database.MaxOpenConns = %v, want 25", cfg.Database.MaxOpenConns)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %v, want info", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Log.Format = %v, want json", cfg.Log.Format)
	}
	if cfg.Log.Output != "stdout" {
		t.Errorf("Log.Output = %v, want stdout", cfg.Log.Output)
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  name: test-app
  version: 2.0.0
  env: test
  debug: true
server:
  port: 9090
  host: 127.0.0.1
database:
  host: test-host
  port: 5433
  user: testuser
  password: testpass
  dbname: testdb
  sslmode: require
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 10m
log:
  level: debug
  format: console
  output: stdout
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "test-app" {
		t.Errorf("App.Name = %v, want test-app", cfg.App.Name)
	}
	if cfg.App.Version != "2.0.0" {
		t.Errorf("App.Version = %v, want 2.0.0", cfg.App.Version)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %v, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %v, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Database.Host != "test-host" {
		t.Errorf("Database.Host = %v, want test-host", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %v, want 5433", cfg.Database.Port)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("Database.User = %v, want testuser", cfg.Database.User)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("Database.Password = %v, want testpass", cfg.Database.Password)
	}
	if cfg.Database.DBName != "testdb" {
		t.Errorf("Database.DBName = %v, want testdb", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Database.SSLMode = %v, want require", cfg.Database.SSLMode)
	}
	if cfg.Database.MaxOpenConns != 50 {
		t.Errorf("Database.MaxOpenConns = %v, want 50", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 10 {
		t.Errorf("Database.MaxIdleConns = %v, want 10", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime != 10*time.Minute {
		t.Errorf("Database.ConnMaxLifetime = %v, want 10m", cfg.Database.ConnMaxLifetime)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %v, want debug", cfg.Log.Level)
	}
	if cfg.Log.Format != "console" {
		t.Errorf("Log.Format = %v, want console", cfg.Log.Format)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "gopay",
		Password: "secret",
		DBName:   "gopay",
		SSLMode:  "disable",
	}

	want := "host=localhost port=5432 user=gopay password=secret dbname=gopay sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %v, want %v", got, want)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("GOPAY_APP_NAME", "env-app")
	t.Setenv("GOPAY_SERVER_PORT", "7777")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "env-app" {
		t.Errorf("App.Name = %v, want env-app", cfg.App.Name)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("Server.Port = %v, want 7777", cfg.Server.Port)
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}
