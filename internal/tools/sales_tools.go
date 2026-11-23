package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datumbridge/odoo-mcp/internal/mcp"
)

var defaultProductFields = []string{
	"id",
	"name",
	"default_code",
	"barcode",
	"list_price",
	"standard_price",
	"qty_available",
	"uom_id",
	"categ_id",
	"type",
	"active",
}

var defaultOrderFields = []string{
	"id",
	"name",
	"partner_id",
	"date_order",
	"state",
	"amount_untaxed",
	"amount_tax",
	"amount_total",
	"currency_id",
	"user_id",
	"team_id",
	"client_order_ref",
}

func registerSalesTools(add toolAdder) {
	// 1. odoo_search_products
	add(
		"odoo_search_products",
		"Search products in Odoo catalog (product.product) by name, SKU/default_code, barcode, or category.",
		baseProps(map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Text search matching across product name, SKU (default_code), or barcode.",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Specific product name filter (partial match).",
			},
			"default_code": map[string]interface{}{
				"type":        "string",
				"description": "Internal reference / SKU filter.",
			},
			"categ_id": map[string]interface{}{
				"type":        "integer",
				"description": "Product category ID.",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Filter active products (default true).",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Specific fields to return.",
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
		handleSearchProducts,
	)

	// 2. odoo_get_product
	add(
		"odoo_get_product",
		"Get detailed product information by product ID (product.product).",
		baseProps(map[string]interface{}{
			"product_id": map[string]interface{}{
				"type":        "integer",
				"description": "Product ID.",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Fields to fetch.",
			},
		}),
		[]string{"product_id"},
		handleGetProduct,
	)

	// 3. odoo_list_quotations
	add(
		"odoo_list_quotations",
		"List quotations and sales orders (sale.order) with customer, state, and date filters.",
		baseProps(map[string]interface{}{
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Filter by customer/partner ID.",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Filter state: 'draft' (quotation), 'sent' (quotation sent), 'sale' (sales order), 'done', 'cancel'.",
			},
			"date_order_from": map[string]interface{}{
				"type":        "string",
				"description": "Filter orders placed on or after (YYYY-MM-DD).",
			},
			"date_order_to": map[string]interface{}{
				"type":        "string",
				"description": "Filter orders placed on or before (YYYY-MM-DD).",
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
		handleListQuotations,
	)

	// 4. odoo_get_quotation
	add(
		"odoo_get_quotation",
		"Get full quotation or sales order details by ID (sale.order), including order lines and prices.",
		baseProps(map[string]interface{}{
			"order_id": map[string]interface{}{
				"type":        "integer",
				"description": "Sales order or quotation ID (sale.order).",
			},
		}),
		[]string{"order_id"},
		handleGetQuotation,
	)

	// 5. odoo_create_quotation
	add(
		"odoo_create_quotation",
		"Create a new quotation / sales order (sale.order) with line items in Odoo.",
		baseProps(map[string]interface{}{
			"partner_id": map[string]interface{}{
				"type":        "integer",
				"description": "Customer/partner ID (res.partner).",
			},
			"lines": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"product_id":      map[string]interface{}{"type": "integer", "description": "Product ID (product.product)."},
						"product_uom_qty": map[string]interface{}{"type": "number", "description": "Quantity to order."},
						"price_unit":      map[string]interface{}{"type": "number", "description": "Unit price (optional, defaults to product list price)."},
						"name":            map[string]interface{}{"type": "string", "description": "Line description (optional)."},
					},
					"required": []string{"product_id", "product_uom_qty"},
				},
				"description": "List of quotation line items.",
			},
			"client_order_ref": map[string]interface{}{
				"type":        "string",
				"description": "Customer reference / PO number.",
			},
			"note": map[string]interface{}{
				"type":        "string",
				"description": "Terms and conditions or notes for the quotation.",
			},
			"user_id": map[string]interface{}{
				"type":        "integer",
				"description": "Salesperson user ID.",
			},
		}),
		[]string{"partner_id", "lines"},
		handleCreateQuotation,
	)

	// 6. odoo_update_quotation
	add(
		"odoo_update_quotation",
		"Update fields on an existing quotation or sales order (sale.order).",
		baseProps(map[string]interface{}{
			"order_id": map[string]interface{}{
				"type":        "integer",
				"description": "Quotation/Order ID to update.",
			},
			"values": map[string]interface{}{
				"type":        "object",
				"description": "Map of field values to update.",
			},
		}),
		[]string{"order_id", "values"},
		handleUpdateQuotation,
	)

	// 7. odoo_action_quotation
	add(
		"odoo_action_quotation",
		"Perform workflow action on quotation/sales order: 'confirm' (turn quotation into sales order) or 'cancel'.",
		baseProps(map[string]interface{}{
			"order_id": map[string]interface{}{
				"type":        "integer",
				"description": "Quotation/Order ID.",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: 'confirm' or 'cancel'.",
			},
		}),
		[]string{"order_id", "action"},
		handleActionQuotation,
	)
}

