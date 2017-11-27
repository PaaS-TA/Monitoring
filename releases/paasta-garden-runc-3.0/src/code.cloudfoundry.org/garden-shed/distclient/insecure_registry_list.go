package distclient

import (
	"net"
	"strings"
)

type InsecureRegistryList []string

func (l InsecureRegistryList) AllowInsecure(host string) bool {
	return contains([]string(l), host)
}

func contains(list []string, element string) bool {
	for _, e := range list {
		if e == element {
			return true
		}

		if checkCIDR(e, element) {
			return true
		}
	}

	return false
}

func checkCIDR(entry, element string) bool {
	_, network, err := net.ParseCIDR(entry)
	if err != nil {
		return false
	}

	parts := strings.Split(element, ":")
	element = parts[0]

	ip := net.ParseIP(element)
	if network.Contains(ip) {
		return true
	}

	return false
}
