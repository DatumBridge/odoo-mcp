package tools

import (
	"encoding/json"
	"fmt"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

var defaultProjectFields = []string{
	"id",
	"name",
	"user_id",
	"partner_id",
	"date_start",
	"date",
	"privacy_visibility",
	"active",
	"task_count",
	"description",
}

var defaultTaskFields = []string{
	"id",
	"name",
	"project_id",
	"user_ids",
	"stage_id",
	"date_deadline",
	"priority",
	"kanban_state",
	"partner_id",
	"active",
}

func registerProjectTools(add toolAdder) {
	// 1. odoo_list_projects
	add(
		"odoo_list_projects",
		"List and search Odoo projects (project.project) with manager, customer, and name filters.",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Partial search by project name.",
			},
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by customer/client partner ID.",
			},
			"user_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by project manager user ID.",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter active projects (default true).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific fields to return. Omit for standard fields.",
			},
			"order": map[string]interface{}{
				"type":        "string",
				"description": "Sort order, e.g. 'name asc' or 'date desc'.",
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
		handleListProjects,
	)

	// 2. odoo_get_project
	add(
		"odoo_get_project",
		"Get comprehensive details of a single Odoo project by ID (project.project) including manager, customer, and task count.",
		baseProps(map[string]interface{}{
			"project_id": map[string]interface{}{
				"type":        "integer",
				"description": "Odoo project ID (project.project).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to return. Omit for full project summary.",
			},
		}),
		[]string{"project_id"},
		handleGetProject,
	)

	// 3. odoo_list_tasks
	add(
		"odoo_list_tasks",
		"List and filter project tasks (project.task) by project, assigned user, stage, deadline range, or priority.",
		baseProps(map[string]interface{}{
			"project_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by project ID.",
			},
			"user_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter tasks assigned to this user ID.",
			},
			"stage_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by kanban stage ID.",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Filter priority: '0' normal, '1' urgent/starred.",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Partial search in task name/title.",
			},
			"date_deadline_before": map[string]interface{}{
				"type":        "string",
				"description": "Filter tasks with deadline on or before this date (YYYY-MM-DD).",
			},
			"date_deadline_after": map[string]interface{}{
				"type":        "string",
				"description": "Filter tasks with deadline on or after this date (YYYY-MM-DD).",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter active tasks (default true).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific fields to return.",
			},
			"order": map[string]interface{}{
				"type":        "string",
				"description": "Sort order, e.g. 'date_deadline asc, priority desc'.",
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
		handleListTasks,
	)

	// 4. odoo_create_task
	add(
		"odoo_create_task",
		"Create a new task in Odoo (project.task) assigned to users, linked to project, with deadline and priority.",
		baseProps(map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Task title or description summary.",
			},
			"project_id": map[string]interface{}{
				"type":        "integer",
				"description": "Target project ID (project.project).",
			},
			"user_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "integer"},
				"description": "Array of user IDs (res.users) assigned to this task.",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Detailed task instructions, requirements, or checklist.",
			},
			"date_deadline": map[string]interface{}{
				"type":        "string",
				"description": "Task deadline date (YYYY-MM-DD).",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Priority: '0' for normal, '1' for urgent/important.",
			},
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Customer contact partner ID (res.partner).",
			},
			"extra_fields": map[string]interface{}{
				"type":        "object",
				"description": "Additional arbitrary fields for project.task.",
			},
		}),
		[]string{"name"},
		handleCreateTask,
	)

	// 5. odoo_update_task
	add(
		"odoo_update_task",
		"Update an existing task (project.task) stage, assignees, deadline, priority, or description.",
		baseProps(map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "integer",
				"description": "Odoo task ID (project.task).",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Updated task title.",
			},
			"project_id": map[string]interface{}{
				"type":        "integer",
				"description": "Move task to another project ID.",
			},
			"stage_id": map[string]interface{}{
				"type":        "integer",
				"description": "Change task stage ID (e.g. In Progress, Done).",
			},
			"user_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "integer"},
				"description": "Update assigned user IDs.",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Updated task description or notes.",
			},
			"date_deadline": map[string]interface{}{
				"type":        "string",
				"description": "Updated deadline date (YYYY-MM-DD).",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Updated priority: '0' or '1'.",
			},
			"kanban_state": map[string]interface{}{
				"type":        "string",
				"description": "Kanban status: 'normal', 'done', 'blocked'.",
			},
			"values": map[string]interface{}{
				"type":        "object",
				"description": "Map of field-value pairs to update.",
			},
			"extra_fields": map[string]interface{}{
				"type":        "object",
				"description": "Additional arbitrary fields for project.task.",
			},
		}),
		[]string{"task_id"},
		handleUpdateTask,
	)
}

