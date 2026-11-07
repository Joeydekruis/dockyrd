package updater

import "testing"

func TestCompareVersions(t *testing.T) {
	if CompareVersions("0.0.14", "v0.0.15") >= 0 {
		t.Fatal("expected older")
	}
	if CompareVersions("1.0.0", "0.9.9") <= 0 {
		t.Fatal("expected newer major")
	}
	if CompareVersions("0.0.14", "0.0.14") != 0 {
		t.Fatal("expected equal")
	}
}

func TestParseChecksumLine(t *testing.T) {
	body := "abc123  dockyrd_0.0.14_linux_amd64.tar.gz\n"
	got := parseChecksumLine(body, "dockyrd_0.0.14_linux_amd64.tar.gz")
	if got != "abc123" {
		t.Fatalf("got %q", got)
	}
}
