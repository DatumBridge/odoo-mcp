package tools

import (
	"encoding/json"
	"fmt"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

var defaultPayslipFields = []string{
	"id",
	"name",
	"number",
	"employee_id",
	"date_from",
	"date_to",
	"state",
	"contract_id",
	"struct_id",
}

var defaultContractFields = []string{
	"id",
	"name",
	"employee_id",
	"department_id",
	"job_id",
	"wage",
	"state",
	"date_start",
	"date_end",
	"structure_type_id",
}

func registerPayrollTools(add toolAdder) {
	// 1. odoo_list_payslips
	add(
		"odoo_list_payslips",
		"List employee payslips / salary slips (hr.payslip) with state and date filters.",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by employee ID.",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Filter status: 'draft', 'verify' (waiting), 'done', 'cancel'.",
			},
			"date_from": map[string]interface{}{
				"type":        "string",
				"description": "Filter period starting on or after (YYYY-MM-DD).",
			},
			"date_to": map[string]interface{}{
				"type":        "string",
				"description": "Filter period ending on or before (YYYY-MM-DD).",
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
		handleListPayslips,
	)

	// 2. odoo_get_payslip
	add(
		"odoo_get_payslip",
		"Get detailed payslip record by ID including salary breakdown rules (BASIC, GROSS, NET, deductions).",
		baseProps(map[string]interface{}{
			"payslip_id": map[string]interface{}{
				"type":        "integer",
				"description": "Payslip ID (hr.payslip).",
			},
		}),
		[]string{"payslip_id"},
		handleGetPayslip,
	)

	// 3. odoo_list_contracts
	add(
		"odoo_list_contracts",
		"List employee salary contracts (hr.contract) with wage, job, and duration info.",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by employee ID.",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Contract status: 'draft', 'open' (running), 'close' (expired), 'cancel'.",
			},
			"department_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by department ID.",
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
		handleListContracts,
	)

	// 4. odoo_get_contract
	add(
		"odoo_get_contract",
		"Get detailed salary contract terms and compensation details by contract ID.",
		baseProps(map[string]interface{}{
			"contract_id": map[string]interface{}{
				"type":        "integer",
				"description": "Contract ID (hr.contract).",
			},
		}),
		[]string{"contract_id"},
		handleGetContract,
	)

	// 5. odoo_create_contract
	add(
		"odoo_create_contract",
		"Create an employment salary contract for an employee in Odoo (hr.contract).",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Contract title or reference, e.g. 'Hợp đồng lao động - Nguyễn Văn A'.",
			},
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Employee ID (hr.employee).",
			},
			"wage": map[string]interface{}{
				"type":        "number",
				"description": "Monthly base salary/wage amount.",
			},
			"date_start": map[string]interface{}{
				"type":        "string",
				"description": "Contract start date (YYYY-MM-DD).",
			},
			"date_end": map[string]interface{}{
				"type":        "string",
				"description": "Contract end date (YYYY-MM-DD, optional for indefinite term).",
			},
			"department_id": map[string]interface{}{
				"type":        "integer",
				"description": "Department ID (hr.department).",
			},
			"job_id": map[string]interface{}{
				"type":        "integer",
				"description": "Job position ID (hr.job).",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Contract status: 'draft', 'open' (running), default 'open'.",
			},
			"extra_fields": map[string]interface{}{
				"type":        "object",
				"description": "Additional fields for hr.contract.",
			},
		}),
		[]string{"name", "employee_id", "wage", "date_start"},
		handleCreateContract,
	)

	// 6. odoo_update_contract
	add(
		"odoo_update_contract",
		"Update terms or wage on an existing employee contract (hr.contract).",
		baseProps(map[string]interface{}{
			"contract_id": map[string]interface{}{
				"type":        "integer",
				"description": "Contract ID (hr.contract).",
			},
			"values": map[string]interface{}{
				"type":        "object",
				"description": "Key-value map of contract fields to update (e.g. wage, date_end, state).",
			},
		}),
		[]string{"contract_id", "values"},
		handleUpdateContract,
	)
}

