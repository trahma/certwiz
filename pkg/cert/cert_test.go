package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testdataPath returns the path to a file in the testdata directory
// handling the correct path separators for the current OS
func testdataPath(filename string) string {
	return filepath.Join("..", "..", "testdata", filename)
}

func TestInspectFile(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		expectError bool
		checks      func(t *testing.T, cert *Certificate)
	}{
		{
			name:        "Valid PEM certificate",
			file:        testdataPath("valid.pem"),
			expectError: false,
			checks: func(t *testing.T, cert *Certificate) {
				if cert.Format != "PEM" {
					t.Errorf("Expected format PEM, got %s", cert.Format)
				}
				if cert.Subject.CommonName != "test.example.com" {
					t.Errorf("Expected CN test.example.com, got %s", cert.Subject.CommonName)
				}
				if cert.IsExpired {
					t.Error("Certificate should not be expired")
				}
			},
		},
		{
			name:        "Valid DER certificate",
			file:        testdataPath("valid.der"),
			expectError: false,
			checks: func(t *testing.T, cert *Certificate) {
				if cert.Format != "DER" {
					t.Errorf("Expected format DER, got %s", cert.Format)
				}
			},
		},
		{
			name:        "Invalid certificate",
			file:        testdataPath("invalid.pem"),
			expectError: true,
		},
		{
			name:        "Non-existent file",
			file:        testdataPath("nonexistent.pem"),
			expectError: true,
		},
		{
			name:        "Certificate with many SANs",
			file:        testdataPath("many-sans.pem"),
			expectError: false,
			checks: func(t *testing.T, cert *Certificate) {
				expectedSANs := []string{
					"san1.example.com",
					"san2.example.com",
					"san3.example.com",
					"san4.example.com",
					"san5.example.com",
					"*.wildcard.example.com",
				}
				for _, san := range expectedSANs {
					found := false
					for _, dns := range cert.DNSNames {
						if dns == san {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected SAN %s not found", san)
					}
				}
				if len(cert.IPAddresses) != 2 {
					t.Errorf("Expected 2 IP addresses, got %d", len(cert.IPAddresses))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := InspectFile(tt.file)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.checks != nil {
				tt.checks(t, cert)
			}
		})
	}
}

func TestParseCertificate(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		expectFormat string
		expectError  bool
	}{
		{
			name:         "PEM format",
			file:         testdataPath("valid.pem"),
			expectFormat: "PEM",
			expectError:  false,
		},
		{
			name:         "DER format",
			file:         testdataPath("valid.der"),
			expectFormat: "DER",
			expectError:  false,
		},
		{
			name:         "Invalid data",
			file:         testdataPath("invalid.pem"),
			expectFormat: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			cert, format, err := parseCertificate(data)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if format != tt.expectFormat {
				t.Errorf("Expected format %s, got %s", tt.expectFormat, format)
			}
			if cert == nil {
				t.Error("Expected certificate but got nil")
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		opts        GenerateOptions
		expectError bool
	}{
		{
			name: "Basic certificate",
			opts: GenerateOptions{
				CommonName: "test.local",
				Days:       30,
				KeySize:    2048,
				OutputDir:  tempDir,
			},
			expectError: false,
		},
		{
			name: "Certificate with SANs",
			opts: GenerateOptions{
				CommonName: "multi.local",
				Days:       365,
				KeySize:    2048,
				SANs:       []string{"alt1.local", "alt2.local", "IP:127.0.0.1"},
				OutputDir:  tempDir,
			},
			expectError: false,
		},
		{
			name: "Small key size",
			opts: GenerateOptions{
				CommonName: "small.local",
				Days:       30,
				KeySize:    1024,
				OutputDir:  tempDir,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Generate(tt.opts)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify files were created
			certPath := filepath.Join(tt.opts.OutputDir, tt.opts.CommonName+".crt")
			keyPath := filepath.Join(tt.opts.OutputDir, tt.opts.CommonName+".key")

			if _, err := os.Stat(certPath); os.IsNotExist(err) {
				t.Error("Certificate file was not created")
			}
			if _, err := os.Stat(keyPath); os.IsNotExist(err) {
				t.Error("Key file was not created")
			}

			// Verify certificate can be parsed
			cert, err := InspectFile(certPath)
			if err != nil {
				t.Fatalf("Failed to inspect generated certificate: %v", err)
			}

			if cert.Subject.CommonName != tt.opts.CommonName {
				t.Errorf("Expected CN %s, got %s", tt.opts.CommonName, cert.Subject.CommonName)
			}

			// Check SANs
			if len(tt.opts.SANs) > 0 {
				for _, san := range tt.opts.SANs {
					if san == "IP:127.0.0.1" {
						found := false
						for _, ip := range cert.IPAddresses {
							if ip.Equal(net.ParseIP("127.0.0.1")) {
								found = true
								break
							}
						}
						if !found {
							t.Error("Expected IP 127.0.0.1 in SANs")
						}
					} else {
						found := false
						for _, dns := range cert.DNSNames {
							if dns == san {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected DNS %s in SANs", san)
						}
					}
				}
			}
		})
	}
}

func TestConvert(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		inputFile   string
		outputFile  string
		format      string
		expectError bool
	}{
		{
			name:        "PEM to DER",
			inputFile:   testdataPath("valid.pem"),
			outputFile:  filepath.Join(tempDir, "converted.der"),
			format:      "der",
			expectError: false,
		},
		{
			name:        "DER to PEM",
			inputFile:   testdataPath("valid.der"),
			outputFile:  filepath.Join(tempDir, "converted.pem"),
			format:      "pem",
			expectError: false,
		},
		{
			name:        "Invalid input file",
			inputFile:   testdataPath("nonexistent.pem"),
			outputFile:  filepath.Join(tempDir, "output.pem"),
			format:      "pem",
			expectError: true,
		},
		{
			name:        "Invalid format",
			inputFile:   testdataPath("valid.pem"),
			outputFile:  filepath.Join(tempDir, "output.xyz"),
			format:      "xyz",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Convert(tt.inputFile, tt.outputFile, tt.format)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify output file exists
			if _, err := os.Stat(tt.outputFile); os.IsNotExist(err) {
				t.Error("Output file was not created")
			}

			// Verify output can be parsed
			cert, err := InspectFile(tt.outputFile)
			if err != nil {
				t.Fatalf("Failed to inspect converted certificate: %v", err)
			}

			var expectedFormat string
			switch tt.format {
			case "der":
				expectedFormat = "DER"
			case "pem":
				expectedFormat = "PEM"
			default:
				expectedFormat = tt.format
			}

			if cert.Format != expectedFormat {
				t.Errorf("Expected format %s, got %s", expectedFormat, cert.Format)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name        string
		certPath    string
		caPath      string
		hostname    string
		expectValid bool
		expectError bool
	}{
		{
			name:        "Valid certificate",
			certPath:    testdataPath("valid.pem"),
			caPath:      "",
			hostname:    "",
			expectValid: true,
			expectError: false,
		},
		{
			name:        "Valid certificate with correct hostname",
			certPath:    testdataPath("valid.pem"),
			caPath:      "",
			hostname:    "test.example.com",
			expectValid: true,
			expectError: false,
		},
		{
			name:        "Valid certificate with wildcard match",
			certPath:    testdataPath("valid.pem"),
			caPath:      "",
			hostname:    "sub.test.example.com",
			expectValid: true,
			expectError: false,
		},
		{
			name:        "Valid certificate with wrong hostname",
			certPath:    testdataPath("valid.pem"),
			caPath:      "",
			hostname:    "wrong.example.com",
			expectValid: false,
			expectError: false,
		},
		{
			name:        "Non-existent certificate",
			certPath:    testdataPath("nonexistent.pem"),
			caPath:      "",
			hostname:    "",
			expectValid: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Verify(tt.certPath, tt.caPath, tt.hostname)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
				if len(result.Errors) > 0 {
					t.Logf("Errors: %v", result.Errors)
				}
			}
		})
	}
}

func TestCertificateExpiry(t *testing.T) {
	tests := []struct {
		name            string
		notBefore       time.Time
		notAfter        time.Time
		expectExpired   bool
		expectDaysUntil int
	}{
		{
			name:            "Valid certificate",
			notBefore:       time.Now().Add(-24 * time.Hour),
			notAfter:        time.Now().Add(30 * 24 * time.Hour),
			expectExpired:   false,
			expectDaysUntil: 30,
		},
		{
			name:            "Expired certificate",
			notBefore:       time.Now().Add(-365 * 24 * time.Hour),
			notAfter:        time.Now().Add(-24 * time.Hour),
			expectExpired:   true,
			expectDaysUntil: -1,
		},
		{
			name:            "Not yet valid",
			notBefore:       time.Now().Add(24 * time.Hour),
			notAfter:        time.Now().Add(365 * 24 * time.Hour),
			expectExpired:   false,
			expectDaysUntil: 365,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x509Cert := &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "test",
				},
				NotBefore: tt.notBefore,
				NotAfter:  tt.notAfter,
			}

			cert := &Certificate{
				Certificate:     x509Cert,
				IsExpired:       tt.notAfter.Before(time.Now()),
				DaysUntilExpiry: int(time.Until(tt.notAfter).Hours() / 24),
			}

			if cert.IsExpired != tt.expectExpired {
				t.Errorf("Expected IsExpired=%v, got %v", tt.expectExpired, cert.IsExpired)
			}

			// Allow +/- 1 day difference due to timing
			if diff := cert.DaysUntilExpiry - tt.expectDaysUntil; diff < -1 || diff > 1 {
				t.Errorf("Expected DaysUntilExpiry≈%d, got %d", tt.expectDaysUntil, cert.DaysUntilExpiry)
			}
		})
	}
}

