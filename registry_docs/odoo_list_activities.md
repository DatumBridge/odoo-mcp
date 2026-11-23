# odoo_list_activities

| Field | Value |
|---|---|
| Audience | internal |

## Description
List scheduled customer activities (mail.activity) on records or assigned to users.

## Target Model
`mail.activity`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
