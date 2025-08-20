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
	"strings"
	"time"
)

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
	// Parse and normalize URL
	if !strings.Contains(targetURL, "://") {
		targetURL = "https://" + targetURL
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Hostname()
	if u.Port() != "" {
		host = net.JoinHostPort(u.Hostname(), u.Port())
	} else {
		host = fmt.Sprintf("%s:%d", host, port)
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", host, &tls.Config{
		InsecureSkipVerify: true, // We want to inspect even invalid certs
		ServerName:         u.Hostname(),
	})
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

	// Add Subject Alternative Names
	if len(opts.SANs) > 0 {
		for _, san := range opts.SANs {
			if strings.Contains(san, ":") {
				parts := strings.SplitN(san, ":", 2)
				if strings.ToLower(parts[0]) == "ip" {
					if ip := net.ParseIP(parts[1]); ip != nil {
						template.IPAddresses = append(template.IPAddresses, ip)
					}
				} else {
					template.DNSNames = append(template.DNSNames, san)
				}
			} else {
				template.DNSNames = append(template.DNSNames, san)
			}
		}
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

	// TODO: Implement CA verification if caPath is provided

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

	// Process SANs
	for _, san := range options.SANs {
		if strings.HasPrefix(san, "IP:") {
			ipStr := strings.TrimPrefix(san, "IP:")
			ip := net.ParseIP(ipStr)
			if ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			}
		} else {
			template.DNSNames = append(template.DNSNames, san)
		}
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

	// Set restrictive permissions on the private key
	if err := os.Chmod(keyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key permissions: %w", err)
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
		for _, san := range options.SANs {
			if strings.HasPrefix(san, "IP:") {
				ipStr := strings.TrimPrefix(san, "IP:")
				ip := net.ParseIP(ipStr)
				if ip != nil {
					template.IPAddresses = append(template.IPAddresses, ip)
				}
			} else {
				template.DNSNames = append(template.DNSNames, san)
			}
		}
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
