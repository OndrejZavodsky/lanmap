package lanmap

import (
	"net/netip"
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
