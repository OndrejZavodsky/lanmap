package main

import (
	"net"
	"net/netip"
	"os"
	"sync"
	"time"
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
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), "80"), 500*time.Millisecond)
		if !os.IsTimeout(err) {
			results <- PingRes{
				Device: Device{IP: ip},
				Err:    nil,
			}
		}
		if err == nil {
			conn.Close()
			results <- PingRes{
				Device: Device{IP: ip},
				Err:    nil,
			}
		} else {
			results <- PingRes{
				Device: Device{},
				Err:    err,
			}
		}
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
