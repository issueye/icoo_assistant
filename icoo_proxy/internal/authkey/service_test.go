package authkey

import (
	"reflect"
	"testing"
)

func TestServiceUpsertListsAndDeletesAuthKeys(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })

	record, err := svc.Upsert(UpsertInput{
		Name:        "Local Client",
		Secret:      "icoo_test_secret",
		Enabled:     true,
		Description: "local dev",
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if record.SecretMasked == "" || record.SecretMasked == "icoo_test_secret" {
		t.Fatalf("expected masked secret, got %q", record.SecretMasked)
	}
	if got := svc.EnabledSecrets(); len(got) != 1 || got[0] != "icoo_test_secret" {
		t.Fatalf("unexpected enabled secrets: %#v", got)
	}

	if _, err := svc.Upsert(UpsertInput{
		ID:      record.ID,
		Name:    "Local Client",
		Enabled: false,
	}); err != nil {
		t.Fatalf("update without secret: %v", err)
	}
	if got := svc.EnabledSecrets(); len(got) != 0 {
		t.Fatalf("expected disabled key to be excluded, got %#v", got)
	}
	if err := svc.Delete(record.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got := svc.List(); len(got) != 0 {
		t.Fatalf("expected empty list, got %#v", got)
	}
}

func TestMergeSecretsDeduplicatesCommaSeparatedValues(t *testing.T) {
	got := MergeSecrets([]string{"alpha"}, []string{"beta,gamma", "alpha", " gamma "})
	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
