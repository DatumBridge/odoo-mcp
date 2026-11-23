# odoo_execute_kw

| Field | Value |
|---|---|
| Audience | internal |

## Description
Execute any arbitrary Odoo ORM method on any model via JSON-RPC execute_kw.

## Target Model
`universal`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
