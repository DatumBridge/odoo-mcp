package tools

import (
	"encoding/json"
	"fmt"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

var defaultPartnerFields = []string{
	"id",
	"name",
	"display_name",
	"email",
	"phone",
	"street",
	"city",
	"zip",
	"country_id",
	"is_company",
	"parent_id",
	"vat",
}

var defaultLeadFields = []string{
	"id",
	"name",
	"partner_id",
	"contact_name",
	"email_from",
	"phone",
	"stage_id",
	"user_id",
	"expected_revenue",
	"probability",
	"priority",
	"type",
	"date_deadline",
}

func registerCRMTools(add toolAdder) {
	// 1. odoo_search_partners
	add(
		"odoo_search_partners",
		"Search customers and business contacts (res.partner) by name, email, phone, or company flag.",
		baseProps(map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Text search matched across name, email, or phone.",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Specific name filter (partial match).",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "Filter by email address.",
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "Filter by phone or mobile.",
			},
			"is_company": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter companies (true) or individual contacts (false).",
			},
			"customer_only": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter only customers with customer_rank > 0.",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to return. Omit for standard contact fields.",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number (1-based, default 1).",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page (default 20, max 100).",
			},
		}),
		nil,
		handleSearchPartners,
	)

	// 2. odoo_get_partner
	add(
		"odoo_get_partner",
		"Get detailed customer/partner profile by partner ID (res.partner).",
		baseProps(map[string]interface{}{
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Partner ID (res.partner).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional fields to fetch.",
			},
		}),
		[]string{"partner_id"},
		handleGetPartner,
	)

	// 3. odoo_create_partner
	add(
		"odoo_create_partner",
		"Create a new contact, customer, or company in Odoo (res.partner).",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Contact or company name.",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "Email address.",
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "Phone number.",
			},
			"mobile": map[string]interface{}{
				"type":        "string",
				"description": "Mobile number.",
			},
			"street": map[string]interface{}{
				"type":        "string",
				"description": "Street address.",
			},
			"city": map[string]interface{}{
				"type":        "string",
				"description": "City.",
			},
			"is_company": map[string]interface{}{
				"type":        "boolean",
				"description": "True if this record is a company/organization.",
			},
			"parent_id": map[string]interface{}{
				"type":        "integer",
				"description": "Parent company partner ID if this is an individual contact.",
			},
			"comment": map[string]interface{}{
				"type":        "string",
				"description": "Internal note about the partner.",
			},
			"extra_fields": map[string]interface{}{
				"type":        "object",
				"description": "Additional fields for res.partner.",
			},
		}),
		[]string{"name"},
		handleCreatePartner,
	)

	// 4. odoo_update_partner
	add(
		"odoo_update_partner",
		"Update customer/contact fields in Odoo (res.partner).",
		baseProps(map[string]interface{}{
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Partner ID to update.",
			},
			"values": map[string]interface{}{
				"type":        "object",
				"description": "Map of field values to update.",
			},
		}),
		[]string{"partner_id", "values"},
		handleUpdatePartner,
	)

	// 5. odoo_list_crm_leads
	add(
		"odoo_list_crm_leads",
		"List CRM leads and opportunities (crm.lead) with stage, salesperson, and revenue filters.",
		baseProps(map[string]interface{}{
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by associated customer/partner ID.",
			},
			"stage_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by CRM stage ID.",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Filter by type: 'lead' or 'opportunity'.",
			},
			"user_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by assigned salesperson user ID.",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number (1-based, default 1).",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page (default 20, max 100).",
			},
		}),
		nil,
		handleListCRMLeads,
	)

	// 6. odoo_get_crm_lead
	add(
		"odoo_get_crm_lead",
		"Get detailed CRM lead or opportunity info by ID (crm.lead).",
		baseProps(map[string]interface{}{
			"lead_id": map[string]interface{}{
				"type":        "integer",
				"description": "Lead/Opportunity ID.",
			},
		}),
		[]string{"lead_id"},
		handleGetCRMLead,
	)

	// 7. odoo_create_crm_lead
	add(
		"odoo_create_crm_lead",
		"Create a new CRM lead or opportunity (crm.lead).",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Opportunity title or requirement summary.",
			},
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Associated customer partner ID (optional).",
			},
			"contact_name": map[string]interface{}{
				"type":        "string",
				"description": "Contact person name.",
			},
			"email_from": map[string]interface{}{
				"type":        "string",
				"description": "Lead contact email.",
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "Lead contact phone.",
			},
			"expected_revenue": map[string]interface{}{
				"type":        "number",
				"description": "Expected revenue amount.",
			},
			"probability": map[string]interface{}{
				"type":        "number",
				"description": "Win probability percentage (0-100).",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Priority ('0': Low, '1': Medium, '2': High, '3': Very High).",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Internal notes or description.",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "'lead' or 'opportunity' (default 'opportunity').",
			},
		}),
		[]string{"name"},
		handleCreateCRMLead,
	)

	// 8. odoo_update_crm_lead
	add(
		"odoo_update_crm_lead",
		"Update a CRM lead or opportunity (stage, probability, revenue, notes).",
		baseProps(map[string]interface{}{
			"lead_id": map[string]interface{}{
				"type":        "integer",
				"description": "Lead ID.",
			},
			"values": map[string]interface{}{
				"type":        "object",
				"description": "Map of field values to update (e.g. stage_id, expected_revenue, probability).",
			},
		}),
		[]string{"lead_id", "values"},
		handleUpdateCRMLead,
	)

	// 9. odoo_create_activity
	add(
		"odoo_create_activity",
		"Schedule or log a customer care activity (call, email, meeting, todo) on a record (mail.activity).",
		baseProps(map[string]interface{}{
			"res_model": map[string]interface{}{
				"type":        "string",
				"description": "Target model: e.g. 'res.partner', 'crm.lead', 'sale.order'.",
			},
			"res_id": map[string]interface{}{
				"type":        "integer",
				"description": "Target record ID.",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Short summary of activity.",
			},
			"note": map[string]interface{}{
				"type":        "string",
				"description": "Details or notes of activity.",
			},
			"activity_type_id": map[string]interface{}{
				"type":        "integer",
				"description": "Activity type ID (e.g. Call=2, Email=1, Todo=4).",
			},
			"date_deadline": map[string]interface{}{
				"type":        "string",
				"description": "Deadline date (YYYY-MM-DD).",
			},
			"user_id": map[string]interface{}{
				"type":        "integer",
				"description": "Assignee user ID (optional).",
			},
		}),
		[]string{"res_model", "res_id", "summary"},
		handleCreateActivity,
	)

	// 10. odoo_list_activities
	add(
		"odoo_list_activities",
		"List scheduled customer activities (mail.activity) on records or assigned to users.",
		baseProps(map[string]interface{}{
			"res_model": map[string]interface{}{
				"type":        "string",
				"description": "Filter by target model (e.g. 'res.partner', 'crm.lead').",
			},
			"res_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by specific record ID.",
			},
			"user_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by assigned user ID.",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number (1-based, default 1).",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page (default 20, max 100).",
			},
		}),
		nil,
		handleListActivities,
	)
}

