package tools

import (
	"encoding/json"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

func registerGenericTools(add toolAdder) {
	// 1. odoo_execute_kw
	add(
		"odoo_execute_kw",
		"Execute any arbitrary Odoo ORM method (search, read, write, create, custom methods) on any model via JSON-RPC execute_kw.",
		baseProps(map[string]interface{}{
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Odoo model name (e.g. 'account.move', 'stock.picking', 'project.task', 'res.users').",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"description": "ORM method name (e.g. 'search_read', 'read', 'create', 'write', 'unlink', 'name_search').",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "Positional arguments to the method.",
			},
			"kwargs": map[string]interface{}{
				"type":        "object",
				"description": "Keyword arguments to the method.",
			},
		}),
		[]string{"model", "method"},
		handleExecuteKW,
	)

	// 2. odoo_describe_model
	add(
		"odoo_describe_model",
		"Inspect field schemas, types, and descriptions of any Odoo model via fields_get.",
		baseProps(map[string]interface{}{
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Odoo model name (e.g. 'hr.employee', 'sale.order', 'res.partner').",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific field names to inspect. If omitted, describes all fields.",
			},
			"attributes": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Field attributes to return (default: ['type', 'string', 'help', 'required', 'readonly', 'selection']).",
			},
		}),
		[]string{"model"},
		handleDescribeModel,
	)
}

func handleExecuteKW(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	model := strArg(m, "model")
	method := strArg(m, "method")
	if model == "" || method == "" {
		return mcp.ToolResultError("model and method are required")
	}

	rawArgs, _ := m["args"].([]interface{})
	kwargs := mapArg(m, "kwargs")

	c, cancel := ctx()
	defer cancel()

	res, err := client.ExecuteKW(c, model, method, rawArgs, kwargs)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"model":  model,
		"method": method,
		"result": res,
	})
}

func handleDescribeModel(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	model := strArg(m, "model")
	if model == "" {
		return mcp.ToolResultError("model is required")
	}

	fields := sliceStrArg(m, "fields")
	attrs := sliceStrArg(m, "attributes")
	if len(attrs) == 0 {
		attrs = []string{"type", "string", "help", "required", "readonly", "selection"}
	}

	c, cancel := ctx()
	defer cancel()

	res, err := client.FieldsGet(c, model, fields, attrs)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"model":  model,
		"fields": res,
	})
}
