# odoo_get_payslip

| Field | Value |
|---|---|
| Audience | internal |

## Description
Get detailed payslip record including salary breakdown rules (BASIC, GROSS, NET, deductions).

## Target Model
`hr.payslip`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
