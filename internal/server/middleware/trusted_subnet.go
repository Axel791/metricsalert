package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware проверяет, что X-Real-IP входит в trusted CIDR.
// Если CIDR пустой, вызывайте next напрямую — поэтому мидлвару следует
// подключать только когда значение в конфиге непустое.
func TrustedSubnetMiddleware(trustedCIDR string) func(http.Handler) http.Handler {
	_, ipNet, err := net.ParseCIDR(trustedCIDR)
	if err != nil {
		panic("invalid trusted_subnet CIDR: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}
			ip := net.ParseIP(ipStr)
			if ip == nil || !ipNet.Contains(ip) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
