package cert

import (
    "net"
    "net/url"
    "strings"
)

// splitSANs parses SAN strings into DNS, IP, Email, and URI slices.
// Supported prefixes:
// - IP:1.2.3.4
// - email:user@example.com
// - uri:https://example.com
// Unprefixed values are treated as DNS names.
func splitSANs(sans []string) (dns []string, ips []net.IP, emails []string, uris []*url.URL) {
    for _, san := range sans {
        switch {
        case strings.HasPrefix(san, "IP:"):
            if ip := net.ParseIP(strings.TrimPrefix(san, "IP:")); ip != nil {
                ips = append(ips, ip)
            }
        case strings.HasPrefix(strings.ToLower(san), "email:"):
            emails = append(emails, san[len("email:"):])
        case strings.HasPrefix(strings.ToLower(san), "uri:"):
            if u, err := url.Parse(san[len("uri:"):]); err == nil {
                uris = append(uris, u)
            }
        default:
            dns = append(dns, san)
        }
    }
    return
}