func handleListProjects(args json.RawMessage) map[string]interface{} {
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
		if partnerID := intArg(m, "partner_id", 0); partnerID > 0 {
			domain = append(domain, []interface{}{"partner_id", "=", partnerID})
		}
		if userID := intArg(m, "user_id", 0); userID > 0 {
			domain = append(domain, []interface{}{"user_id", "=", userID})
		}
		if boolArg(m, "active", true) {
			domain = append(domain, []interface{}{"active", "=", true})
		}
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultProjectFields
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

	records, total, err := client.SearchRead(c, "project.project", domain, fields, offset, limit, order)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("projects", records, page, limit, total)
}

func handleGetProject(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	projectID := intArg(m, "project_id", 0)
	if projectID <= 0 {
		return mcp.ToolResultError("project_id is required and must be > 0")
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultProjectFields
	}

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "project.project", []int{projectID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("project with id %d not found", projectID))
	}

	return jsonResult(map[string]interface{}{
		"project": records[0],
	})
}

func handleListTasks(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if rawDomain := domainArg(m, "domain"); len(rawDomain) > 0 {
		domain = rawDomain
	} else {
		if projID := intArg(m, "project_id", 0); projID > 0 {
			domain = append(domain, []interface{}{"project_id", "=", projID})
		}
		if userID := intArg(m, "user_id", 0); userID > 0 {
			domain = append(domain, []interface{}{"user_ids", "in", []int{userID}})
		}
		if stageID := intArg(m, "stage_id", 0); stageID > 0 {
			domain = append(domain, []interface{}{"stage_id", "=", stageID})
		}
		if prio := strArg(m, "priority"); prio != "" {
			domain = append(domain, []interface{}{"priority", "=", prio})
		}
		if name := strArg(m, "name"); name != "" {
			domain = append(domain, []interface{}{"name", "ilike", name})
		}
		if deadlineBefore := strArg(m, "date_deadline_before"); deadlineBefore != "" {
			domain = append(domain, []interface{}{"date_deadline", "<=", deadlineBefore})
		}
		if deadlineAfter := strArg(m, "date_deadline_after"); deadlineAfter != "" {
			domain = append(domain, []interface{}{"date_deadline", ">=", deadlineAfter})
		}
		if boolArg(m, "active", true) {
			domain = append(domain, []interface{}{"active", "=", true})
		}
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultTaskFields
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	order := strArg(m, "order")
	if order == "" {
		order = "date_deadline asc, priority desc, id desc"
	}

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "project.task", domain, fields, offset, limit, order)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("tasks", records, page, limit, total)
}

func handleCreateTask(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	name := strArg(m, "name")
	if name == "" {
		return mcp.ToolResultError("name is required to create a task")
	}

	vals := map[string]interface{}{
		"name": name,
	}
	if projID := intArg(m, "project_id", 0); projID > 0 {
		vals["project_id"] = projID
	}
	if userIDs := sliceIntArg(m, "user_ids"); len(userIDs) > 0 {
		vals["user_ids"] = []interface{}{[]interface{}{6, 0, userIDs}}
	}
	if desc := strArg(m, "description"); desc != "" {
		vals["description"] = desc
	}
	if deadline := strArg(m, "date_deadline"); deadline != "" {
		vals["date_deadline"] = deadline
	}
	if prio := strArg(m, "priority"); prio != "" {
		vals["priority"] = prio
	}
	if partnerID := intArg(m, "partner_id", 0); partnerID > 0 {
		vals["partner_id"] = partnerID
	}
	for k, v := range mapArg(m, "extra_fields") {
		vals[k] = v
	}

	c, cancel := ctx()
	defer cancel()

	taskID, err := client.Create(c, "project.task", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"task_id": taskID,
		"name":    name,
		"success": true,
	})
}

func handleUpdateTask(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	taskID := intArg(m, "task_id", 0)
	if taskID <= 0 {
		return mcp.ToolResultError("task_id is required and must be > 0")
	}

	vals := map[string]interface{}{}
	if name := strArg(m, "name"); name != "" {
		vals["name"] = name
	}
	if projID := intArg(m, "project_id", 0); projID > 0 {
		vals["project_id"] = projID
	}
	if stageID := intArg(m, "stage_id", 0); stageID > 0 {
		vals["stage_id"] = stageID
	}
	if userIDs := sliceIntArg(m, "user_ids"); len(userIDs) > 0 {
		vals["user_ids"] = []interface{}{[]interface{}{6, 0, userIDs}}
	}
	if desc := strArg(m, "description"); desc != "" {
		vals["description"] = desc
	}
	if deadline := strArg(m, "date_deadline"); deadline != "" {
		vals["date_deadline"] = deadline
	}
	if prio := strArg(m, "priority"); prio != "" {
		vals["priority"] = prio
	}
	if kState := strArg(m, "kanban_state"); kState != "" {
		vals["kanban_state"] = kState
	}
	for k, v := range mapArg(m, "values") {
		vals[k] = v
	}
	for k, v := range mapArg(m, "extra_fields") {
		vals[k] = v
	}

	if len(vals) == 0 {
		return mcp.ToolResultError("no update values provided for task")
	}

	c, cancel := ctx()
	defer cancel()

	ok, err := client.Write(c, "project.task", []int{taskID}, vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"task_id": taskID,
		"success": ok,
	})
}
