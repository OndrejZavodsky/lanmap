package lanmap

import (
	"net/netip"
	"sync"
	"testing"
)

func TestGenerateAddresses(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.1.0/30")
	got := GenerateAddresses(prefix)

	if len(got) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(got))
	}
	if got[0].String() != "192.168.1.1" {
		t.Errorf("expected first IP to be 192.168.1.1, got %s", got[0].String())
	}
}

func TestPingWorker_Unreachable(t *testing.T) {
	jobs := make(chan netip.Addr, 1)
	results := make(chan PingRes, 1)
	var wg sync.WaitGroup

	testIP := netip.MustParseAddr("192.0.2.1")

	wg.Add(1)
	go pingWorker(jobs, results, &wg)

	jobs <- testIP
	close(jobs)
	wg.Wait()
	close(results)

	res := <-results
	if res.Err == nil {
		t.Errorf("expected an error for an unreachable IP, got nil")
	}
	if res.Device.IP != (netip.Addr{}) {
		t.Errorf("expected empty device IP on error, got %s", res.Device.IP)
	}
}

func TestPingSweep_Empty(t *testing.T) {
	got := PingSweep([]netip.Addr{})
	if len(got) != 0 {
		t.Errorf("expected empty results for empty targets, got length %d", len(got))
	}
}
