# odoo_create_partner

| Field | Value |
|---|---|
| Audience | internal |

## Description
Create a new contact, customer, or company in Odoo (res.partner).

## Target Model
`res.partner`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
