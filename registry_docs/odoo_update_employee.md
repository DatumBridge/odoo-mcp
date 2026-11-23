# odoo_update_employee

| Field | Value |
|---|---|
| Audience | internal |

## Description
Update employee details, contact info, department, or manager.

## Target Model
`hr.employee`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
