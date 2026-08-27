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
	for _, host := range hosts {
		fmt.Println(host)
	}
}
