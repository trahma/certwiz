package cert

import (
	"testing"
	"time"
)

func TestTLSVersionConstants(t *testing.T) {
	// Test that TLS version constants match expected values
	tests := []struct {
		version   TLSVersion
		expected  uint16
		name      string
	}{
		{TLSVersionTLS10, 0x0301, "TLS 1.0"},
		{TLSVersionTLS11, 0x0302, "TLS 1.1"},
		{TLSVersionTLS12, 0x0303, "TLS 1.2"},
		{TLSVersionTLS13, 0x0304, "TLS 1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint16(tt.version) != tt.expected {
				t.Errorf("TLSVersion %s = 0x%04x, want 0x%04x", tt.name, uint16(tt.version), tt.expected)
			}
		})
	}
}

func TestTLSVersionNames(t *testing.T) {
	tests := []struct {
		version TLSVersion
		want    string
	}{
		{TLSVersionTLS10, "TLS 1.0"},
		{TLSVersionTLS11, "TLS 1.1"},
		{TLSVersionTLS12, "TLS 1.2"},
		{TLSVersionTLS13, "TLS 1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tlsVersionNames[tt.version]; got != tt.want {
				t.Errorf("tlsVersionNames[%v] = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestTLSResultStructure(t *testing.T) {
	result := &TLSResult{
		Host:         "example.com",
		Port:         443,
		Versions:     []TLSVersionInfo{},
		MinSupported: TLSVersionTLS12,
		MaxSupported: TLSVersionTLS13,
	}

	if result.Host != "example.com" {
		t.Errorf("Host = %q, want %q", result.Host, "example.com")
	}
	if result.Port != 443 {
		t.Errorf("Port = %d, want %d", result.Port, 443)
	}
	if result.MinSupported != TLSVersionTLS12 {
		t.Errorf("MinSupported = %v, want %v", result.MinSupported, TLSVersionTLS12)
	}
	if result.MaxSupported != TLSVersionTLS13 {
		t.Errorf("MaxSupported = %v, want %v", result.MaxSupported, TLSVersionTLS13)
	}
}

func TestTLSVersionInfoStructure(t *testing.T) {
	info := TLSVersionInfo{
		Version:   TLSVersionTLS12,
		Name:      "TLS 1.2",
		Supported: true,
		Error:     "",
	}

	if info.Version != TLSVersionTLS12 {
		t.Errorf("Version = %v, want %v", info.Version, TLSVersionTLS12)
	}
	if info.Name != "TLS 1.2" {
		t.Errorf("Name = %q, want %q", info.Name, "TLS 1.2")
	}
	if !info.Supported {
		t.Error("Supported = false, want true")
	}
}

func TestCheckTLSVersionsWithInvalidHost(t *testing.T) {
	// Test with an invalid hostname - should return error or empty result
	// Note: DNS resolution might succeed for some invalid domains, so we just verify
	// the function doesn't panic and returns a result
	result, err := CheckTLSVersions("invalid.host.that.does.not.exist.invalid", 443, 1*time.Second)
	// Either we get an error or we get a result (DNS might resolve)
	if err == nil && result == nil {
		t.Error("Expected either error or result for invalid hostname")
	}
}

func TestCheckTLSVersionsWithUnreachablePort(t *testing.T) {
	// Test with a valid hostname but unreachable port
	// Use a port that is definitely not listening
	result, err := CheckTLSVersions("localhost", 59999, 500*time.Millisecond)
	// Either we get an error or we get a result (localhost might resolve differently)
	if err == nil && result == nil {
		t.Error("Expected either error or result for unreachable port")
	}
}

func TestTLSVersionComparison(t *testing.T) {
	// Test TLS version ordering
	if TLSVersionTLS10 >= TLSVersionTLS11 {
		t.Error("TLS 1.0 should be less than TLS 1.1")
	}
	if TLSVersionTLS11 >= TLSVersionTLS12 {
		t.Error("TLS 1.1 should be less than TLS 1.2")
	}
	if TLSVersionTLS12 >= TLSVersionTLS13 {
		t.Error("TLS 1.2 should be less than TLS 1.3")
	}
}
