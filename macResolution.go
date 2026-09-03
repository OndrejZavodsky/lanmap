package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
			cleanedToken := strings.Trim(t, "()[],")

			if ip, err := netip.ParseAddr(cleanedToken); err == nil {
				foundIP = ip
				continue
			}

			if mac, err := net.ParseMAC(cleanedToken); err == nil {
				foundMac = mac
			}
		}

		if foundIP.IsValid() && foundMac != nil {
			macMap[foundIP] = foundMac
		}
	}

	return macMap, scanner.Err()
}
func AddMac(devices *[]Device) error {
	reader, err := GetARPTable()
	if err != nil {
		return err
	}

	macMap, err := ParseARP(reader)
	if err != nil {
		return err
	}

	for i := range *devices {
		device := &(*devices)[i]
		if device.Mac == nil {
			if hwAddr, found := macMap[device.IP]; found {
				device.Mac = hwAddr
			}
		}
	}

	return nil
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

func downloadOUIDB(client *http.Client) error {
	systemCacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to locate system cache directory: %w", err)
	}

	appCacheDir := filepath.Join(systemCacheDir, "lanmap")
	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create application cache directory: %w", err)
	}

	destinationFilePath := filepath.Join(appCacheDir, "oui.txt")

	response, err := client.Get("https://www.wireshark.org/download/automated/data/manuf")
	if err != nil {
		return fmt.Errorf("http get request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned unexpected status: %d %s", response.StatusCode, response.Status)
	}

	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}

	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, response.Body); err != nil {
		return fmt.Errorf("failed writing stream to disk: %w", err)
	}

	return nil
}

func parseOUIDatabase(source io.Reader) (map[string]string, error) {
	vendorMap := make(map[string]string, 35000)
	scanner := bufio.NewScanner(source)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		rawPrefix := fields[0]

		if maskIndex := strings.Index(rawPrefix, "/"); maskIndex != -1 {
			rawPrefix = rawPrefix[:maskIndex]
		}

		normalizedPrefix := strings.ToUpper(strings.NewReplacer("-", ":", ".", ":").Replace(rawPrefix))

		var vendorName string
		if len(fields) >= 3 {
			vendorName = strings.Join(fields[2:], " ")
		} else {
			vendorName = fields[1]
		}

		vendorMap[normalizedPrefix] = vendorName
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading Wireshark manuf stream: %w", err)
	}

	return vendorMap, nil
}

// AddVendorToDevice checks if the OUI database exists, downloads it if missing,
// parses it, and resolves the vendor for any device lacking one by matching
// the first half of its MAC address (the OUI prefix).
func AddVendorToDevice(devices *[]Device) error {
	if !ensureInstalledDB() {
		client := &http.Client{Timeout: 30 * time.Second}
		if err := downloadOUIDB(client); err != nil {
			return err
		}
	}

	systemCacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to locate cache directory: %w", err)
	}

	ouiFilePath := filepath.Join(systemCacheDir, "lanmap", "oui.txt")
	file, err := os.Open(ouiFilePath)
	if err != nil {
		return fmt.Errorf("failed to open OUI database: %w", err)
	}
	defer file.Close()

	ouiMap, err := parseOUIDatabase(file)
	if err != nil {
		return err
	}

	for i := range *devices {
		device := &(*devices)[i]
		if device.Vendor == "" && device.Mac != nil {
			macStr := device.Mac.String()

			if len(macStr) >= 8 {
				ouiPrefix := strings.ToUpper(macStr[:8])
				if vendor, exists := ouiMap[ouiPrefix]; exists {
					device.Vendor = vendor
				}
			}
		}
	}

	return nil
}
