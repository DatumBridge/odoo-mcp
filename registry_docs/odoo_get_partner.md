# odoo_get_partner

| Field | Value |
|---|---|
| Audience | internal |

## Description
Get detailed customer/contact profile by partner ID (res.partner).

## Target Model
`res.partner`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
