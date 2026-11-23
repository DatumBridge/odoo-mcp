# odoo_describe_model

| Field | Value |
|---|---|
| Audience | internal |

## Description
Inspect field schemas, types, and descriptions of any Odoo model via fields_get.

## Target Model
`universal`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
