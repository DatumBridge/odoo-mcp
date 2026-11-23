package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
	"github.com/datumbridge/odoo-mcp/internal/odoo"
)

// toolAdder registers one tool's descriptor + handler with the registry.
type toolAdder func(name, desc string, props map[string]interface{}, required []string, h mcp.ToolHandler)

func baseProps(extra map[string]interface{}) map[string]interface{} {
	props := map[string]interface{}{
		"credentials_json": map[string]interface{}{
			"type":        "string",
			"description": "Vault JSON: {base_url, db, username, auth_type, api_key, password, token}",
		},
		"credentials_path": map[string]interface{}{
			"type":        "string",
			"description": "Optional path under ODOO_CREDENTIALS_DIR",
		},
	}
	for k, v := range extra {
		props[k] = v
	}
	return props
}

func schema(props map[string]interface{}, required []string) map[string]interface{} {
	s := map[string]interface{}{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func clientFrom(raw json.RawMessage) (*odoo.Client, map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		m = map[string]interface{}{}
	}
	credsJSON, _ := m["credentials_json"].(string)
	credsPath, _ := m["credentials_path"].(string)
	creds, err := odoo.ParseCredentials(credsJSON, credsPath)
	if err != nil {
		return nil, m, err
	}
	return odoo.NewClient(creds), m, nil
}

func strArg(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func intArg(m map[string]interface{}, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		var i int
		if _, err := fmt.Sscanf(t, "%d", &i); err == nil {
			return i
		}
	}
	return def
}

func floatArg(m map[string]interface{}, key string, def float64) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil {
			return f
		}
	}
	return def
}

func boolArg(m map[string]interface{}, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "true" || s == "1" || s == "yes"
	}
	return def
}

func mapArg(m map[string]interface{}, key string) map[string]interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	obj, _ := v.(map[string]interface{})
	return obj
}

func sliceMapArg(m map[string]interface{}, key string) []map[string]interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var out []map[string]interface{}
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			out = append(out, obj)
		}
	}
	return out
}

func sliceStrArg(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return nil
		}
		var parts []string
		for _, p := range strings.Split(s, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
		return parts
	case []interface{}:
		var parts []string
		for _, item := range val {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				parts = append(parts, s)
			}
		}
		return parts
	case []string:
		return val
	}
	return nil
}

func sliceIntArg(m map[string]interface{}, key string) []int {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	var out []int
	switch val := v.(type) {
	case []interface{}:
		for _, item := range val {
			switch num := item.(type) {
			case float64:
				out = append(out, int(num))
			case int:
				out = append(out, num)
			case string:
				if n, err := strconv.Atoi(strings.TrimSpace(num)); err == nil {
					out = append(out, n)
				}
			}
		}
	case []int:
		return val
	case float64:
		out = append(out, int(val))
	case int:
		out = append(out, val)
	case string:
		for _, p := range strings.Split(val, ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				out = append(out, n)
			}
		}
	}
	return out
}

func domainArg(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return []interface{}{}
	}
	switch val := v.(type) {
	case []interface{}:
		return val
	case string:
		raw := strings.TrimSpace(val)
		if raw == "" || raw == "[]" {
			return []interface{}{}
		}
		var parsed []interface{}
		if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
			return parsed
		}
	}
	return []interface{}{}
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func jsonResult(v interface{}) map[string]interface{} {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText(string(b))
}

func rawResult(b []byte) map[string]interface{} {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, b, "", "  "); err == nil {
		return mcp.ToolResultText(pretty.String())
	}
	return mcp.ToolResultText(string(b))
}

func pagedResult(key string, items interface{}, page, limit, total int) map[string]interface{} {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	totalPage := 0
	if total > 0 {
		totalPage = (total + limit - 1) / limit
	}
	return jsonResult(map[string]interface{}{
		key: items,
		"paging": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_count": total,
			"total_page":  totalPage,
		},
	})
}
