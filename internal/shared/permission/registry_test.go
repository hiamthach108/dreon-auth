package permission

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hiamthach108/dreon-auth/config"
)

func TestNewRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "permissions.json")
	err := os.WriteFile(path, []byte(`[
		{"name": "View", "code": "view"},
		{"name": "Edit", "code": "edit"}
	]`), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	r, err := NewRegistry(path)
	if err != nil {
		t.Fatalf("NewRegistry() err = %v", err)
	}
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	list := r.List()
	if len(list) != 2 {
		t.Errorf("List() len = %d, want 2", len(list))
	}
}

func TestNewRegistry_missingFile(t *testing.T) {
	_, err := NewRegistry("/nonexistent/path.json")
	if err == nil {
		t.Fatal("NewRegistry(missing file) err = nil, want non-nil")
	}
}

func TestNewRegistry_invalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	err := os.WriteFile(path, []byte("not json"), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	_, err = NewRegistry(path)
	if err == nil {
		t.Fatal("NewRegistry(invalid JSON) err = nil, want non-nil")
	}
}

func TestRegistry_List(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.json")
	err := os.WriteFile(path, []byte(`[{"name": "A", "code": "a"}, {"name": "B", "code": "b"}]`), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	r, err := NewRegistry(path)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	list := r.List()
	if len(list) != 2 {
		t.Fatalf("List() len = %d, want 2", len(list))
	}
	codes := make(map[string]bool)
	for _, p := range list {
		codes[p.Code] = true
	}
	if !codes["a"] || !codes["b"] {
		t.Errorf("List() missing expected codes, got %v", list)
	}
}

func TestRegistry_List_nilReceiver(t *testing.T) {
	var r *Registry
	list := r.List()
	if list != nil {
		t.Errorf("(*Registry)(nil).List() = %v, want nil", list)
	}
}

func TestRegistry_ValidateCodes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.json")
	err := os.WriteFile(path, []byte(`[{"name": "View", "code": "view"}, {"name": "Edit", "code": "edit"}]`), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	r, err := NewRegistry(path)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	if err := r.ValidateCodes([]string{"view", "edit"}); err != nil {
		t.Errorf("ValidateCodes(valid) err = %v, want nil", err)
	}
	if err := r.ValidateCodes([]string{"view"}); err != nil {
		t.Errorf("ValidateCodes(single valid) err = %v, want nil", err)
	}
	if err := r.ValidateCodes(nil); err != nil {
		t.Errorf("ValidateCodes(nil) err = %v, want nil", err)
	}
	if err := r.ValidateCodes([]string{""}); err != nil {
		t.Errorf("ValidateCodes(empty string skipped) err = %v, want nil", err)
	}

	err = r.ValidateCodes([]string{"unknown"})
	if err == nil {
		t.Fatal("ValidateCodes(invalid code) err = nil, want non-nil")
	}
	if err != nil && err.Error() != "invalid permission code: unknown" {
		t.Errorf("ValidateCodes(invalid) err = %v", err)
	}
}

func TestRegistry_ValidateCodes_nilReceiver(t *testing.T) {
	var r *Registry
	if err := r.ValidateCodes([]string{"any"}); err != nil {
		t.Errorf("(*Registry)(nil).ValidateCodes err = %v, want nil", err)
	}
}

func TestRegistry_GetByCode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.json")
	err := os.WriteFile(path, []byte(`[{"name": "View Users", "code": "users.view"}]`), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	r, err := NewRegistry(path)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	p, ok := r.GetByCode("users.view")
	if !ok {
		t.Fatal("GetByCode(users.view) ok = false, want true")
	}
	if p.Code != "users.view" || p.Name != "View Users" {
		t.Errorf("GetByCode() = %+v", p)
	}

	_, ok = r.GetByCode("missing")
	if ok {
		t.Error("GetByCode(missing) ok = true, want false")
	}
}

func TestRegistry_GetByCode_nilReceiver(t *testing.T) {
	var r *Registry
	p, ok := r.GetByCode("any")
	if ok {
		t.Error("(*Registry)(nil).GetByCode ok = true, want false")
	}
	if p.Code != "" || p.Name != "" {
		t.Errorf("(*Registry)(nil).GetByCode = %+v", p)
	}
}

func TestNewRegistry_skipsEmptyCode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.json")
	err := os.WriteFile(path, []byte(`[{"name": "Valid", "code": "valid"}, {"name": "Empty", "code": ""}]`), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	r, err := NewRegistry(path)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	_, ok := r.GetByCode("valid")
	if !ok {
		t.Error("GetByCode(valid) after empty code entry: ok = false")
	}
}

func TestNewRegistryFromConfig(t *testing.T) {
	t.Run("empty path uses default", func(t *testing.T) {
		cfg := &config.AppConfig{}
		cfg.Permissions.FilePath = ""
		// Default is config/permissions.json; may or may not exist
		_, err := NewRegistryFromConfig(cfg)
		// We only check it doesn't panic and returns either nil error (file exists) or error (file missing)
		_ = err
	})
	t.Run("custom path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "custom.json")
		err := os.WriteFile(path, []byte(`[{"name": "X", "code": "x"}]`), 0644)
		if err != nil {
			t.Fatalf("write temp file: %v", err)
		}
		cfg := &config.AppConfig{}
		cfg.Permissions.FilePath = path
		r, err := NewRegistryFromConfig(cfg)
		if err != nil {
			t.Fatalf("NewRegistryFromConfig: %v", err)
		}
		if r == nil {
			t.Fatal("NewRegistryFromConfig returned nil")
		}
		p, ok := r.GetByCode("x")
		if !ok || p.Name != "X" {
			t.Errorf("GetByCode(x) = %+v, ok = %v", p, ok)
		}
	})
}
