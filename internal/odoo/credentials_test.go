package odoo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCredentials_DirectJSON(t *testing.T) {
	raw := `{"base_url":"https://test.odoo.com","db":"testdb","username":"admin","api_key":"secret123"}`
	creds, err := ParseCredentials(raw, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.BaseURL != "https://test.odoo.com" {
		t.Errorf("got %q, want https://test.odoo.com", creds.BaseURL)
	}
	if creds.DB != "testdb" {
		t.Errorf("got %q, want testdb", creds.DB)
	}
	if creds.Username != "admin" {
		t.Errorf("got %q, want admin", creds.Username)
	}
	if creds.Credential() != "secret123" {
		t.Errorf("got %q, want secret123", creds.Credential())
	}
}

func TestParseCredentials_PasswordAuth(t *testing.T) {
	raw := `{"base_url":"http://localhost:8069/","db":"demo","username":"demo","auth_type":"password","password":"pwd"}`
	creds, err := ParseCredentials(raw, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.BaseURL != "http://localhost:8069" {
		t.Errorf("got %q, want http://localhost:8069", creds.BaseURL)
	}
	if creds.Credential() != "pwd" {
		t.Errorf("got %q, want pwd", creds.Credential())
	}
}

func TestParseCredentials_EnvFallback(t *testing.T) {
	_ = os.Setenv("ODOO_BASE_URL", "https://env.odoo.com")
	_ = os.Setenv("ODOO_DB", "envdb")
	_ = os.Setenv("ODOO_USERNAME", "envuser")
	_ = os.Setenv("ODOO_API_KEY", "envkey")
	defer func() {
		_ = os.Unsetenv("ODOO_BASE_URL")
		_ = os.Unsetenv("ODOO_DB")
		_ = os.Unsetenv("ODOO_USERNAME")
		_ = os.Unsetenv("ODOO_API_KEY")
	}()

	creds, err := ParseCredentials("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.BaseURL != "https://env.odoo.com" || creds.DB != "envdb" || creds.Username != "envuser" || creds.Credential() != "envkey" {
		t.Errorf("unexpected creds: %+v", creds)
	}
}

func TestParseCredentials_PathJail(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("ODOO_CREDENTIALS_DIR", tmpDir)
	defer os.Unsetenv("ODOO_CREDENTIALS_DIR")

	filePath := filepath.Join(tmpDir, "creds.json")
	_ = os.WriteFile(filePath, []byte(`{"base_url":"https://file.odoo.com","db":"fdb","username":"fuser","api_key":"fkey"}`), 0600)

	creds, err := ParseCredentials("", filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.BaseURL != "https://file.odoo.com" || creds.Credential() != "fkey" {
		t.Errorf("unexpected creds: %+v", creds)
	}

	// Test traversal outside jail
	_, err = ParseCredentials("", "../outside.json")
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	}
}