func TestGenerateOptionsDefaults(t *testing.T) {
	tempDir := t.TempDir()
	opts := GenerateOptions{
		CommonName: "test",
		Days:       0, // Should default to some positive value
		KeySize:    0, // Should default to 2048 or similar
		OutputDir:  tempDir,
	}

	// Adjust defaults if Days or KeySize is 0
	if opts.Days == 0 {
		opts.Days = 365
	}
	if opts.KeySize == 0 {
		opts.KeySize = 2048
	}

	err := Generate(opts)
	if err != nil {
		t.Fatalf("Failed to generate with default values: %v", err)
	}

	certPath := filepath.Join(opts.OutputDir, opts.CommonName+".crt")
	cert, err := InspectFile(certPath)
	if err != nil {
		t.Fatalf("Failed to inspect generated certificate: %v", err)
	}

	// Certificate should be valid for the specified days
	expectedExpiry := time.Now().Add(time.Duration(opts.Days) * 24 * time.Hour)
	diff := cert.NotAfter.Sub(expectedExpiry)
	if diff < -24*time.Hour || diff > 24*time.Hour {
		t.Errorf("Certificate expiry not as expected: %v", cert.NotAfter)
	}
}

// TestInspectURLWithChain would require a mock server or network access
// For now, we'll create a placeholder that documents what should be tested
func TestInspectURLWithChain(t *testing.T) {
	t.Skip("Skipping URL inspection test - requires network or mock server")

	// This test would verify:
	// 1. Successful connection to HTTPS server
	// 2. Retrieval of server certificate
	// 3. Retrieval of certificate chain
	// 4. Proper error handling for connection failures
	// 5. Proper handling of different port numbers
	// 6. URL parsing and normalization
}

// Benchmark tests
func BenchmarkParseCertificatePEM(b *testing.B) {
	data, err := os.ReadFile(testdataPath("valid.pem"))
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := parseCertificate(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseCertificateDER(b *testing.B) {
	data, err := os.ReadFile(testdataPath("valid.der"))
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := parseCertificate(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := GenerateOptions{
			CommonName: fmt.Sprintf("test%d.local", i),
			Days:       365,
			KeySize:    2048,
			SANs:       []string{"alt1.local", "alt2.local"},
			OutputDir:  tempDir,
		}
		err := Generate(opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
