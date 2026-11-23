# odoo_action_quotation

| Field | Value |
|---|---|
| Audience | internal |

## Description
Perform workflow action on quotation/sales order: confirm into sales order or cancel.

## Target Model
`sale.order`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
