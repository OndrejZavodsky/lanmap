package lanmap

import (
	"net"
	"net/netip"
)

type Network struct {
	Network net.IPNet
	Devices []Device
}
type Device struct {
	IP         netip.Addr
	Name       string
	Vendor     string
	Neighbours []netip.Addr
	Mac        net.HardwareAddr
}
