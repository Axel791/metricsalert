package validatiors

import (
	"net"
	"strings"
)

// IsValidAddress проверяет валидность адреса.
// Если withScheme == true, учитывается схема (http:// или https://).
func IsValidAddress(addr string, withScheme bool) bool {
	if withScheme {
		addr = strings.TrimPrefix(addr, "http://")
		addr = strings.TrimPrefix(addr, "https://")
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	if _, err := net.LookupPort("tcp", port); err != nil {
		return false
	}
	if host == "" {
		return false
	}

	if _, err := net.LookupHost(host); err != nil {
		ip := net.ParseIP(host)
		if ip == nil {
			return false
		}
	}

	return true
}
