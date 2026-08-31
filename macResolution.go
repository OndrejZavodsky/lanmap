package main

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
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
		var foundIP netip.Addr
		var foundMac net.HardwareAddr

		for _, t := range tokens {
			if ip, err := netip.ParseAddr(t); err == nil {
				foundIP = ip
				continue
			}

			if mac, err := net.ParseMAC(t); err == nil {
				foundMac = mac
			}
		}

		if foundIP.IsValid() && foundMac != nil {
			macMap[foundIP] = foundMac
		}
	}

	return macMap, scanner.Err()
}

func ensureInstalledDB() bool {
	cashDir, err := os.UserCacheDir()
	if err != nil {
		return false
	}
	path := filepath.Join(cashDir, "lanmap/oui.txt")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
