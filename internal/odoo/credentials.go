package odoo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Credentials holds connection parameters for an Odoo instance.
// In production, this matches the DatumBridge vault OdooTokenBundle.
type Credentials struct {
	Type     string `json:"type"`
	BaseURL  string `json:"base_url"`
	DB       string `json:"db"`
	Username string `json:"username"`
	AuthType string `json:"auth_type,omitempty"` // "api_key" or "password"
	APIKey   string `json:"api_key,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

// Credential returns the active secret string (APIKey, Password, or Token).
func (c *Credentials) Credential() string {
	if c == nil {
		return ""
	}
	if c.AuthType == "password" && strings.TrimSpace(c.Password) != "" {
		return strings.TrimSpace(c.Password)
	}
	if strings.TrimSpace(c.APIKey) != "" {
		return strings.TrimSpace(c.APIKey)
	}
	if strings.TrimSpace(c.Password) != "" {
		return strings.TrimSpace(c.Password)
	}
	return strings.TrimSpace(c.Token)
}

// NormalizeBaseURL normalizes raw URL into scheme://host[:port].
func NormalizeBaseURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return ""
	}
	return strings.TrimRight(u.Scheme+"://"+u.Host+u.Path, "/")
}

// ParseCredentials parses credentials from JSON string, path, or environment variables.
func ParseCredentials(credentialsJSON, credentialsPath string) (*Credentials, error) {
	c := &Credentials{}
	raw := strings.TrimSpace(credentialsJSON)

	if raw == "" && strings.TrimSpace(credentialsPath) != "" {
		path := filepath.Clean(credentialsPath)
		if strings.Contains(path, "..") {
			return nil, fmt.Errorf("credentials_path must not contain ..")
		}
		allowedRoot := strings.TrimSpace(os.Getenv("ODOO_CREDENTIALS_DIR"))
		if allowedRoot == "" {
			allowedRoot = "/credentials"
		}
		allowedRoot = filepath.Clean(allowedRoot)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("credentials_path: %w", err)
		}
		absRoot, err := filepath.Abs(allowedRoot)
		if err != nil {
			return nil, fmt.Errorf("ODOO_CREDENTIALS_DIR: %w", err)
		}
		resolvedRoot, err := filepath.EvalSymlinks(absRoot)
		if err != nil {
			resolvedRoot = absRoot
		}
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("credentials_path: %w", err)
		}
		if resolvedPath != resolvedRoot && !strings.HasPrefix(resolvedPath, resolvedRoot+string(os.PathSeparator)) {
			return nil, fmt.Errorf("credentials_path must be under %s", absRoot)
		}
		b, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("read credentials_path: %w", err)
		}
		raw = strings.TrimSpace(string(b))
	}

	if raw != "" {
		if err := json.Unmarshal([]byte(raw), c); err != nil {
			return nil, fmt.Errorf("invalid credentials_json: %w", err)
		}
	}

	// Fallback to environment variables
	if strings.TrimSpace(c.BaseURL) == "" {
		c.BaseURL = os.Getenv("ODOO_BASE_URL")
	}
	c.BaseURL = NormalizeBaseURL(c.BaseURL)

	if strings.TrimSpace(c.DB) == "" {
		c.DB = os.Getenv("ODOO_DB")
	}
	c.DB = strings.TrimSpace(c.DB)

	if strings.TrimSpace(c.Username) == "" {
		c.Username = os.Getenv("ODOO_USERNAME")
	}
	c.Username = strings.TrimSpace(c.Username)

	if c.Credential() == "" {
		if envKey := os.Getenv("ODOO_API_KEY"); envKey != "" {
			c.APIKey = envKey
			c.AuthType = "api_key"
		} else if envPass := os.Getenv("ODOO_PASSWORD"); envPass != "" {
			c.Password = envPass
			c.AuthType = "password"
		}
	}

	if c.BaseURL == "" {
		return nil, fmt.Errorf("missing odoo base_url")
	}
	if c.DB == "" {
		return nil, fmt.Errorf("missing odoo database name")
	}
	if c.Username == "" {
		return nil, fmt.Errorf("missing odoo username/login")
	}
	if c.Credential() == "" {
		return nil, fmt.Errorf("missing odoo api_key or password")
	}

	return c, nil
}
