package cert

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultDialTimeout = 5 * time.Second

const (
	// Format constants
	FormatPEM = "PEM"
	FormatDER = "DER"
)

// Certificate represents a parsed X.509 certificate with additional metadata
type Certificate struct {
	*x509.Certificate
	Source          string // file path or URL
	Format          string // PEM or DER
	IsExpired       bool
	DaysUntilExpiry int
}

// InspectFile reads and parses a certificate file
func InspectFile(filepath string) (*Certificate, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	cert, format, err := parseCertificate(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &Certificate{
		Certificate:     cert,
		Source:          filepath,
		Format:          format,
		IsExpired:       cert.NotAfter.Before(time.Now()),
		DaysUntilExpiry: int(time.Until(cert.NotAfter).Hours() / 24),
	}, nil
}

// InspectURL connects to a URL and retrieves its certificate
func InspectURL(targetURL string, port int) (*Certificate, error) {
	cert, _, err := InspectURLWithChain(targetURL, port)
	return cert, err
}

// InspectURLWithChain connects to a URL and retrieves its certificate and chain
func InspectURLWithChain(targetURL string, port int) (*Certificate, []*Certificate, error) {
	return InspectURLWithConnect(targetURL, port, "")
}

// InspectURLWithConnect connects to a specific host but validates the certificate for the target URL
// This is useful for testing certificates through proxies, tunnels, or local services
// If connectHost is empty, it connects directly to the target
// InspectURLWithConnect uses a default timeout for the TLS connection.
func InspectURLWithConnect(targetURL string, port int, connectHost string) (*Certificate, []*Certificate, error) {
    return InspectURLWithConnectTimeout(targetURL, port, connectHost, defaultDialTimeout)
}

// InspectURLWithConnectTimeout connects with a specific timeout.
func InspectURLWithConnectTimeout(targetURL string, port int, connectHost string, timeout time.Duration) (*Certificate, []*Certificate, error) {
    return InspectURLWithOptions(targetURL, port, connectHost, timeout, "auto")
}

// InspectURLWithOptions connects with a specific timeout and signature algorithm preference.
// sigAlg can be "auto", "ecdsa", or "rsa" to control cipher suite selection.
func InspectURLWithOptions(targetURL string, port int, connectHost string, timeout time.Duration, sigAlg string) (*Certificate, []*Certificate, error) {
    // Parse and normalize URL
    if !strings.Contains(targetURL, "://") {
        targetURL = "https://" + targetURL
    }

	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Determine the actual host to connect to
	var dialHost string
	serverName := u.Hostname() // Always use the target hostname for TLS verification
	
	if connectHost != "" {
		// Use the provided connect host
		dialHost = fmt.Sprintf("%s:%d", connectHost, port)
	} else {
		// Connect directly to the target
		host := u.Hostname()
		if u.Port() != "" {
			dialHost = net.JoinHostPort(u.Hostname(), u.Port())
		} else {
			dialHost = fmt.Sprintf("%s:%d", host, port)
		}
	}

    // Configure TLS with cipher suite preferences based on signature algorithm
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true, // We want to inspect even invalid certs
        ServerName:         serverName, // Use the target hostname for SNI
    }
    
    // Set cipher suites based on signature algorithm preference
    switch strings.ToLower(sigAlg) {
    case "ecdsa":
        // Only ECDSA cipher suites - server will be forced to use ECDSA cert if available
        tlsConfig.CipherSuites = []uint16{
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
        }
        // Ensure we don't negotiate TLS 1.3 where cipher suites don't control cert selection
        tlsConfig.MaxVersion = tls.VersionTLS12
    case "rsa":
        // Only RSA cipher suites - server will be forced to use RSA cert if available
        tlsConfig.CipherSuites = []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
            tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
            tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
        }
        // Ensure we don't negotiate TLS 1.3 where cipher suites don't control cert selection
        tlsConfig.MaxVersion = tls.VersionTLS12
    default:
        // "auto" or any other value - use default cipher suites
        // Let Go choose the best cipher suites
    }
    
    // Connect with TLS using a timeout to avoid hanging
    dialer := &net.Dialer{Timeout: timeout}
    conn, err := tls.DialWithDialer(dialer, "tcp", dialHost, tlsConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Get the peer certificates
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, nil, fmt.Errorf("no certificates found")
	}

	// First certificate is the server certificate
	serverCert := &Certificate{
		Certificate:     certs[0],
		Source:          u.String(),
		Format:          FormatDER,
		IsExpired:       certs[0].NotAfter.Before(time.Now()),
		DaysUntilExpiry: int(time.Until(certs[0].NotAfter).Hours() / 24),
	}

	// Build chain from remaining certificates
	var chain []*Certificate
	for i := 1; i < len(certs); i++ {
		chainCert := &Certificate{
			Certificate:     certs[i],
			Source:          fmt.Sprintf("Chain[%d]", i),
			Format:          FormatDER,
			IsExpired:       certs[i].NotAfter.Before(time.Now()),
			DaysUntilExpiry: int(time.Until(certs[i].NotAfter).Hours() / 24),
		}
		chain = append(chain, chainCert)
	}

	return serverCert, chain, nil
}

