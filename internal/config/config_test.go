package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadAndMethods(t *testing.T) {
	dbCfg := &DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "user",
		Password: "password",
		DBName:   "testdb",
		SSLMode:  "disable",
	}
	expectedDSN := "host=127.0.0.1 port=5432 user=user password=password dbname=testdb sslmode=disable"
	if dbCfg.DSN() != expectedDSN {
		t.Errorf("expected DSN %s, got %s", expectedDSN, dbCfg.DSN())
	}

	redisCfg := &RedisConfig{Host: "localhost", Port: "6379"}
	if redisCfg.Address() != "localhost:6379" {
		t.Errorf("expected redis address localhost:6379, got %s", redisCfg.Address())
	}

	serverCfg := &ServerConfig{Host: "0.0.0.0", Port: "8080"}
	if serverCfg.Address() != "0.0.0.0:8080" {
		t.Errorf("expected server address 0.0.0.0:8080, got %s", serverCfg.Address())
	}

	// Environment variable overrides
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("SERVER_READ_TIMEOUT", "30s")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("SERVER_READ_TIMEOUT")
	}()

	cfg := Load()
	if cfg.Server.Port != "9090" {
		t.Errorf("expected SERVER_PORT 9090, got %s", cfg.Server.Port)
	}

	// Helper tests
	if val := getEnv("NON_EXISTENT_KEY", "fallback"); val != "fallback" {
		t.Errorf("expected fallback")
	}

	// getIntEnv
	os.Setenv("TEST_INT", "42")
	os.Setenv("TEST_BAD_INT", "notanint")
	defer func() {
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_BAD_INT")
	}()
	if getIntEnv("TEST_INT", 10) != 42 {
		t.Errorf("expected 42")
	}
	if getIntEnv("TEST_BAD_INT", 10) != 10 {
		t.Errorf("expected 10")
	}

	// getBoolEnv
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_BAD_BOOL", "notabool")
	defer func() {
		os.Unsetenv("TEST_BOOL")
		os.Unsetenv("TEST_BAD_BOOL")
	}()
	if !getBoolEnv("TEST_BOOL", false) {
		t.Errorf("expected true")
	}
	if getBoolEnv("TEST_BAD_BOOL", false) {
		t.Errorf("expected false")
	}

	// getDurationEnv
	os.Setenv("TEST_DUR", "10m")
	os.Setenv("TEST_BAD_DUR", "notadur")
	defer func() {
		os.Unsetenv("TEST_DUR")
		os.Unsetenv("TEST_BAD_DUR")
	}()
	if getDurationEnv("TEST_DUR", time.Second) != 10*time.Minute {
		t.Errorf("expected 10m")
	}
	if getDurationEnv("TEST_BAD_DUR", time.Second) != time.Second {
		t.Errorf("expected 1s")
	}
}
