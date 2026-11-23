package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

var defaultLeaveFields = []string{
	"id",
	"name",
	"employee_id",
	"holiday_status_id",
	"date_from",
	"date_to",
	"request_date_from",
	"request_date_to",
	"number_of_days",
	"state",
}

func registerLeaveTools(add toolAdder) {
	// 1. odoo_list_leaves
	add(
		"odoo_list_leaves",
		"List time off / leave requests (hr.leave) with filters by employee, state, or dates.",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by employee ID.",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Filter by status: 'draft', 'confirm' (submitted), 'validate' (approved), 'refuse'.",
			},
			"holiday_status_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by leave type ID (hr.leave.type).",
			},
			"date_from": map[string]interface{}{
				"type":        "string",
				"description": "Filter leaves starting on or after this date (YYYY-MM-DD).",
			},
			"date_to": map[string]interface{}{
				"type":        "string",
				"description": "Filter leaves ending on or before this date (YYYY-MM-DD).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to return. Omit for standard leave fields.",
			},
			"order": map[string]interface{}{
				"type":        "string",
				"description": "Sort order, default 'date_from desc'.",
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
		handleListLeaves,
	)

	// 2. odoo_get_leave
	add(
		"odoo_get_leave",
		"Get detailed information of a specific leave request (hr.leave) by ID.",
		baseProps(map[string]interface{}{
			"leave_id": map[string]interface{}{
				"type":        "integer",
				"description": "Leave request ID.",
			},
		}),
		[]string{"leave_id"},
		handleGetLeave,
	)

	// 3. odoo_create_leave_request
	add(
		"odoo_create_leave_request",
		"Create a new time off / leave request (hr.leave) in Odoo.",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Employee ID requesting leave.",
			},
			"holiday_status_id": map[string]interface{}{
				"type":        "integer",
				"description": "Leave type ID (from odoo_list_leave_types).",
			},
			"request_date_from": map[string]interface{}{
				"type":        "string",
				"description": "Start date (YYYY-MM-DD).",
			},
			"request_date_to": map[string]interface{}{
				"type":        "string",
				"description": "End date (YYYY-MM-DD).",
			},
			"number_of_days": map[string]interface{}{
				"type":        "number",
				"description": "Duration in days.",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Reason or description for the leave request.",
			},
		}),
		[]string{"employee_id", "holiday_status_id", "request_date_from", "request_date_to"},
		handleCreateLeaveRequest,
	)

	// 4. odoo_action_leave
	add(
		"odoo_action_leave",
		"Perform workflow action on a leave request: 'approve', 'validate', or 'refuse'.",
		baseProps(map[string]interface{}{
			"leave_id": map[string]interface{}{
				"type":        "integer",
				"description": "Leave request ID.",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: 'approve', 'validate', or 'refuse'.",
			},
		}),
		[]string{"leave_id", "action"},
		handleActionLeave,
	)

	// 5. odoo_list_leave_types
	add(
		"odoo_list_leave_types",
		"List all available leave types / time off categories (hr.leave.type).",
		baseProps(map[string]interface{}{
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter active leave types (default true).",
			},
		}),
		nil,
		handleListLeaveTypes,
	)

	// 6. odoo_get_leave_balances
	add(
		"odoo_get_leave_balances",
		"Get remaining leave balances and allocations (hr.leave.allocation) for an employee.",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Employee ID.",
			},
		}),
		[]string{"employee_id"},
		handleGetLeaveBalances,
	)
}

