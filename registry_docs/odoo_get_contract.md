# odoo_get_contract

| Field | Value |
|---|---|
| Audience | internal |

## Description
Get detailed contract terms and compensation details by contract ID.

## Target Model
`hr.contract`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
