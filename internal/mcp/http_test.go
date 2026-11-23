package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStreamableHTTP_Lifecycle(t *testing.T) {
	store := NewSessionStore()
	tools := []ToolDesc{
		{Name: "test_tool", Description: "Test tool description"},
	}
	handlers := map[string]ToolHandler{
		"test_tool": func(args json.RawMessage) map[string]interface{} {
			return ToolResultText("hello from test_tool")
		},
	}
	srv := &Server{
		Sessions: store,
		Tools:    tools,
		Handlers: handlers,
	}

	// 1. Initialize
	initReq := `{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}`
	rec := httptest.NewRecorder()
	srv.HandleStreamableHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(initReq)))

	if rec.Code != http.StatusOK {
		t.Fatalf("initialize status: %d", rec.Code)
	}
	sid := rec.Header().Get(SessionHeader)
	if sid == "" {
		t.Fatal("missing session header in initialize response")
	}

	// 2. tools/list without session -> 400 or rpc error
	rec = httptest.NewRecorder()
	listReq := `{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}`
	srv.HandleStreamableHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(listReq)))
	var errResp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp["error"] == nil {
		t.Errorf("expected session error, got %v", rec.Body.String())
	}

	// 3. tools/list with session
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(listReq))
	req.Header.Set(SessionHeader, sid)
	rec = httptest.NewRecorder()
	srv.HandleStreamableHTTP(rec, req)
	var listResp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	res, ok := listResp["result"].(map[string]interface{})
	if !ok || res["tools"] == nil {
		t.Fatalf("expected tools in list, got %v", listResp)
	}

	// 4. tools/call with session
	callReq := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"test_tool","arguments":{}},"id":3}`
	req = httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(callReq))
	req.Header.Set(SessionHeader, sid)
	rec = httptest.NewRecorder()
	srv.HandleStreamableHTTP(rec, req)
	var callResp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &callResp)
	callRes, ok := callResp["result"].(map[string]interface{})
	if !ok || callRes["content"] == nil {
		t.Fatalf("expected tool content, got %v", callResp)
	}
}
