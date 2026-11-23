# Design — odoo-mcp

## Class

**Tool-server** (Streamable HTTP MCP), matching `shopify-mcp` and `magento-mcp`.

## Language

Go 1.23, matching the platform standard established in `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`.

## Upstream

- Odoo JSON-RPC API (`{base_url}/jsonrpc`)
- `{base_url}`, `{db}`, `{username}`, and `{api_key}` come directly from the injected credential bundle (`credentials_json`), never from tool arguments (preventing multi-tenant confusion).

## Auth

- Method: Odoo JSON-RPC authentication (`common.authenticate`) and ORM execution (`object.execute_kw`).
- Production: Vault `credentials_json` injected per tool call via `POST /internal/v1/inject` on `datumbridge-integrations`.
- Fallback: `ODOO_BASE_URL`, `ODOO_DB`, `ODOO_USERNAME`, `ODOO_API_KEY` or `ODOO_PASSWORD` env vars (local dev only).
- `credentials_path` under `ODOO_CREDENTIALS_DIR` supported as a path-jailed alternative.

## Tool Categories

1. **Human Resources (HR)**: Employee directory, departments, manager hierarchy (`hr.employee`, `hr.department`).
2. **Time Off (Leaves)**: Leave requests, approvals/refusals, balance checking (`hr.leave`, `hr.leave.type`).
3. **Payroll & Compensation**: Payslips, salary contracts, wage details (`hr.payslip`, `hr.contract`).
4. **CRM & Customer Care**: Customer contacts, sales leads, communication activities (`res.partner`, `crm.lead`, `mail.activity`).
5. **Sales & Quotations**: Product catalog, quotes, sales orders, confirmations (`sale.order`, `product.product`).
6. **Universal Discovery**: Generic ORM execution & model schema inspection (`execute_kw`, `fields_get`).

## Security Baseline

- Never log secrets or tokens.
- Mask keys and credentials in logs and status outputs.
- Base URL and Database pinned from credentials bundle.
- 10 MiB cap on upstream response bodies.
