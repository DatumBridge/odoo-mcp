package registrydocs

import (
	"strings"
	"testing"
)

var allOdooTools = []string{
	// Human Resources (5)
	"odoo_search_employees", "odoo_get_employee", "odoo_create_employee",
	"odoo_update_employee", "odoo_list_departments",

	// Time Off / Leaves (6)
	"odoo_list_leaves", "odoo_get_leave", "odoo_create_leave_request",
	"odoo_action_leave", "odoo_list_leave_types", "odoo_get_leave_balances",

	// Payroll & Contracts (4)
	"odoo_list_payslips", "odoo_get_payslip", "odoo_list_contracts", "odoo_get_contract",

	// CRM & Partners (10)
	"odoo_search_partners", "odoo_get_partner", "odoo_create_partner",
	"odoo_update_partner", "odoo_list_crm_leads", "odoo_get_crm_lead",
	"odoo_create_crm_lead", "odoo_update_crm_lead", "odoo_create_activity",
	"odoo_list_activities",

	// Sales & Products (7)
	"odoo_search_products", "odoo_get_product", "odoo_list_quotations",
	"odoo_get_quotation", "odoo_create_quotation", "odoo_update_quotation",
	"odoo_action_quotation",

	// Generic & Discovery (2)
	"odoo_execute_kw", "odoo_describe_model",
}

func TestMarkdown_AllToolsExist(t *testing.T) {
	if len(allOdooTools) != 34 {
		t.Fatalf("expected 34 tools, got %d", len(allOdooTools))
	}

	for _, name := range allOdooTools {
		doc := Markdown(name)
		if strings.TrimSpace(doc) == "" {
			t.Errorf("missing or empty documentation file for tool: %s", name)
		}
		if !strings.HasPrefix(doc, "# "+name) {
			t.Errorf("doc for %s should start with '# %s'", name, name)
		}
		if !strings.Contains(doc, "| Audience | internal |") {
			t.Errorf("doc for %s missing Audience row", name)
		}
	}
}
