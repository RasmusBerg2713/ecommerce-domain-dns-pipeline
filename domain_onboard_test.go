package main

import "testing"

func TestRootDomain(t *testing.T) {
	got := rootDomain("mail.receipts.shop.example")
	if got != "shop.example" {
		t.Fatalf("rootDomain returned %q", got)
	}
}
