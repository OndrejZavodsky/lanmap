package lanmap

import (
	"net/netip"
	"sync"
)

func GenerateAddresses(prefix netip.Prefix) []netip.Addr {
	network := prefix.Masked().Addr()
	current := network.Next()
	var generated []netip.Addr
	for prefix.Contains(current) && prefix.Contains(current.Next()) {
		generated = append(generated, current)
		current = current.Next()
	}
	return generated
}

type PingRes struct {
	Device Device
	Err    error
}

func pingWorker(
	jobs <-chan netip.Addr,
	results chan<- PingRes,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for ip := range jobs {
		//ping the device
		// 2. Send PingResult{Device: ..., Err: ...} to 'results' channel
	}
}

func PingSweep(targets []netip.Addr) []Device {
	leng := len(targets)

	jobs := make(chan netip.Addr, leng)
	results := make(chan PingRes, leng)

	var wg sync.WaitGroup
	workerCount := leng / 64
	if workerCount < 1 {
		workerCount = 1
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go pingWorker(jobs, results, &wg)
	}

	for _, ip := range targets {
		jobs <- ip
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()

	var discovered []Device
	for res := range results {
		if res.Err == nil {
			discovered = append(discovered, res.Device)
		}
	}

	return discovered
}
