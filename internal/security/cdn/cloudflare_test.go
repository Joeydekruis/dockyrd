package cdn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCloudflareRanges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if stringsContains(r.URL.Path, "v4") {
			_, _ = w.Write([]byte("173.245.48.0/20\n"))
			return
		}
		_, _ = w.Write([]byte("2400:cb00::/32\n"))
	}))
	defer srv.Close()

	client := srv.Client()
	// Override URLs via synthetic handler paths is not possible without injection;
	// test fetchPlainLines instead.
	lines, err := fetchPlainLines(context.Background(), client, srv.URL+"/v4")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "173.245.48.0/20" {
		t.Fatalf("got %v", lines)
	}
}

func stringsContains(hay, needle string) bool {
	return len(hay) >= len(needle) && (hay == needle || len(needle) == 0 || indexOf(hay, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
