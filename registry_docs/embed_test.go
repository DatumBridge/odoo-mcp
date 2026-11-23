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

func TestMarkdown_AllToolsExistAndConformToVibePrinciples(t *testing.T) {
	if len(allOdooTools) != 34 {
		t.Fatalf("expected 34 tools, got %d", len(allOdooTools))
	}

	for _, name := range allOdooTools {
		doc := Markdown(name)
		if strings.TrimSpace(doc) == "" {
			t.Fatalf("missing or empty documentation file for tool: %s", name)
		}
		if len(doc) > 24000 {
			t.Errorf("doc for %s exceeds 24000 character limit: %d", name, len(doc))
		}
		if !strings.HasPrefix(doc, "# "+name) {
			t.Errorf("doc for %s should start with '# %s'", name, name)
		}
		if !strings.Contains(doc, "## Audience") {
			t.Errorf("doc for %s missing ## Audience section", name)
		}
		if !strings.Contains(doc, "| internal | required |") && !strings.Contains(doc, "| external | required |") {
			t.Errorf("doc for %s missing Audience table row", name)
		}
		if !strings.Contains(doc, "## Agent rules") {
			t.Errorf("doc for %s missing ## Agent rules section", name)
		}
		if !strings.Contains(doc, "Do not invent a token") {
			t.Errorf("doc for %s missing 'Do not invent a token' rule", name)
		}
		if !strings.Contains(doc, "untrusted") {
			t.Errorf("doc for %s missing untrusted payload warning", name)
		}
		if !strings.Contains(doc, "## When to call") {
			t.Errorf("doc for %s missing ## When to call section", name)
		}
		if !strings.Contains(doc, "## Parameters") {
			t.Errorf("doc for %s missing ## Parameters section", name)
		}
		if !strings.Contains(doc, "| Name | Required | Type | Sample | Related tool | Meaning |") {
			t.Errorf("doc for %s missing standardized Parameters table headers", name)
		}
		if !strings.Contains(doc, "## Success fields") {
			t.Errorf("doc for %s missing ## Success fields section", name)
		}
		if !strings.Contains(doc, "## Error codes") {
			t.Errorf("doc for %s missing ## Error codes section", name)
		}
		if !strings.Contains(doc, "## Cases") {
			t.Errorf("doc for %s missing ## Cases section", name)
		}
		if !strings.Contains(doc, "### Typical call") {
			t.Errorf("doc for %s missing ### Typical call case", name)
		}

		// Ensure typical call input does NOT contain credentials_json or tokens
		parts := strings.Split(doc, "### Typical call")
		if len(parts) > 1 {
			typicalPart := parts[1]
			if strings.Contains(typicalPart, `"credentials_json"`) {
				t.Errorf("typical call case in %s must not contain credentials_json", name)
			}
			if strings.Contains(typicalPart, `"api_key"`) {
				t.Errorf("typical call case in %s must not contain api_key", name)
			}
		}
	}
}
