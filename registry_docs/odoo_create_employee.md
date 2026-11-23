# odoo_create_employee

| Field | Value |
|---|---|
| Audience | internal |

## Description
Create a new employee record in Odoo with department and job assignment.

## Target Model
`hr.employee`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
