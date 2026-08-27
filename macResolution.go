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
	output, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(output), nil
}

func ParseARP(r io.Reader) (map[netip.Addr]net.HardwareAddr, error) {
	macMap := make(map[netip.Addr]net.HardwareAddr)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		tokens := strings.Fields(scanner.Text())
		var foundIp netip.Addr
		var foundMac net.HardwareAddr

		for _, t := range tokens {
			if ip, err := netip.ParseAddr(t); err == nil {
				foundIp = ip
				continue
			}

			if mac, err := net.ParseMAC(t); err == nil {
				foundMac = mac
			}
		}

		if foundIp.IsValid() && foundMac != nil {
			macMap[foundIp] = foundMac
		}
	}

	return macMap, scanner.Err()
}
