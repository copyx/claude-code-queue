package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jingikim/ccq/internal/config"
)

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	// 파일 없으면 기본값
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Prefix != "" {
		t.Errorf("expected empty prefix for new config, got %q", cfg.Prefix)
	}

	// 저장
	cfg.Prefix = "C-Space"
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// 다시 로드
	cfg2, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if cfg2.Prefix != "C-Space" {
		t.Errorf("expected 'C-Space', got %q", cfg2.Prefix)
	}
}

func TestDefaultPath(t *testing.T) {
	p := config.DefaultPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ccq", "config")
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}
