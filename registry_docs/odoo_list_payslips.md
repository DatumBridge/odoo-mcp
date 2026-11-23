# odoo_list_payslips

| Field | Value |
|---|---|
| Audience | internal |

## Description
List employee payslips (hr.payslip) with status and date filters.

## Target Model
`hr.payslip`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
