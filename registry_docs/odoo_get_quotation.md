# odoo_get_quotation

| Field | Value |
|---|---|
| Audience | internal |

## Description
Get full quotation or sales order details by ID (sale.order), including order lines and prices.

## Target Model
`sale.order`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
