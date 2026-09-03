package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseARP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMap int
		wantIP  string
		wantMAC string
	}{
		{
			name:    "Windows Format",
			input:   "  192.168.1.1           00-11-22-33-44-55     dynamic\n",
			wantMap: 1,
			wantIP:  "192.168.1.1",
			wantMAC: "00:11:22:33:44:55",
		},
		{
			name:    "Linux Format",
			input:   "192.168.1.1 dev eth0 lladdr 00:11:22:33:44:55 REACHABLE\n",
			wantMap: 1,
			wantIP:  "192.168.1.1",
			wantMAC: "00:11:22:33:44:55",
		},
		{
			name:    "Invalid or Empty",
			input:   "Interface: 192.168.1.254 --- 0x2\n  Internet Address      Physical Address      Type\n",
			wantMap: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := ParseARP(r)

			if err != nil {
				t.Fatalf("ParseARP returned unexpected error: %v", err)
			}
			if len(got) != tt.wantMap {
				t.Fatalf("expected map length %d, got %d", tt.wantMap, len(got))
			}

			if tt.wantMap > 0 {
				ip := netip.MustParseAddr(tt.wantIP)
				mac, exists := got[ip]
				if !exists {
					t.Fatalf("expected IP %s not found in map", tt.wantIP)
				}
				if mac.String() != tt.wantMAC {
					t.Errorf("expected MAC %s, got %s", tt.wantMAC, mac.String())
				}
			}
		})
	}
}

type urlRewriteTransport struct {
	BaseTransport http.RoundTripper
	TestServerURL string
}

func (u *urlRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	parsedTestURL, err := url.Parse(u.TestServerURL)
	if err != nil {
		return nil, err
	}

	modifiedReq := req.Clone(req.Context())
	modifiedReq.URL.Scheme = parsedTestURL.Scheme
	modifiedReq.URL.Host = parsedTestURL.Host

	return u.BaseTransport.RoundTrip(modifiedReq)
}

func TestDownloadOUIDB(t *testing.T) {
	isolatedCacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", isolatedCacheDir)
	t.Setenv("HOME", isolatedCacheDir)

	mockOUIData := []byte("00-00-00   (hex)\t\tXEROX CORPORATION")
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(mockOUIData)
	}))
	defer mockServer.Close()

	interceptClient := mockServer.Client()
	interceptClient.Transport = &urlRewriteTransport{
		BaseTransport: interceptClient.Transport,
		TestServerURL: mockServer.URL,
	}

	err := downloadOUIDB(interceptClient)
	if err != nil {
		t.Fatalf("DownloadOUIDB failed unexpectedly: %v", err)
	}

	expectedDestination := filepath.Join(isolatedCacheDir, "lanmap", "oui.txt")

	savedData, err := os.ReadFile(expectedDestination)
	if err != nil {
		t.Fatalf("Failed to read the generated OUI file: %v", err)
	}

	if !bytes.Equal(savedData, mockOUIData) {
		t.Errorf("File contents did not match. Got %q, wanted %q", savedData, mockOUIData)
	}
}