func handleListLeaves(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if empID := intArg(m, "employee_id", 0); empID > 0 {
		domain = append(domain, []interface{}{"employee_id", "=", empID})
	}
	if state := strArg(m, "state"); state != "" {
		domain = append(domain, []interface{}{"state", "=", state})
	}
	if statusID := intArg(m, "holiday_status_id", 0); statusID > 0 {
		domain = append(domain, []interface{}{"holiday_status_id", "=", statusID})
	}
	if dFrom := strArg(m, "date_from"); dFrom != "" {
		domain = append(domain, []interface{}{"date_from", ">=", dFrom})
	}
	if dTo := strArg(m, "date_to"); dTo != "" {
		domain = append(domain, []interface{}{"date_to", "<=", dTo})
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultLeaveFields
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	order := strArg(m, "order")
	if order == "" {
		order = "date_from desc"
	}

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "hr.leave", domain, fields, offset, limit, order)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("leaves", records, page, limit, total)
}

func handleGetLeave(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	leaveID := intArg(m, "leave_id", 0)
	if leaveID <= 0 {
		return mcp.ToolResultError("leave_id must be a positive integer")
	}

	fields := []string{
		"id", "name", "employee_id", "holiday_status_id",
		"date_from", "date_to", "request_date_from", "request_date_to",
		"number_of_days", "number_of_hours", "state", "user_id", "notes",
	}

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "hr.leave", []int{leaveID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("leave request %d not found", leaveID))
	}

	return jsonResult(map[string]interface{}{"leave": records[0]})
}

func handleCreateLeaveRequest(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	empID := intArg(m, "employee_id", 0)
	statusID := intArg(m, "holiday_status_id", 0)
	dateFrom := strArg(m, "request_date_from")
	dateTo := strArg(m, "request_date_to")

	if empID <= 0 || statusID <= 0 || dateFrom == "" || dateTo == "" {
		return mcp.ToolResultError("employee_id, holiday_status_id, request_date_from, and request_date_to are required")
	}

	vals := map[string]interface{}{
		"employee_id":       empID,
		"holiday_status_id": statusID,
		"request_date_from": dateFrom,
		"request_date_to":   dateTo,
	}
	if name := strArg(m, "name"); name != "" {
		vals["name"] = name
	}
	if days := floatArg(m, "number_of_days", 0); days > 0 {
		vals["number_of_days"] = days
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "hr.leave", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"leave_id": newID,
		"success":  true,
	})
}

func handleActionLeave(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	leaveID := intArg(m, "leave_id", 0)
	action := strings.ToLower(strArg(m, "action"))
	if leaveID <= 0 || action == "" {
		return mcp.ToolResultError("leave_id and action are required")
	}

	var method string
	switch action {
	case "approve":
		method = "action_approve"
	case "validate":
		method = "action_validate"
	case "refuse":
		method = "action_refuse"
	default:
		return mcp.ToolResultError(fmt.Sprintf("unsupported action %q: must be 'approve', 'validate', or 'refuse'", action))
	}

	c, cancel := ctx()
	defer cancel()

	res, err := client.CallMethod(c, "hr.leave", method, []int{leaveID}, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"leave_id": leaveID,
		"action":   action,
		"result":   res,
		"success":  true,
	})
}

func handleListLeaveTypes(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if boolArg(m, "active", true) {
		domain = append(domain, []interface{}{"active", "=", true})
	}

	fields := []string{
		"id", "name", "requires_allocation", "max_leaves", "leaves_taken", "color",
	}

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "hr.leave.type", domain, fields, 0, 100, "sequence asc, name asc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"leave_types": records,
		"total_count": total,
	})
}

func handleGetLeaveBalances(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	empID := intArg(m, "employee_id", 0)
	if empID <= 0 {
		return mcp.ToolResultError("employee_id must be a positive integer")
	}

	domain := []interface{}{
		[]interface{}{"employee_id", "=", empID},
		[]interface{}{"state", "=", "validate"},
	}

	fields := []string{
		"id", "name", "holiday_status_id", "number_of_days", "date_from", "date_to", "state",
	}

	c, cancel := ctx()
	defer cancel()

	allocations, total, err := client.SearchRead(c, "hr.leave.allocation", domain, fields, 0, 50, "id desc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"employee_id": empID,
		"allocations": allocations,
		"total_count": total,
	})
}
