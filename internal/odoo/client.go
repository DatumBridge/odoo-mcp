package odoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client calls the Odoo JSON-RPC API.
type Client struct {
	creds      *Credentials
	httpClient *http.Client
	uid        int
	mu         sync.Mutex
	apiBase    string // optional override for testing
}

// NewClient creates a new Odoo JSON-RPC client.
func NewClient(creds *Credentials) *Client {
	timeoutSec := 30
	if v := strings.TrimSpace(os.Getenv("ODOO_HTTP_TIMEOUT_SEC")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSec = n
		}
	}
	base := strings.TrimSpace(os.Getenv("ODOO_API_BASE"))
	return &Client{
		creds:      creds,
		httpClient: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
		apiBase:    strings.TrimRight(base, "/"),
	}
}

type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  jsonRPCParams `json:"params"`
	ID      int64         `json:"id"`
}

type jsonRPCParams struct {
	Service string        `json:"service"`
	Method  string        `json:"method"`
	Args    []interface{} `json:"args"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *jsonRPCError) Error() string {
	if e == nil {
		return ""
	}
	if e.Data != nil {
		if dataMap, ok := e.Data.(map[string]interface{}); ok {
			if msg, ok := dataMap["message"].(string); ok && strings.TrimSpace(msg) != "" {
				return fmt.Sprintf("odoo error (%d): %s: %s", e.Code, e.Message, msg)
			}
		}
		dataBytes, _ := json.Marshal(e.Data)
		return fmt.Sprintf("odoo error (%d): %s - %s", e.Code, e.Message, string(dataBytes))
	}
	return fmt.Sprintf("odoo error (%d): %s", e.Code, e.Message)
}

func (c *Client) endpoint() string {
	if c.apiBase != "" {
		return c.apiBase + "/jsonrpc"
	}
	return c.creds.BaseURL + "/jsonrpc"
}

func (c *Client) callRPC(ctx context.Context, service, method string, args []interface{}) (json.RawMessage, error) {
	reqBody := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "call",
		Params: jsonRPCParams{
			Service: service,
			Method:  method,
			Args:    args,
		},
		ID: time.Now().UnixNano(),
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal jsonrpc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("odoo jsonrpc request failed: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, 10<<20) // 10 MiB cap
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("odoo http error %d: %s", resp.StatusCode, string(data))
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal odoo response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

// Authenticate verifies credentials and caches the user's UID.
func (c *Client) Authenticate(ctx context.Context) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.uid > 0 {
		return c.uid, nil
	}

	if c.creds == nil {
		return 0, fmt.Errorf("odoo client credentials not configured")
	}

	args := []interface{}{
		c.creds.DB,
		c.creds.Username,
		c.creds.Credential(),
		map[string]interface{}{},
	}

	rawRes, err := c.callRPC(ctx, "common", "authenticate", args)
	if err != nil {
		return 0, fmt.Errorf("odoo authenticate failed: %w", err)
	}

	// Result can be integer UID or boolean false
	var boolVal bool
	if err := json.Unmarshal(rawRes, &boolVal); err == nil && !boolVal {
		return 0, fmt.Errorf("odoo authenticate failed: invalid credentials or database")
	}

	var uid int
	if err := json.Unmarshal(rawRes, &uid); err != nil || uid <= 0 {
		return 0, fmt.Errorf("odoo authenticate failed: unexpected response %s", string(rawRes))
	}

	c.uid = uid
	return uid, nil
}

// ExecuteKW executes an arbitrary method on an Odoo model via object.execute_kw.
func (c *Client) ExecuteKW(ctx context.Context, model, method string, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	uid, err := c.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if args == nil {
		args = []interface{}{}
	}
	if kwargs == nil {
		kwargs = map[string]interface{}{}
	}

	rpcArgs := []interface{}{
		c.creds.DB,
		uid,
		c.creds.Credential(),
		model,
		method,
		args,
		kwargs,
	}

	rawRes, err := c.callRPC(ctx, "object", "execute_kw", rpcArgs)
	if err != nil {
		return nil, err
	}

	var res interface{}
	if len(rawRes) > 0 && string(rawRes) != "null" {
		if err := json.Unmarshal(rawRes, &res); err != nil {
			return nil, fmt.Errorf("decode execute_kw result: %w", err)
		}
	}
	return res, nil
}

// SearchRead searches records in an Odoo model and reads specified fields, along with total record count.
func (c *Client) SearchRead(ctx context.Context, model string, domain []interface{}, fields []string, offset, limit int, order string) ([]map[string]interface{}, int, error) {
	if domain == nil {
		domain = []interface{}{}
	}

	kwargs := map[string]interface{}{}
	if len(fields) > 0 {
		kwargs["fields"] = fields
	}
	if offset > 0 {
		kwargs["offset"] = offset
	}
	if limit > 0 {
		kwargs["limit"] = limit
	}
	if order != "" {
		kwargs["order"] = order
	}

	res, err := c.ExecuteKW(ctx, model, "search_read", []interface{}{domain}, kwargs)
	if err != nil {
		return nil, 0, err
	}

	var records []map[string]interface{}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, 0, fmt.Errorf("encode search_read result: %w", err)
	}
	if err := json.Unmarshal(b, &records); err != nil {
		return nil, 0, fmt.Errorf("decode records list: %w", err)
	}

	// Fetch count
	countRes, err := c.ExecuteKW(ctx, model, "search_count", []interface{}{domain}, map[string]interface{}{})
	total := len(records)
	if err == nil && countRes != nil {
		switch v := countRes.(type) {
		case float64:
			total = int(v)
		case int:
			total = v
		}
	}

	return records, total, nil
}

// Read reads specific fields for a set of record IDs.
func (c *Client) Read(ctx context.Context, model string, ids []int, fields []string) ([]map[string]interface{}, error) {
	kwargs := map[string]interface{}{}
	if len(fields) > 0 {
		kwargs["fields"] = fields
	}

	res, err := c.ExecuteKW(ctx, model, "read", []interface{}{ids}, kwargs)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("encode read result: %w", err)
	}
	if err := json.Unmarshal(b, &records); err != nil {
		return nil, fmt.Errorf("decode records: %w", err)
	}
	return records, nil
}

// Create inserts a new record into an Odoo model and returns its new ID.
func (c *Client) Create(ctx context.Context, model string, values map[string]interface{}) (int, error) {
	res, err := c.ExecuteKW(ctx, model, "create", []interface{}{values}, map[string]interface{}{})
	if err != nil {
		return 0, err
	}

	switch v := res.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected create response id type: %T (%v)", res, res)
	}
}

// Write updates existing records with specified values.
func (c *Client) Write(ctx context.Context, model string, ids []int, values map[string]interface{}) (bool, error) {
	res, err := c.ExecuteKW(ctx, model, "write", []interface{}{ids, values}, map[string]interface{}{})
	if err != nil {
		return false, err
	}
	if b, ok := res.(bool); ok {
		return b, nil
	}
	return true, nil
}

// Unlink deletes records by IDs.
func (c *Client) Unlink(ctx context.Context, model string, ids []int) (bool, error) {
	res, err := c.ExecuteKW(ctx, model, "unlink", []interface{}{ids}, map[string]interface{}{})
	if err != nil {
		return false, err
	}
	if b, ok := res.(bool); ok {
		return b, nil
	}
	return true, nil
}

// CallMethod calls an action/button method on records, e.g. action_confirm, action_approve.
func (c *Client) CallMethod(ctx context.Context, model, method string, ids []int, kwargs map[string]interface{}) (interface{}, error) {
	if kwargs == nil {
		kwargs = map[string]interface{}{}
	}
	return c.ExecuteKW(ctx, model, method, []interface{}{ids}, kwargs)
}

// FieldsGet inspects the schema of a model.
func (c *Client) FieldsGet(ctx context.Context, model string, allfields []string, attributes []string) (map[string]interface{}, error) {
	kwargs := map[string]interface{}{}
	if len(allfields) > 0 {
		kwargs["allfields"] = allfields
	}
	if len(attributes) > 0 {
		kwargs["attributes"] = attributes
	}

	res, err := c.ExecuteKW(ctx, model, "fields_get", []interface{}{}, kwargs)
	if err != nil {
		return nil, err
	}

	m, ok := res.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected fields_get return type: %T", res)
	}
	return m, nil
}
