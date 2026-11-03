package storage

import (
	"testing"
)

func TestValidateInstallRequest(t *testing.T) {
	if err := ValidateInstallRequest("garage"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateInstallRequest("minio"); err == nil {
		t.Fatal("expected error for minio")
	}
}