func handleSearchProducts(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if q := strArg(m, "query"); q != "" {
		domain = append(domain, "|", "|",
			[]interface{}{"name", "ilike", q},
			[]interface{}{"default_code", "ilike", q},
			[]interface{}{"barcode", "ilike", q},
		)
	}
	if name := strArg(m, "name"); name != "" {
		domain = append(domain, []interface{}{"name", "ilike", name})
	}
	if sku := strArg(m, "default_code"); sku != "" {
		domain = append(domain, []interface{}{"default_code", "ilike", sku})
	}
	if categID := intArg(m, "categ_id", 0); categID > 0 {
		domain = append(domain, []interface{}{"categ_id", "=", categID})
	}
	if boolArg(m, "active", true) {
		domain = append(domain, []interface{}{"active", "=", true})
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = defaultProductFields
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "product.product", domain, fields, offset, limit, "name asc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("products", records, page, limit, total)
}

func handleGetProduct(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	prodID := intArg(m, "product_id", 0)
	if prodID <= 0 {
		return mcp.ToolResultError("product_id must be a positive integer")
	}

	fields := sliceStrArg(m, "fields")
	if len(fields) == 0 {
		fields = append(defaultProductFields,
			"description_sale", "weight", "volume", "taxes_id",
			"seller_ids", "virtual_available",
		)
	}

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "product.product", []int{prodID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("product %d not found", prodID))
	}

	return jsonResult(map[string]interface{}{"product": records[0]})
}

