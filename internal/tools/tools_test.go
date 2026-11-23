package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	descs, handlers := Register()
	if len(descs) == 0 {
		t.Fatal("expected registered tools, got 0")
	}
	if len(descs) != len(handlers) {
		t.Errorf("mismatch between descs (%d) and handlers (%d)", len(descs), len(handlers))
	}

	expectedTools := []string{
		"odoo_search_employees",
		"odoo_get_employee",
		"odoo_create_employee",
		"odoo_update_employee",
		"odoo_list_departments",
		"odoo_list_leaves",
		"odoo_get_leave",
		"odoo_create_leave_request",
		"odoo_action_leave",
		"odoo_list_leave_types",
		"odoo_get_leave_balances",
		"odoo_list_payslips",
		"odoo_get_payslip",
		"odoo_list_contracts",
		"odoo_get_contract",
		"odoo_create_contract",
		"odoo_update_contract",
		"odoo_search_partners",
		"odoo_get_partner",
		"odoo_create_partner",
		"odoo_update_partner",
		"odoo_list_crm_leads",
		"odoo_get_crm_lead",
		"odoo_create_crm_lead",
		"odoo_update_crm_lead",
		"odoo_create_activity",
		"odoo_list_activities",
		"odoo_search_products",
		"odoo_get_product",
		"odoo_list_quotations",
		"odoo_get_quotation",
		"odoo_create_quotation",
		"odoo_update_quotation",
		"odoo_action_quotation",
		"odoo_execute_kw",
		"odoo_describe_model",
	}

	for _, name := range expectedTools {
		if _, ok := handlers[name]; !ok {
			t.Errorf("expected tool %s to be registered", name)
		}
	}
}

func TestToolHandlers_MockOdoo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
			Params struct {
				Service string        `json:"service"`
				Method  string        `json:"method"`
				Args    []interface{} `json:"args"`
			} `json:"params"`
			ID int64 `json:"id"`
		}
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
			model := req.Params.Args[3].(string)
			method := req.Params.Args[4].(string)

			switch model {
			case "hr.employee":
				if method == "search_read" {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"result":  []map[string]interface{}{{"id": 1, "name": "Test Employee"}},
					})
					return
				}
				if method == "search_count" {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"result":  1,
					})
					return
				}
			case "sale.order":
				if method == "create" {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"result":  42,
					})
					return
				}
			}
		}

		http.Error(w, "not found", 404)
	}))
	defer ts.Close()

	_, handlers := Register()

	credsJSON := `{"base_url":"` + ts.URL + `","db":"testdb","username":"admin","api_key":"testkey"}`

	// Test odoo_search_employees
	args, _ := json.Marshal(map[string]interface{}{
		"credentials_json": credsJSON,
		"name":             "Test",
	})
	res := handlers["odoo_search_employees"](args)
	if res["isError"] == true {
		t.Fatalf("search_employees error: %v", res)
	}

	// Test odoo_create_quotation
	qArgs, _ := json.Marshal(map[string]interface{}{
		"credentials_json": credsJSON,
		"partner_id":       10,
		"lines": []map[string]interface{}{
			{"product_id": 5, "product_uom_qty": 2, "price_unit": 50.0},
		},
	})
	qRes := handlers["odoo_create_quotation"](qArgs)
	if qRes["isError"] == true {
		t.Fatalf("create_quotation error: %v", qRes)
	}
}