func handleListPayslips(args json.RawMessage) map[string]interface{} {
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
	if dFrom := strArg(m, "date_from"); dFrom != "" {
		domain = append(domain, []interface{}{"date_from", ">=", dFrom})
	}
	if dTo := strArg(m, "date_to"); dTo != "" {
		domain = append(domain, []interface{}{"date_to", "<=", dTo})
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "hr.payslip", domain, defaultPayslipFields, offset, limit, "date_from desc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("payslips", records, page, limit, total)
}

func handleGetPayslip(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	slipID := intArg(m, "payslip_id", 0)
	if slipID <= 0 {
		return mcp.ToolResultError("payslip_id must be a positive integer")
	}

	fields := append(defaultPayslipFields, "line_ids", "note")

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "hr.payslip", []int{slipID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("payslip %d not found", slipID))
	}

	payslip := records[0]

	// Fetch salary breakdown lines if line_ids present
	var lineIDs []int
	if rawLines, ok := payslip["line_ids"].([]interface{}); ok {
		for _, lid := range rawLines {
			if idNum, ok := lid.(float64); ok {
				lineIDs = append(lineIDs, int(idNum))
			}
		}
	}

	if len(lineIDs) > 0 {
		lineFields := []string{"name", "code", "category_id", "total", "amount", "rate"}
		lines, err := client.Read(c, "hr.payslip.line", lineIDs, lineFields)
		if err == nil {
			payslip["salary_breakdown"] = lines
		}
	}

	return jsonResult(map[string]interface{}{"payslip": payslip})
}

func handleListContracts(args json.RawMessage) map[string]interface{} {
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
	if deptID := intArg(m, "department_id", 0); deptID > 0 {
		domain = append(domain, []interface{}{"department_id", "=", deptID})
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "hr.contract", domain, defaultContractFields, offset, limit, "date_start desc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("contracts", records, page, limit, total)
}

func handleGetContract(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	contractID := intArg(m, "contract_id", 0)
	if contractID <= 0 {
		return mcp.ToolResultError("contract_id must be a positive integer")
	}

	fields := append(defaultContractFields, "notes", "hr_responsible_id")

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "hr.contract", []int{contractID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("contract %d not found", contractID))
	}

	return jsonResult(map[string]interface{}{"contract": records[0]})
}

func handleCreateContract(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	name := strArg(m, "name")
	empID := intArg(m, "employee_id", 0)
	wage := floatArg(m, "wage", 0)
	dateStart := strArg(m, "date_start")

	if name == "" || empID <= 0 || wage <= 0 || dateStart == "" {
		return mcp.ToolResultError("name, employee_id, wage, and date_start are required to create a contract")
	}

	vals := map[string]interface{}{
		"name":        name,
		"employee_id": empID,
		"wage":        wage,
		"date_start":  dateStart,
	}
	if dateEnd := strArg(m, "date_end"); dateEnd != "" {
		vals["date_end"] = dateEnd
	}
	if deptID := intArg(m, "department_id", 0); deptID > 0 {
		vals["department_id"] = deptID
	}
	if jobID := intArg(m, "job_id", 0); jobID > 0 {
		vals["job_id"] = jobID
	}
	if state := strArg(m, "state"); state != "" {
		vals["state"] = state
	} else {
		vals["state"] = "open"
	}
	if extra := mapArg(m, "extra_fields"); len(extra) > 0 {
		for k, v := range extra {
			vals[k] = v
		}
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "hr.contract", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"contract_id": newID,
		"name":        name,
		"success":     true,
	})
}

func handleUpdateContract(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	contractID := intArg(m, "contract_id", 0)
	if contractID <= 0 {
		return mcp.ToolResultError("contract_id must be a positive integer")
	}

	values := mapArg(m, "values")
	if len(values) == 0 {
		return mcp.ToolResultError("values map cannot be empty")
	}

	c, cancel := ctx()
	defer cancel()

	ok, err := client.Write(c, "hr.contract", []int{contractID}, values)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"contract_id": contractID,
		"success":     ok,
	})
}

