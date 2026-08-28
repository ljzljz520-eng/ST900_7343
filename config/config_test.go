package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "127.0.0.1:8080" || cfg.DatabasePath != "studio.db" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
