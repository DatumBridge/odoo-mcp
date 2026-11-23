# odoo_create_quotation

| Field | Value |
|---|---|
| Audience | internal |

## Description
Create a new quotation / sales order (sale.order) with line items in Odoo.

## Target Model
`sale.order`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
