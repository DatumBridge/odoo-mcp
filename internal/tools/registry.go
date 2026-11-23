package tools

import (
	"strings"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
	registrydocs "github.com/datumbridge/odoo-mcp/registry_docs"
)

var toolOutputs = map[string]string{
	// Human Resources (5)
	"odoo_search_employees": "employees, paging",
	"odoo_get_employee":     "employee",
	"odoo_create_employee":  "employee_id, name, success",
	"odoo_update_employee":  "employee_id, success",
	"odoo_list_departments": "departments, paging",

	// Time Off / Leaves (6)
	"odoo_list_leaves":          "leaves, paging",
	"odoo_get_leave":            "leave",
	"odoo_create_leave_request": "leave_id, success",
	"odoo_action_leave":         "leave_id, action, success",
	"odoo_list_leave_types":     "leave_types, total_count",
	"odoo_get_leave_balances":   "allocations, total_count",

	// Payroll & Contracts (4)
	"odoo_list_payslips": "payslips, paging",
	"odoo_get_payslip":   "payslip",
	"odoo_list_contracts": "contracts, paging",
	"odoo_get_contract":  "contract",

	// CRM & Partners (10)
	"odoo_search_partners": "partners, paging",
	"odoo_get_partner":     "partner",
	"odoo_create_partner":  "partner_id, name, success",
	"odoo_update_partner":  "partner_id, success",
	"odoo_list_crm_leads":  "leads, paging",
	"odoo_get_crm_lead":    "lead",
	"odoo_create_crm_lead": "lead_id, name, success",
	"odoo_update_crm_lead": "lead_id, success",
	"odoo_create_activity": "activity_id, success",
	"odoo_list_activities": "activities, paging",

	// Sales & Products (7)
	"odoo_search_products":   "products, paging",
	"odoo_get_product":       "product",
	"odoo_list_quotations":   "quotations, paging",
	"odoo_get_quotation":     "quotation",
	"odoo_create_quotation":  "order_id, success",
	"odoo_update_quotation":  "order_id, success",
	"odoo_action_quotation":  "order_id, action, success",

	// Generic & Discovery (2)
	"odoo_execute_kw":      "result, model, method",
	"odoo_describe_model":  "fields, model",
}

type relatedMeta struct {
	tool  string
	field string
}

var paramRelated = map[string]relatedMeta{
	"employee_id":   {tool: "odoo_search_employees", field: "employees[0].id"},
	"department_id": {tool: "odoo_list_departments", field: "departments[0].id"},
	"partner_id":    {tool: "odoo_search_partners", field: "partners[0].id"},
	"product_id":    {tool: "odoo_search_products", field: "products[0].id"},
	"order_id":      {tool: "odoo_list_quotations", field: "quotations[0].id"},
	"lead_id":       {tool: "odoo_list_crm_leads", field: "leads[0].id"},
	"leave_id":      {tool: "odoo_list_leaves", field: "leaves[0].id"},
	"contract_id":   {tool: "odoo_list_contracts", field: "contracts[0].id"},
	"payslip_id":    {tool: "odoo_list_payslips", field: "payslips[0].id"},
}

func annotateRelated(props map[string]interface{}) {
	for k, v := range props {
		propMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if rel, found := paramRelated[k]; found {
			propMap["x-datumbridge-related-tool"] = rel.tool
			propMap["x-datumbridge-output-field"] = rel.field
		}
	}
}

func formatToolDescription(name, desc string) string {
	res := strings.TrimSpace(desc)
	if !strings.Contains(strings.ToLower(res), "capabilities:") {
		res += "\n\nCapabilities: odoo." + name
	}
	if !strings.Contains(strings.ToLower(res), "outputs:") {
		outs := toolOutputs[name]
		if outs != "" {
			res += "\nOutputs: " + outs + ", success"
		} else {
			res += "\nOutputs: success"
		}
	}
	return res
}

// Register returns the full odoo-mcp tool catalogue.
func Register() ([]mcp.ToolDesc, map[string]mcp.ToolHandler) {
	cfg := LoadServerConfig()
	handlers := map[string]mcp.ToolHandler{}
	var descs []mcp.ToolDesc

	add := func(name, desc string, props map[string]interface{}, required []string, h mcp.ToolHandler) {
		if !cfg.toolEnabled(name) {
			return
		}
		annotateRelated(props)
		input := schema(props, required)
		input["x-datumbridge-capabilities"] = []string{"odoo." + name}
		if doc := registrydocs.Markdown(name); doc != "" {
			input["x-datumbridge-docs"] = doc
		}
		descs = append(descs, mcp.ToolDesc{
			Name:        name,
			Description: formatToolDescription(name, desc),
			InputSchema: input,
		})
		handlers[name] = h
	}

	registerHRTools(add)
	registerLeaveTools(add)
	registerPayrollTools(add)
	registerCRMTools(add)
	registerSalesTools(add)
	registerGenericTools(add)

	return descs, handlers
}
