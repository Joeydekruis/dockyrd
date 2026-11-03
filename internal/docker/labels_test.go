package docker

import "testing"

func TestIdentityLabelsRoundTrip(t *testing.T) {
	id := Identity{
		ServerID:     "01SERVER",
		ProjectID:    "01PROJECT",
		ServiceID:    "01SERVICE",
		Kind:         KindApp,
		DeploymentID: "01DEPLOY",
	}
	labels := id.Labels()
	parsed, ok := ParseIdentity(labels)
	if !ok {
		t.Fatal("expected managed identity")
	}
	if parsed != id {
		t.Fatalf("round trip: got %+v want %+v", parsed, id)
	}
}

func TestParseIdentityNotManaged(t *testing.T) {
	_, ok := ParseIdentity(map[string]string{"foo": "bar"})
	if ok {
		t.Fatal("expected false")
	}
}

func TestProjectNetworkName(t *testing.T) {
	if got := ProjectNetworkName("Client X"); got != "dy-p-client-x" {
		t.Fatalf("got %q", got)
	}
}

func TestAppContainerName(t *testing.T) {
	got := AppContainerName("api", DeployShort("01ARZ3NDEKTSV4RRFFQ69G5FAV"))
	if got != "dy-api-01arz3nd" {
		t.Fatalf("got %q", got)
	}
}