func handleListQuotations(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	var domain []interface{}
	if partnerID := intArg(m, "partner_id", 0); partnerID > 0 {
		domain = append(domain, []interface{}{"partner_id", "=", partnerID})
	}
	if state := strArg(m, "state"); state != "" {
		domain = append(domain, []interface{}{"state", "=", state})
	}
	if dFrom := strArg(m, "date_order_from"); dFrom != "" {
		domain = append(domain, []interface{}{"date_order", ">=", dFrom})
	}
	if dTo := strArg(m, "date_order_to"); dTo != "" {
		domain = append(domain, []interface{}{"date_order", "<=", dTo})
	}

	page := intArg(m, "page", 1)
	limit := intArg(m, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	c, cancel := ctx()
	defer cancel()

	records, total, err := client.SearchRead(c, "sale.order", domain, defaultOrderFields, offset, limit, "date_order desc")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return pagedResult("quotations", records, page, limit, total)
}

func handleGetQuotation(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	orderID := intArg(m, "order_id", 0)
	if orderID <= 0 {
		return mcp.ToolResultError("order_id must be a positive integer")
	}

	fields := append(defaultOrderFields, "order_line", "note", "payment_term_id", "incoterm")

	c, cancel := ctx()
	defer cancel()

	records, err := client.Read(c, "sale.order", []int{orderID}, fields)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	if len(records) == 0 {
		return mcp.ToolResultError(fmt.Sprintf("quotation/order %d not found", orderID))
	}

	order := records[0]

	// Fetch line items details
	var lineIDs []int
	if rawLines, ok := order["order_line"].([]interface{}); ok {
		for _, lid := range rawLines {
			if idNum, ok := lid.(float64); ok {
				lineIDs = append(lineIDs, int(idNum))
			}
		}
	}

	if len(lineIDs) > 0 {
		lineFields := []string{
			"id", "product_id", "name", "product_uom_qty", "product_uom",
			"price_unit", "price_subtotal", "price_total", "discount",
		}
		lines, err := client.Read(c, "sale.order.line", lineIDs, lineFields)
		if err == nil {
			order["order_lines"] = lines
		}
	}

	return jsonResult(map[string]interface{}{"quotation": order})
}

func handleCreateQuotation(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	partnerID := intArg(m, "partner_id", 0)
	if partnerID <= 0 {
		return mcp.ToolResultError("partner_id is required to create a quotation")
	}

	rawLines := sliceMapArg(m, "lines")
	if len(rawLines) == 0 {
		return mcp.ToolResultError("lines must contain at least one order line item")
	}

	// Build Odoo ORM command tuples: (0, 0, values)
	var orderLines []interface{}
	for _, l := range rawLines {
		prodID := intArg(l, "product_id", 0)
		qty := floatArg(l, "product_uom_qty", 0)
		if prodID <= 0 || qty <= 0 {
			return mcp.ToolResultError("each line requires product_id > 0 and product_uom_qty > 0")
		}

		lineVal := map[string]interface{}{
			"product_id":      prodID,
			"product_uom_qty": qty,
		}
		if price := floatArg(l, "price_unit", -1); price >= 0 {
			lineVal["price_unit"] = price
		}
		if name := strArg(l, "name"); name != "" {
			lineVal["name"] = name
		}

		orderLines = append(orderLines, []interface{}{0, 0, lineVal})
	}

	vals := map[string]interface{}{
		"partner_id": partnerID,
		"order_line": orderLines,
	}
	if ref := strArg(m, "client_order_ref"); ref != "" {
		vals["client_order_ref"] = ref
	}
	if note := strArg(m, "note"); note != "" {
		vals["note"] = note
	}
	if userID := intArg(m, "user_id", 0); userID > 0 {
		vals["user_id"] = userID
	}

	c, cancel := ctx()
	defer cancel()

	newID, err := client.Create(c, "sale.order", vals)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"order_id": newID,
		"success":  true,
	})
}

func handleUpdateQuotation(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	orderID := intArg(m, "order_id", 0)
	if orderID <= 0 {
		return mcp.ToolResultError("order_id must be a positive integer")
	}

	values := mapArg(m, "values")
	if len(values) == 0 {
		return mcp.ToolResultError("values map cannot be empty")
	}

	c, cancel := ctx()
	defer cancel()

	ok, err := client.Write(c, "sale.order", []int{orderID}, values)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"order_id": orderID,
		"success":  ok,
	})
}

func handleActionQuotation(args json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(args)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	orderID := intArg(m, "order_id", 0)
	action := strings.ToLower(strArg(m, "action"))
	if orderID <= 0 || action == "" {
		return mcp.ToolResultError("order_id and action are required")
	}

	var method string
	switch action {
	case "confirm":
		method = "action_confirm"
	case "cancel":
		method = "action_cancel"
	default:
		return mcp.ToolResultError(fmt.Sprintf("unsupported action %q: must be 'confirm' or 'cancel'", action))
	}

	c, cancel := ctx()
	defer cancel()

	res, err := client.CallMethod(c, "sale.order", method, []int{orderID}, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}

	return jsonResult(map[string]interface{}{
		"order_id": orderID,
		"action":   action,
		"result":   res,
		"success":  true,
	})
}