func handleSearchPartners(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if q := strArg(m, "query"); q != "" {
		domain = append(domain, "|", "|",
			[]interface{}{"name", "ilike", q},
			[]interface{}{"email", "ilike", q},
			[]interface{}{"phone", "ilike", q},
		)
	}
	if name := strArg(m, "name"); name != "" {
		domain = append(domain, []interface{}{"name", "ilike", name})
	}
	if email := strArg(m, "email"); email != "" {
		domain = append(domain, []interface{}{"email", "ilike", email})
	}
	if phone := strArg(m, "phone"); phone != "" {
		domain = append(domain, []interface{}{"phone", "ilike", phone})
	}
	if isCompany, ok := m["is_company"].(bool); ok {
		domain = append(domain, []interface{}{"is_company", "=", isCompany})
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultPartnerFields
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "res.partner", domain, fields, offset, limit, "name asc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("partners", records, page, limit, total)
}

func handleGetPartner(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	partnerID := intArg(m, "partner_id", 0)
	if partnerID <= 0 {
		return mcp.ToolResultError("partner_id must be a positive integer")
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = append(defaultPartnerFields,
			"website", "comment", "child_ids", "category_id",
		)
	}

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "res.partner", []int{partnerID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("partner %d not found", partnerID))
	}

	return jsonResult(map[string]interface{}{"partner": records[0]})
}

func handleCreatePartner(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	name := strArg(m, "name")
	if name == "" {
		return mcp.ToolResultError("name is required to create a partner")
	}

	vals := map[string]interface{}{"name": name}
	if email := strArg(m, "email"); email != "" {
		vals["email"] = email
	}
	if phone := strArg(m, "phone"); phone != "" {
		vals["phone"] = phone
	}
	if mobile := strArg(m, "mobile"); mobile != "" {
		vals["mobile"] = mobile
	}
	if street := strArg(m, "street"); street != "" {
		vals["street"] = street
	}
	if city := strArg(m, "city"); city != "" {
		vals["city"] = city
	}
	if isComp, ok := m["is_company"].(bool); ok {
		vals["is_company"] = isComp
	}
	if parentID := intArg(m, "parent_id", 0); parentID > 0 {
		vals["parent_id"] = parentID
	}
	if comment := strArg(m, "comment"); comment != "" {
		vals["comment"] = comment
	}
	if extra := mapArg(m, "extra_fields"); len(extra) > 0 {
		for k, v := range extra {
			vals[k] = v
		}
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "res.partner", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"partner_id": newID,
		"name":       name,
		"success":    true,
	})
}

