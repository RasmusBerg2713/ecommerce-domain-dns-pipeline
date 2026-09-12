package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type dnsRow struct {
	Stage  string `json:"stage"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
}

func main() {
	domain := os.Getenv("SENDING_DOMAIN")
	if domain == "" {
		fmt.Fprintln(os.Stderr, "set SENDING_DOMAIN, for example mail.shop.example")
		os.Exit(2)
	}

	setup, err := infrai.email.domain.verify(domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, record := range setup.DNSRecords {
		emit(dnsRow{Stage: "publish", Type: record.Type, Name: record.Name, Value: record.Value, Domain: domain})
	}
	emit(dnsRow{
		Stage:  "publish",
		Type:   "TXT",
		Name:   "_dmarc." + domain,
		Value:  "v=DMARC1; p=quarantine; rua=mailto:dmarc@" + rootDomain(domain),
		Domain: domain,
	})

	current, err := infrai.email.domain.get(domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	emit(dnsRow{Stage: "verification", Type: "status", Name: domain, Value: current.Verification.Status, Domain: domain})
}

func emit(row dnsRow) {
	encoded, err := json.Marshal(row)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))
}

func rootDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	return strings.Join(parts[len(parts)-2:], ".")
}