// Generate creates a new self-signed certificate
func Generate(opts GenerateOptions) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, opts.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: opts.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, opts.Days),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

    // Add Subject Alternative Names (DNS, IP)
    if len(opts.SANs) > 0 {
        dns, ips, _, _ := splitSANs(opts.SANs)
        template.DNSNames = append(template.DNSNames, dns...)
        template.IPAddresses = append(template.IPAddresses, ips...)
    }

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write certificate file
	certPath := filepath.Join(opts.OutputDir, opts.CommonName+".crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer func() { _ = certFile.Close() }()

	if err := pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key file
	keyPath := filepath.Join(opts.OutputDir, opts.CommonName+".key")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() { _ = keyFile.Close() }()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Set restrictive permissions on the private key (Unix-like systems only)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(keyPath, 0600); err != nil {
			return fmt.Errorf("failed to set key permissions: %w", err)
		}
	}

	return nil
}

// Convert changes certificate format
func Convert(inputPath, outputPath, format string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	cert, _, err := parseCertificate(data)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	var output []byte

	switch strings.ToLower(format) {
	case "pem":
		output = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
	case "der":
		output = cert.Raw
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// Verify checks certificate validity and hostname matching
func Verify(certPath, caPath, hostname string) (*VerificationResult, error) {
    cert, err := InspectFile(certPath)
    if err != nil {
        return nil, err
    }

	result := &VerificationResult{
		Certificate: cert,
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
	}

	// Check expiration
	now := time.Now()
	if cert.NotBefore.After(now) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Certificate is not yet valid")
	} else if cert.NotAfter.Before(now) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Certificate has expired")
	}

    // Check hostname if provided
    if hostname != "" {
        if err := cert.VerifyHostname(hostname); err != nil {
            result.IsValid = false
            result.Errors = append(result.Errors, fmt.Sprintf("Hostname verification failed: %v", err))
        }
    }

    // CA chain verification when a CA bundle/path is provided
    if caPath != "" {
        caData, err := os.ReadFile(caPath)
        if err != nil {
            return nil, fmt.Errorf("failed to read CA file: %w", err)
        }

        roots := x509.NewCertPool()
        // Try PEM first
        if ok := roots.AppendCertsFromPEM(caData); !ok {
            // Fallback: try single DER certificate
            if caCert, err := x509.ParseCertificate(caData); err == nil {
                roots.AddCert(caCert)
            } else {
                return nil, fmt.Errorf("failed to parse CA certificate(s)")
            }
        }

        opts := x509.VerifyOptions{Roots: roots}
        if hostname != "" {
            opts.DNSName = hostname
        }

        if _, err := cert.Certificate.Verify(opts); err != nil {
            result.IsValid = false
            result.Errors = append(result.Errors, fmt.Sprintf("Chain verification failed: %v", err))
        }
    }

    return result, nil
}

// parseCertificate tries to parse certificate data as PEM or DER
func parseCertificate(data []byte) (*x509.Certificate, string, error) {
	// Try PEM first
	if block, _ := pem.Decode(data); block != nil {
		cert, err := x509.ParseCertificate(block.Bytes)
		return cert, FormatPEM, err
	}

	// Try DER
	cert, err := x509.ParseCertificate(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse as PEM or DER: %w", err)
	}

	return cert, FormatDER, nil
}

// getPublicKeyAlgorithm returns the algorithm name for a public key
func getPublicKeyAlgorithm(pubKey interface{}) string {
	switch pubKey.(type) {
	case *rsa.PublicKey:
		return "RSA"
	case *ecdsa.PublicKey:
		return "ECDSA"
	default:
		return "Unknown"
	}
}

// getPublicKeySize returns the size of a public key in bits
func getPublicKeySize(pubKey interface{}) int {
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		return key.N.BitLen()
	case *ecdsa.PublicKey:
		return key.Params().BitSize
	default:
		return 0
	}
}

