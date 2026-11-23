package tools

import (
	"encoding/json"
	"fmt"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

var defaultEmployeeFields = []string{
	"id",
	"name",
	"work_email",
	"work_phone",
	"mobile_phone",
	"department_id",
	"job_id",
	"job_title",
	"parent_id",
	"coach_id",
	"work_location_id",
	"active",
}

func registerHRTools(add toolAdder) {
	// 1. odoo_search_employees
	add(
		"odoo_search_employees",
		"Search and list Odoo employees with filters by name, work email, department, job, or manager.",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Partial employee name search.",
			},
			"work_email": map[string]interface{}{
				"type":        "string",
				"description": "Filter by employee work email.",
			},
			"department_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by department ID.",
			},
			"job_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by job position ID.",
			},
			"parent_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by manager/parent employee ID.",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter active employees (default true).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific fields to return. Omit for standard fields.",
			},
			"domain": map[string]interface{}{
				"type":        "array",
				"description": "Optional raw Odoo domain filter expression.",
			},
			"order": map[string]interface{}{
				"type":        "string",
				"description": "Sort order, e.g. 'name asc' or 'id desc'.",
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
		handleSearchEmployees,
	)

	// 2. odoo_get_employee
	add(
		"odoo_get_employee",
		"Get detailed information of a single Odoo employee by employee ID.",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Odoo employee ID (hr.employee).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to return. Omit for full employee profile.",
			},
		}),
		[]string{"employee_id"},
		handleGetEmployee,
	)

	// 3. odoo_create_employee
	add(
		"odoo_create_employee",
		"Create a new employee record in Odoo (hr.employee).",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Employee full name.",
			},
			"work_email": map[string]interface{}{
				"type":        "string",
				"description": "Work email address.",
			},
			"work_phone": map[string]interface{}{
				"type":        "string",
				"description": "Work phone number.",
			},
			"department_id": map[string]interface{}{
				"type":        "integer",
				"description": "Department ID (hr.department).",
			},
			"job_id": map[string]interface{}{
				"type":        "integer",
				"description": "Job position ID (hr.job).",
			},
			"job_title": map[string]interface{}{
				"type":        "string",
				"description": "Custom job title string.",
			},
			"parent_id": map[string]interface{}{
				"type":        "integer",
				"description": "Manager/Supervisor employee ID.",
			},
			"coach_id": map[string]interface{}{
				"type":        "integer",
				"description": "Mentor/Coach employee ID.",
			},
			"extra_fields": map[string]interface{}{
				"type":        "object",
				"description": "Additional arbitrary fields for hr.employee.",
			},
		}),
		[]string{"name"},
		handleCreateEmployee,
	)

	// 4. odoo_update_employee
	add(
		"odoo_update_employee",
		"Update fields on an existing employee record (hr.employee).",
		baseProps(map[string]interface{}{
			"employee_id": map[string]interface{}{
				"type":        "integer",
				"description": "Odoo employee ID.",
			},
			"values": map[string]interface{}{
				"type":        "object",
				"description": "Key-value map of fields to update.",
			},
		}),
		[]string{"employee_id", "values"},
		handleUpdateEmployee,
	)

	// 5. odoo_list_departments
	add(
		"odoo_list_departments",
		"List Odoo departments (hr.department) with hierarchy and manager info.",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Filter by department name (partial match).",
			},
			"parent_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by parent department ID.",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to return. Omit for default fields.",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number (1-based, default 1).",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page (default 50).",
			},
		}),
		nil,
		handleListDepartments,
	)
}

func handleSearchEmployees(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if rawDomain := domainArg(m, "domain"); len(rawDomain) > 0 {
		domain = rawDomain
	} else {
		if name := strArg(m, "name"); name != "" {
			domain = append(domain, []interface{}{"name", "ilike", name})
		}
		if email := strArg(m, "work_email"); email != "" {
			domain = append(domain, []interface{}{"work_email", "ilike", email})
		}
		if deptID := intArg(m, "department_id", 0); deptID > 0 {
			domain = append(domain, []interface{}{"department_id", "=", deptID})
		}
		if jobID := intArg(m, "job_id", 0); jobID > 0 {
			domain = append(domain, []interface{}{"job_id", "=", jobID})
		}
		if parentID := intArg(m, "parent_id", 0); parentID > 0 {
			domain = append(domain, []interface{}{"parent_id", "=", parentID})
		}
		if boolArg(m, "active", true) {
			domain = append(domain, []interface{}{"active", "=", true})
		}
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultEmployeeFields
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	order := strArg(m, "order")
	if order == "" {
		order = "name asc"
	}

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "hr.employee", domain, fields, offset, limit, order)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("employees", records, page, limit, total)
}

func handleGetEmployee(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	empID := intArg(m, "employee_id", 0)
	if empID <= 0 {
		return mcp.ToolResultError("employee_id must be a positive integer")
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = append(defaultEmployeeFields,
			"birthday", "gender", "marital", "identification_id", "passport_id",
			"emergency_contact", "emergency_phone", "notes",
		)
	}

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "hr.employee", []int{empID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("employee with id %d not found", empID))
	}

	return jsonResult(map[string]interface{}{"employee": records[0]})
}

func handleCreateEmployee(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	name := strArg(m, "name")
	if name == "" {
		return mcp.ToolResultError("name is required to create an employee")
	}

	vals := map[string]interface{}{
		"name": name,
	}
	if email := strArg(m, "work_email"); email != "" {
		vals["work_email"] = email
	}
	if phone := strArg(m, "work_phone"); phone != "" {
		vals["work_phone"] = phone
	}
	if deptID := intArg(m, "department_id", 0); deptID > 0 {
		vals["department_id"] = deptID
	}
	if jobID := intArg(m, "job_id", 0); jobID > 0 {
		vals["job_id"] = jobID
	}
	if jobTitle := strArg(m, "job_title"); jobTitle != "" {
		vals["job_title"] = jobTitle
	}
	if parentID := intArg(m, "parent_id", 0); parentID > 0 {
		vals["parent_id"] = parentID
	}
	if coachID := intArg(m, "coach_id", 0); coachID > 0 {
		vals["coach_id"] = coachID
	}
	if extra := mapArg(m, "extra_fields"); len(extra) > 0 {
		for k, v := range extra {
			vals[k] = v
		}
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "hr.employee", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"employee_id": newID,
		"name":        name,
		"success":     true,
	})
}

func handleUpdateEmployee(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	empID := intArg(m, "employee_id", 0)
	if empID <= 0 {
		return mcp.ToolResultError("employee_id must be a positive integer")
	}

	values := mapArg(m, "values")
	if len(values) == 0 {
		return mcp.ToolResultError("values map cannot be empty")
	}

	c, cancel := ctx()
	defer cancel()

	ok, err := client.Write(c, "hr.employee", []int{empID}, values)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"employee_id": empID,
		"success":     ok,
	})
}

func handleListDepartments(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if name := strArg(m, "name"); name != "" {
		domain = append(domain, []interface{}{"name", "ilike", name})
	}
	if parentID := intArg(m, "parent_id", 0); parentID > 0 {
		domain = append(domain, []interface{}{"parent_id", "=", parentID})
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = []string{"id", "name", "complete_name", "manager_id", "parent_id", "company_id"}
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 50)
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "hr.department", domain, fields, offset, limit, "name asc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("departments", records, page, limit, total)
}
