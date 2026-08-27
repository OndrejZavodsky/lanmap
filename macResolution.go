package lanmap

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/netip"
	"os/exec"
	"strings"
)

func GetARPTable() (io.Reader, error) {
	// Note: While 'arp -a' works on Windows/macOS, modern Linux and NixOS environments
	// often omit 'net-tools' by default in favor of 'iproute2' (e.g., 'ip neighbor').
	// 'arp -a' is fine for this milestone, but keep that in mind for broad Linux support.
	output, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(output), nil
}
func ParseARP() (map[netip.Addr]net.HardwareAddr, error) {
	arpBytes, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil, err
	}
	macMap := make(map[netip.Addr]net.HardwareAddr)
	scanner := bufio.NewScanner(bytes.NewReader(arpBytes))
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		var foundIp netip.Addr
		var foundMac net.HardwareAddr
		for _, t := range tokens {
			ip, err := netip.ParseAddr(t)
			if err != nil {
				mac, err := net.ParseMAC(t)
				if err != nil {
					continue
				}
				foundMac = mac
			}
			foundIp = ip
		}
		if foundIp.IsValid() && foundMac != nil {
			macMap[foundIp] = foundMac

		}
	}
	return macMap, nil
}
