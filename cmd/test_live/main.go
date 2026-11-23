package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/datumbridge/odoo-mcp/internal/tools"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_, handlers := tools.Register()

	credsJSON := os.Getenv("ODOO_CREDENTIALS_JSON")

	// Test odoo_search_partners
	partnerPayload, _ := json.Marshal(map[string]interface{}{
		"credentials_json": credsJSON,
		"limit":            5,
	})
	partnerRes := handlers["odoo_search_partners"](partnerPayload)
	pb, _ := json.MarshalIndent(partnerRes, "", "  ")
	fmt.Printf("=== odoo_search_partners ===\n%s\n\n", string(pb))

	// Test odoo_list_crm_leads
	listPayload, _ := json.Marshal(map[string]interface{}{
		"credentials_json": credsJSON,
		"limit":            5,
	})
	listRes := handlers["odoo_list_crm_leads"](listPayload)
	listB, _ := json.MarshalIndent(listRes, "", "  ")
	fmt.Printf("=== odoo_list_crm_leads ===\n%s\n\n", string(listB))
}
