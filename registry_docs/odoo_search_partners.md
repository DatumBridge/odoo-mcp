# odoo_search_partners

| Field | Value |
|---|---|
| Audience | internal |

## Description
Search customers, vendors, and contacts (res.partner) by name, email, phone, or company flag.

## Target Model
`res.partner`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
