# odoo_list_leaves

| Field | Value |
|---|---|
| Audience | internal |

## Description
List time off and leave requests (hr.leave) with filters by employee, state, and date range.

## Target Model
`hr.leave`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
