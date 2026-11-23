# odoo_create_activity

| Field | Value |
|---|---|
| Audience | internal |

## Description
Schedule or log a customer care activity (call, email, meeting, todo) on a record (mail.activity).

## Target Model
`mail.activity`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
