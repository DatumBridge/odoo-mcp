# odoo_list_departments

| Field | Value |
|---|---|
| Audience | internal |

## Description
List Odoo organizational departments (hr.department) with hierarchy and manager details.

## Target Model
`hr.department`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