// GenerateCSR generates a Certificate Signing Request
func GenerateCSR(options CSROptions, csrPath, keyPath string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, options.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare subject
	subject := pkix.Name{
		CommonName: options.CommonName,
	}

	if options.Organization != "" {
		subject.Organization = []string{options.Organization}
	}
	if options.OrganizationalUnit != "" {
		subject.OrganizationalUnit = []string{options.OrganizationalUnit}
	}
	if options.Country != "" {
		subject.Country = []string{options.Country}
	}
	if options.Province != "" {
		subject.Province = []string{options.Province}
	}
	if options.Locality != "" {
		subject.Locality = []string{options.Locality}
	}

	// Prepare CSR template
	template := x509.CertificateRequest{
		Subject: subject,
	}

	// Add email if provided
	if options.EmailAddress != "" {
		template.EmailAddresses = []string{options.EmailAddress}
	}

    // Process SANs (DNS, IP, optional email/URI)
    dns, ips, emails, uris := splitSANs(options.SANs)
    template.DNSNames = append(template.DNSNames, dns...)
    template.IPAddresses = append(template.IPAddresses, ips...)
    if len(emails) > 0 {
        template.EmailAddresses = append(template.EmailAddresses, emails...)
    }
    if len(uris) > 0 {
        template.URIs = append(template.URIs, uris...)
    }

	// Generate CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	// Write CSR to file
	csrFile, err := os.Create(csrPath)
	if err != nil {
		return fmt.Errorf("failed to create CSR file: %w", err)
	}
	defer csrFile.Close()

	if err := pem.Encode(csrFile, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	}); err != nil {
		return fmt.Errorf("failed to write CSR: %w", err)
	}

	// Write private key to file
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyFile.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// ParseCSR parses a CSR from PEM-encoded data
func ParseCSR(data []byte) (*CSRInfo, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR: %w", err)
	}

	info := &CSRInfo{
		Subject:            csr.Subject,
		SignatureAlgorithm: csr.SignatureAlgorithm.String(),
	}

	// Collect SANs
	info.SANs = append(info.SANs, csr.DNSNames...)
	for _, ip := range csr.IPAddresses {
		info.SANs = append(info.SANs, "IP:"+ip.String())
	}
	for _, email := range csr.EmailAddresses {
		info.SANs = append(info.SANs, "email:"+email)
	}

	// Determine public key info
	switch pub := csr.PublicKey.(type) {
	case *rsa.PublicKey:
		info.PublicKeyAlgorithm = "RSA"
		info.KeySize = pub.N.BitLen()
	default:
		info.PublicKeyAlgorithm = "Unknown"
	}

	return info, nil
}

// GenerateCA generates a self-signed Certificate Authority certificate
func GenerateCA(options CAOptions, certPath, keyPath string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, options.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare subject
	subject := pkix.Name{
		CommonName: options.CommonName,
	}

	if options.Organization != "" {
		subject.Organization = []string{options.Organization}
	}
	if options.Country != "" {
		subject.Country = []string{options.Country}
	}

	// Prepare CA certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, options.Days),

		// CA specific settings
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            -1, // No path length constraint

		// Key usage for CA
		KeyUsage: x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign |
			x509.KeyUsageDigitalSignature,

		// Extended key usage (optional for CA, but can be useful)
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageEmailProtection,
			x509.ExtKeyUsageTimeStamping,
		},
	}

	// Generate certificate
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template, // Self-signed, so parent is itself
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key to file
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyFile.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Set restrictive permissions on the private key (Unix-like systems only)
	// Windows has different permission semantics
	if runtime.GOOS != "windows" {
		if err := os.Chmod(keyPath, 0600); err != nil {
			return fmt.Errorf("failed to set key permissions: %w", err)
		}
	}

	return nil
}

// SignCSR signs a Certificate Signing Request with a CA
func SignCSR(options SignOptions, certPath string) error {
	// Read CSR
	csrData, err := os.ReadFile(options.CSRPath)
	if err != nil {
		return fmt.Errorf("failed to read CSR: %w", err)
	}

	block, _ := pem.Decode(csrData)
	if block == nil {
		return fmt.Errorf("failed to parse CSR PEM block")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CSR: %w", err)
	}

	// Verify CSR signature
	if err := csr.CheckSignature(); err != nil {
		return fmt.Errorf("CSR signature verification failed: %w", err)
	}

	// Read CA certificate
	caCertData, err := os.ReadFile(options.CACert)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caBlock, _ := pem.Decode(caCertData)
	if caBlock == nil {
		return fmt.Errorf("failed to parse CA certificate PEM block")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Read CA private key
	caKeyData, err := os.ReadFile(options.CAKey)
	if err != nil {
		return fmt.Errorf("failed to read CA private key: %w", err)
	}

	keyBlock, _ := pem.Decode(caKeyData)
	if keyBlock == nil {
		return fmt.Errorf("failed to parse CA private key PEM block")
	}

	caKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		// Try PKCS1 format
		if rsaKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes); err == nil {
			caKey = rsaKey
		} else {
			return fmt.Errorf("failed to parse CA private key: %w", err)
		}
	}

	// Generate a random serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Prepare certificate template based on CSR
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, options.Days),

		// Standard certificate settings
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Handle SANs - use provided SANs or fall back to CSR SANs
    if len(options.SANs) > 0 {
        // Override with provided SANs
        dns, ips, emails, uris := splitSANs(options.SANs)
        template.DNSNames = append(template.DNSNames, dns...)
        template.IPAddresses = append(template.IPAddresses, ips...)
        if len(emails) > 0 { template.EmailAddresses = append(template.EmailAddresses, emails...) }
        if len(uris) > 0 { template.URIs = append(template.URIs, uris...) }
    } else {
		// Use SANs from CSR
		template.DNSNames = csr.DNSNames
		template.IPAddresses = csr.IPAddresses
		template.EmailAddresses = csr.EmailAddresses
		template.URIs = csr.URIs
	}

	// Create certificate
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		caCert,
		csr.PublicKey,
		caKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	return nil
}

