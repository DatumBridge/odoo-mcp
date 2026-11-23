# odoo_search_products

| Field | Value |
|---|---|
| Audience | internal |

## Description
Search products in Odoo catalog (product.product) by name, SKU/default_code, barcode, or category.

## Target Model
`product.product`

## Usage Notes
- Uses credentials from injected vault bundle (`credentials_json`) or environment fallback.
- Never logs sensitive keys or secrets.
