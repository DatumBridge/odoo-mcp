# odoo_list_quotations

| Field | Value |
|---|---|
| Audience | internal |

## Description
List quotations and sales orders (sale.order) with customer, state, and date filters.

## Target Model
`sale.order`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
