package main

import (
	"strings"
	"testing"

	"icoo_proxy/internal/routepolicy"
	"icoo_proxy/internal/supplier"
)

func TestDeleteSupplierRejectsEnabledPolicyReference(t *testing.T) {
	root := t.TempDir()
	suppliers, err := supplier.NewService(root)
	if err != nil {
		t.Fatalf("new suppliers: %v", err)
	}
	t.Cleanup(func() { _ = suppliers.Close() })

	record, err := suppliers.Upsert(supplier.UpsertInput{
		Name:         "Policy Vendor",
		Protocol:     "openai-responses",
		BaseURL:      "https://example.com",
		Enabled:      true,
		Models:       "gpt-4.1-mini",
		DefaultModel: "gpt-4.1-mini",
		UserAgent:    "PolicyVendor/1.0",
	})
	if err != nil {
		t.Fatalf("upsert supplier: %v", err)
	}

	policies, err := routepolicy.NewService(root, suppliers)
	if err != nil {
		t.Fatalf("new policies: %v", err)
	}
	t.Cleanup(func() { _ = policies.Close() })

	if _, err := policies.Upsert(routepolicy.UpsertInput{
		DownstreamProtocol: "openai-chat",
		SupplierID:         record.ID,
		Enabled:            true,
	}); err != nil {
		t.Fatalf("upsert policy: %v", err)
	}

	app := &App{
		root:      root,
		suppliers: suppliers,
		policies:  policies,
	}

	_, err = app.DeleteSupplier(record.ID)
	if err == nil {
		t.Fatalf("expected delete supplier to be blocked")
	}
	if !strings.Contains(err.Error(), "supplier is used by enabled route policy \"openai-chat\"") {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if _, ok := suppliers.Resolve(record.ID); !ok {
		t.Fatalf("expected blocked supplier deletion to keep supplier record")
	}
}
