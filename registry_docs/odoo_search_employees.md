# odoo_search_employees

| Field | Value |
|---|---|
| Audience | internal |

## Description
Search and list Odoo employees with filters by name, email, department, job, manager, and pagination.

## Target Model
`hr.employee`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
