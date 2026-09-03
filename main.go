package main

import (
	"fmt"
	"net/netip"
)

func main() {
	prefix, err := netip.ParsePrefix("10.100.1.0/24")
	if err != nil {
		return
	}

	hosts := PingSweep(GenerateAddresses(prefix))
	if err := AddMac(&hosts); err != nil {
		fmt.Println("error: %W", err)
	}
	if err := AddVendorToDevice(&hosts); err != nil {
		fmt.Println("errof: %w", err)
		return
	}
	for _, host := range hosts {
		fmt.Println(host)
	}
}