// GenerateOptions contains options for certificate generation
type GenerateOptions struct {
	CommonName string
	Days       int
	KeySize    int
	SANs       []string
	OutputDir  string
}

// VerificationResult contains the results of certificate verification
type VerificationResult struct {
	Certificate *Certificate
	IsValid     bool
	Errors      []string
	Warnings    []string
}

// CSROptions contains options for CSR generation
type CSROptions struct {
	CommonName         string
	Organization       string
	OrganizationalUnit string
	Country            string
	Province           string
	Locality           string
	EmailAddress       string
	SANs               []string
	KeySize            int
}

// CSRInfo contains parsed CSR information for display
type CSRInfo struct {
	Subject            pkix.Name
	SANs               []string
	SignatureAlgorithm string
	PublicKeyAlgorithm string
	KeySize            int
}

// CAOptions contains options for CA certificate generation
type CAOptions struct {
	CommonName   string
	Organization string
	Country      string
	Days         int
	KeySize      int
}

// SignOptions contains options for signing a CSR
type SignOptions struct {
	CSRPath string
	CACert  string
	CAKey   string
	Days    int
	SANs    []string // Optional: override CSR SANs
}

// TLSVersion represents a TLS version constant
type TLSVersion uint16

const (
	TLSVersionTLS10 TLSVersion = tls.VersionTLS10
	TLSVersionTLS11 TLSVersion = tls.VersionTLS11
	TLSVersionTLS12 TLSVersion = tls.VersionTLS12
	TLSVersionTLS13 TLSVersion = tls.VersionTLS13
)

// TLSVersionInfo contains information about a single TLS version test
type TLSVersionInfo struct {
	Version TLSVersion
	Name    string
	Supported bool
	Error   string
}

// TLSResult contains the results of TLS version testing
type TLSResult struct {
	Host         string
	Port         int
	Versions     []TLSVersionInfo
	MinSupported TLSVersion
	MaxSupported TLSVersion
}

// tlsVersionNames maps TLS versions to their human-readable names
var tlsVersionNames = map[TLSVersion]string{
	TLSVersionTLS10: "TLS 1.0",
	TLSVersionTLS11: "TLS 1.1",
	TLSVersionTLS12: "TLS 1.2",
	TLSVersionTLS13: "TLS 1.3",
}

// CheckTLSVersions tests which TLS versions are supported by a server
func CheckTLSVersions(host string, port int, timeout time.Duration) (*TLSResult, error) {
	result := &TLSResult{
		Host:     host,
		Port:     port,
		Versions: make([]TLSVersionInfo, 0, 4),
	}

	dialHost := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{Timeout: timeout}

	// Test each TLS version
	versions := []TLSVersion{
		TLSVersionTLS10,
		TLSVersionTLS11,
		TLSVersionTLS12,
		TLSVersionTLS13,
	}

	var minSupported, maxSupported TLSVersion

	for _, version := range versions {
		info := TLSVersionInfo{
			Version: version,
			Name:    tlsVersionNames[version],
			Supported: false,
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
			MinVersion:         uint16(version),
			MaxVersion:         uint16(version),
		}

		conn, err := tls.DialWithDialer(dialer, "tcp", dialHost, tlsConfig)
		if err != nil {
			// Check if it's a version-specific error
			info.Error = err.Error()
			info.Supported = false
		} else {
			info.Supported = true
			_ = conn.Close()
		}

		result.Versions = append(result.Versions, info)

		// Track min and max supported versions
		if info.Supported {
			if minSupported == 0 || version < minSupported {
				minSupported = version
			}
			if version > maxSupported {
				maxSupported = version
			}
		}
	}

	result.MinSupported = minSupported
	result.MaxSupported = maxSupported

	return result, nil
}
