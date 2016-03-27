package docker

import (
	"net"
	"strings"
)

// Note that this depends on the context in which it is run.
// If this is run from the host (outside container), then it will return the address at eth0,
// but if it's run from inside a container, the eth0 interface is actually the docker0 interface
// on the host.
func GetEth0Ip() ([]string, error) {
	ips := []string{}
	intf, err := net.InterfaceByName("eth0")
	if err != nil {
		if !strings.Contains(err.Error(), "no such network interface") {
			return ips, err
		} else {
			return nil, nil
		}
	}

	addrs, err := intf.Addrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		// parse the ip in CIDR form
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip.String())
	}
	return ips, nil
}
