# odoo_get_leave_balances

| Field | Value |
|---|---|
| Audience | internal |

## Description
Get remaining leave balances and allocations (hr.leave.allocation) for an employee.

## Target Model
`hr.leave.allocation`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
