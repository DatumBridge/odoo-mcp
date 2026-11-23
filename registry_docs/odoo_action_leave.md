# odoo_action_leave

| Field | Value |
|---|---|
| Audience | internal |

## Description
Approve, validate, or refuse a leave request (hr.leave).

## Target Model
`hr.leave`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
