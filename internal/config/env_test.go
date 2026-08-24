package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfigFileRespectsSPMConfigOverride(t *testing.T) {
	custom := filepath.Join("custom", "spm.toml")
	t.Setenv("SPM_CONFIG", custom)

	if got := UserConfigFile(); got != custom {
		t.Errorf("UserConfigFile() = %q, want override %q", got, custom)
	}
}

func TestDBPathRespectsSPMDBOverride(t *testing.T) {
	custom := filepath.Join("data", "other.db")
	t.Setenv("SPM_DB", custom)

	if got := DBPath(); got != custom {
		t.Errorf("DBPath() = %q, want override %q", got, custom)
	}
}

func TestNoTUI(t *testing.T) {
	t.Setenv("SPM_NO_TUI", "1")
	if !NoTUI() {
		t.Error("NoTUI() = false with SPM_NO_TUI=1, want true")
	}

	t.Setenv("SPM_NO_TUI", "")
	if NoTUI() {
		t.Error("NoTUI() = true when SPM_NO_TUI is empty, want false")
	}
}

func TestEditorPrefersSPMEditorOverFallback(t *testing.T) {
	t.Setenv("SPM_EDITOR", "hx")
	if got := Editor("vim"); got != "hx" {
		t.Errorf("Editor() = %q, want %q from SPM_EDITOR", got, "hx")
	}

	os.Unsetenv("SPM_EDITOR")
	if got := Editor("vim"); got != "vim" {
		t.Errorf("Editor() = %q without SPM_EDITOR, want fallback %q", got, "vim")
	}
}

func TestLogLevel(t *testing.T) {
	t.Setenv("SPM_LOG_LEVEL", "debug")
	if got := LogLevel(); got != "debug" {
		t.Errorf("LogLevel() = %q, want %q", got, "debug")
	}
}
