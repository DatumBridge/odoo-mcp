# odoo-mcp

Streamable HTTP MCP tool-server for Odoo ERP integration.

## Features

Supports comprehensive business domains in Odoo:
- **Human Resources (HR)**: `odoo_search_employees`, `odoo_get_employee`, `odoo_create_employee`, `odoo_update_employee`, `odoo_list_departments`
- **Time Off & Leaves**: `odoo_list_leaves`, `odoo_get_leave`, `odoo_create_leave_request`, `odoo_action_leave`, `odoo_list_leave_types`, `odoo_get_leave_balances`
- **Payroll & Contracts**: `odoo_list_payslips`, `odoo_get_payslip`, `odoo_list_contracts`, `odoo_get_contract`, `odoo_create_contract`, `odoo_update_contract`
- **Projects & Tasks**: `odoo_list_projects`, `odoo_get_project`, `odoo_list_tasks`, `odoo_create_task`, `odoo_update_task`
- **CRM & Customer Care**: `odoo_search_partners`, `odoo_get_partner`, `odoo_create_partner`, `odoo_update_partner`, `odoo_list_crm_leads`, `odoo_get_crm_lead`, `odoo_create_crm_lead`, `odoo_update_crm_lead`, `odoo_create_activity`, `odoo_list_activities`
- **Sales & Quotations**: `odoo_search_products`, `odoo_get_product`, `odoo_list_quotations`, `odoo_get_quotation`, `odoo_create_quotation`, `odoo_update_quotation`, `odoo_action_quotation`
- **Universal ORM Execution**: `odoo_execute_kw`, `odoo_describe_model`

## Protocol & Ports

- **Protocol**: MCP Streamable HTTP (`2024-11-05`), `Mcp-Session-Id` header
- **Endpoints**:
  - `GET /health`: Health check
  - `POST /mcp`: MCP JSON-RPC protocol endpoint
- **Port**: Default `8018` (configurable via `ODOO_MCP_PORT` or `PORT`)

## Run & Test

```bash
# Run unit & integration tests
go test -v ./...

# Build binary
go build -o bin/odoo-mcp ./cmd/api

# Run server
./bin/odoo-mcp
```