func handleUpdatePartner(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	partnerID := intArg(m, "partner_id", 0)
	if partnerID <= 0 {
		return mcp.ToolResultError("partner_id must be a positive integer")
	}

	values := mapArg(m, "values")
	if len(values) == 0 {
		return mcp.ToolResultError("values map cannot be empty")
	}

	c, cancel := ctx()
	defer cancel()

	ok, err := client.Write(c, "res.partner", []int{partnerID}, values)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"partner_id": partnerID,
		"success":    ok,
	})
}

func handleListCRMLeads(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if partnerID := intArg(m, "partner_id", 0); partnerID > 0 {
		domain = append(domain, []interface{}{"partner_id", "=", partnerID})
	}
	if stageID := intArg(m, "stage_id", 0); stageID > 0 {
		domain = append(domain, []interface{}{"stage_id", "=", stageID})
	}
	if t := strArg(m, "type"); t != "" {
		domain = append(domain, []interface{}{"type", "=", t})
	}
	if userID := intArg(m, "user_id", 0); userID > 0 {
		domain = append(domain, []interface{}{"user_id", "=", userID})
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "crm.lead", domain, defaultLeadFields, offset, limit, "id desc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("leads", records, page, limit, total)
}

func handleGetCRMLead(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	leadID := intArg(m, "lead_id", 0)
	if leadID <= 0 {
		return mcp.ToolResultError("lead_id must be a positive integer")
	}

	fields := append(defaultLeadFields, "description", "referred", "tag_ids", "lost_reason_id")

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "crm.lead", []int{leadID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("lead %d not found", leadID))
	}

	return jsonResult(map[string]interface{}{"lead": records[0]})
}

func handleCreateCRMLead(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	name := strArg(m, "name")
	if name == "" {
		return mcp.ToolResultError("name is required to create a CRM lead")
	}

	leadType := strArg(m, "type")
	if leadType == "" {
		leadType = "opportunity"
	}

	vals := map[string]interface{}{
		"name": name,
		"type": leadType,
	}
	if partnerID := intArg(m, "partner_id", 0); partnerID > 0 {
		vals["partner_id"] = partnerID
	}
	if contactName := strArg(m, "contact_name"); contactName != "" {
		vals["contact_name"] = contactName
	}
	if email := strArg(m, "email_from"); email != "" {
		vals["email_from"] = email
	}
	if phone := strArg(m, "phone"); phone != "" {
		vals["phone"] = phone
	}
	if rev := floatArg(m, "expected_revenue", 0); rev > 0 {
		vals["expected_revenue"] = rev
	}
	if prob := floatArg(m, "probability", -1); prob >= 0 {
		vals["probability"] = prob
	}
	if prio := strArg(m, "priority"); prio != "" {
		vals["priority"] = prio
	}
	if desc := strArg(m, "description"); desc != "" {
		vals["description"] = desc
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "crm.lead", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"lead_id": newID,
		"name":    name,
		"success": true,
	})
}

func handleUpdateCRMLead(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	leadID := intArg(m, "lead_id", 0)
	if leadID <= 0 {
		return mcp.ToolResultError("lead_id must be a positive integer")
	}

	values := mapArg(m, "values")
	if len(values) == 0 {
		return mcp.ToolResultError("values map cannot be empty")
	}

	c, cancel := ctx()
	defer cancel()

	ok, err := client.Write(c, "crm.lead", []int{leadID}, values)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"lead_id": leadID,
		"success": ok,
	})
}

func handleCreateActivity(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	resModel := strArg(m, "res_model")
	resID := intArg(m, "res_id", 0)
	summary := strArg(m, "summary")
	if resModel == "" || resID <= 0 || summary == "" {
		return mcp.ToolResultError("res_model, res_id, and summary are required")
	}

	vals := map[string]interface{}{
		"res_model": resModel,
		"res_id":    resID,
		"summary":   summary,
	}
	if note := strArg(m, "note"); note != "" {
		vals["note"] = note
	}
	if actTypeID := intArg(m, "activity_type_id", 0); actTypeID > 0 {
		vals["activity_type_id"] = actTypeID
	}
	if deadline := strArg(m, "date_deadline"); deadline != "" {
		vals["date_deadline"] = deadline
	}
	if userID := intArg(m, "user_id", 0); userID > 0 {
		vals["user_id"] = userID
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "mail.activity", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"activity_id": newID,
		"res_model":   resModel,
		"res_id":      resID,
		"success":     true,
	})
}

func handleListActivities(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if model := strArg(m, "res_model"); model != "" {
		domain = append(domain, []interface{}{"res_model", "=", model})
	}
	if resID := intArg(m, "res_id", 0); resID > 0 {
		domain = append(domain, []interface{}{"res_id", "=", resID})
	}
	if userID := intArg(m, "user_id", 0); userID > 0 {
		domain = append(domain, []interface{}{"user_id", "=", userID})
	}

	fields := []string{
		"id", "summary", "activity_type_id", "res_model", "res_id", "res_name",
		"date_deadline", "user_id", "state", "note",
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "mail.activity", domain, fields, offset, limit, "date_deadline asc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("activities", records, page, limit, total)
}
