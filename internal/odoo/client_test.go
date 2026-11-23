package odoo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Authenticate_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jsonrpc" {
			http.NotFound(w, r)
			return
		}
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if req.Params.Service == "common" && req.Params.Method == "authenticate" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  2,
			})
			return
		}
		http.Error(w, "unexpected call", 400)
	}))
	defer ts.Close()

	creds := &Credentials{
		BaseURL:  ts.URL,
		DB:       "demo",
		Username: "admin",
		APIKey:   "secret",
	}
	client := NewClient(creds)
	client.apiBase = ts.URL

	uid, err := client.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("authenticate error: %v", err)
	}
	if uid != 2 {
		t.Errorf("got uid %d, want 2", uid)
	}
}

func TestClient_SearchReadAndCount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Params.Service == "common" && req.Params.Method == "authenticate" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  1,
			})
			return
		}
		if req.Params.Service == "object" && req.Params.Method == "execute_kw" {
			method := req.Params.Args[4].(string)
			switch method {
			case "search_read":
				records := []map[string]interface{}{
					{"id": 10, "name": "Alice"},
					{"id": 20, "name": "Bob"},
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result":  records,
				})
				return
			case "search_count":
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result":  2,
				})
				return
			case "create":
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result":  99,
				})
				return
			}
		}
		http.Error(w, "unknown", 400)
	}))
	defer ts.Close()

	creds := &Credentials{
		BaseURL:  ts.URL,
		DB:       "demo",
		Username: "admin",
		APIKey:   "secret",
	}
	client := NewClient(creds)
	client.apiBase = ts.URL

	records, total, err := client.SearchRead(context.Background(), "hr.employee", nil, []string{"id", "name"}, 0, 10, "id asc")
	if err != nil {
		t.Fatalf("search_read failed: %v", err)
	}
	if len(records) != 2 || total != 2 {
		t.Errorf("got len=%d, total=%d, want 2, 2", len(records), total)
	}

	newID, err := client.Create(context.Background(), "res.partner", map[string]interface{}{"name": "Acme Corp"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if newID != 99 {
		t.Errorf("got newID %d, want 99", newID)
	}
}
