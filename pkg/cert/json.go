package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"time"
)

// JSONCertificate represents certificate data in JSON format
type JSONCertificate struct {
	Subject            JSONSubject       `json:"subject"`
	Issuer             JSONSubject       `json:"issuer"`
	SerialNumber       string            `json:"serial_number"`
	NotBefore          time.Time         `json:"not_before"`
	NotAfter           time.Time         `json:"not_after"`
	IsCA               bool              `json:"is_ca"`
	IsExpired          bool              `json:"is_expired"`
	DaysUntilExpiry    int               `json:"days_until_expiry"`
	SignatureAlgorithm string            `json:"signature_algorithm"`
	PublicKeyAlgorithm string            `json:"public_key_algorithm"`
	PublicKeySize      int               `json:"public_key_size"`
	DNSNames           []string          `json:"dns_names,omitempty"`
	IPAddresses        []string          `json:"ip_addresses,omitempty"`
	EmailAddresses     []string          `json:"email_addresses,omitempty"`
	URIs               []string          `json:"uris,omitempty"`
	KeyUsage           []string          `json:"key_usage,omitempty"`
	ExtKeyUsage        []string          `json:"ext_key_usage,omitempty"`
	Source             string            `json:"source,omitempty"`
	Format             string            `json:"format,omitempty"`
	Chain              []JSONCertSummary `json:"chain,omitempty"`
}

// JSONSubject represents certificate subject/issuer in JSON format
type JSONSubject struct {
	CommonName         string   `json:"common_name,omitempty"`
	Organization       []string `json:"organization,omitempty"`
	OrganizationalUnit []string `json:"organizational_unit,omitempty"`
	Country            []string `json:"country,omitempty"`
	Province           []string `json:"province,omitempty"`
	Locality           []string `json:"locality,omitempty"`
	StreetAddress      []string `json:"street_address,omitempty"`
	PostalCode         []string `json:"postal_code,omitempty"`
}

// JSONCertSummary represents a certificate in a chain
type JSONCertSummary struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	IsExpired    bool      `json:"is_expired"`
	SerialNumber string    `json:"serial_number"`
}

// JSONCSRInfo represents CSR data in JSON format
type JSONCSRInfo struct {
	Subject            JSONSubject `json:"subject"`
	SignatureAlgorithm string      `json:"signature_algorithm"`
	PublicKeyAlgorithm string      `json:"public_key_algorithm"`
	PublicKeySize      int         `json:"public_key_size"`
	DNSNames           []string    `json:"dns_names,omitempty"`
	IPAddresses        []string    `json:"ip_addresses,omitempty"`
	EmailAddresses     []string    `json:"email_addresses,omitempty"`
	URIs               []string    `json:"uris,omitempty"`
}

// JSONVerificationResult represents verification result in JSON format
type JSONVerificationResult struct {
	IsValid     bool            `json:"is_valid"`
	Errors      []string        `json:"errors,omitempty"`
	Warnings    []string        `json:"warnings,omitempty"`
	Certificate JSONCertificate `json:"certificate"`
}

// JSONOperationResult represents the result of certificate operations
type JSONOperationResult struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Files   []string `json:"files,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// JSONTLSVersionInfo represents TLS version test info in JSON format
type JSONTLSVersionInfo struct {
	Version   string `json:"version"`
	Name      string `json:"name"`
	Supported bool   `json:"supported"`
	Error     string `json:"error,omitempty"`
}

// JSONTLSResult represents TLS version test results in JSON format
type JSONTLSResult struct {
	Host         string               `json:"host"`
	Port         int                  `json:"port"`
	Versions     []JSONTLSVersionInfo `json:"versions"`
	MinSupported string               `json:"min_supported"`
	MaxSupported string               `json:"max_supported"`
}

// ToJSON converts a Certificate to JSONCertificate
func (c *Certificate) ToJSON() JSONCertificate {
	jc := JSONCertificate{
		Subject:            subjectToJSON(c.Subject),
		Issuer:             subjectToJSON(c.Issuer),
		SerialNumber:       c.SerialNumber.Text(16),
		NotBefore:          c.NotBefore,
		NotAfter:           c.NotAfter,
		IsCA:               c.IsCA,
		IsExpired:          c.IsExpired,
		DaysUntilExpiry:    c.DaysUntilExpiry,
		SignatureAlgorithm: c.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: getPublicKeyAlgorithm(c.PublicKey),
		PublicKeySize:      getPublicKeySize(c.PublicKey),
		DNSNames:           c.DNSNames,
		Source:             c.Source,
		Format:             c.Format,
	}

	// Convert IP addresses to strings
	for _, ip := range c.IPAddresses {
		jc.IPAddresses = append(jc.IPAddresses, ip.String())
	}

	// Add email addresses
	jc.EmailAddresses = c.EmailAddresses

	// Convert URIs to strings
	for _, uri := range c.URIs {
		jc.URIs = append(jc.URIs, uri.String())
	}

	// Add key usage
	jc.KeyUsage = getKeyUsageStrings(c.KeyUsage)

	// Add extended key usage
	jc.ExtKeyUsage = getExtKeyUsageStrings(c.ExtKeyUsage)

	return jc
}

// ToJSON converts CSRInfo to JSONCSRInfo
func (info *CSRInfo) ToJSON() JSONCSRInfo {
	ji := JSONCSRInfo{
		Subject:            subjectToJSON(info.Subject),
		SignatureAlgorithm: info.SignatureAlgorithm,
		PublicKeyAlgorithm: info.PublicKeyAlgorithm,
		PublicKeySize:      info.KeySize,
	}

	// Process SANs
	for _, san := range info.SANs {
		if len(san) > 3 && san[:3] == "IP:" {
			ji.IPAddresses = append(ji.IPAddresses, san[3:])
		} else if len(san) > 6 && san[:6] == "email:" {
			ji.EmailAddresses = append(ji.EmailAddresses, san[6:])
		} else {
			ji.DNSNames = append(ji.DNSNames, san)
		}
	}

	return ji
}

// ToJSON converts VerificationResult to JSONVerificationResult
func (vr *VerificationResult) ToJSON() JSONVerificationResult {
	return JSONVerificationResult{
		IsValid:     vr.IsValid,
		Errors:      vr.Errors,
		Warnings:    vr.Warnings,
		Certificate: vr.Certificate.ToJSON(),
	}
}

// ToJSON converts TLSResult to JSONTLSResult
func (tr *TLSResult) ToJSON() JSONTLSResult {
	jsonResult := JSONTLSResult{
		Host:         tr.Host,
		Port:         tr.Port,
		Versions:     make([]JSONTLSVersionInfo, 0, len(tr.Versions)),
		MinSupported: "",
		MaxSupported: "",
	}

	for _, v := range tr.Versions {
		jsonVersion := JSONTLSVersionInfo{
			Version:   fmt.Sprintf("0x%04x", uint16(v.Version)),
			Name:      v.Name,
			Supported: v.Supported,
			Error:     v.Error,
		}
		jsonResult.Versions = append(jsonResult.Versions, jsonVersion)
	}

	if tr.MinSupported != 0 {
		jsonResult.MinSupported = tlsVersionNames[tr.MinSupported]
	}
	if tr.MaxSupported != 0 {
		jsonResult.MaxSupported = tlsVersionNames[tr.MaxSupported]
	}

	return jsonResult
}

// MarshalJSON implements json.Marshaler for TLSResult
func (tr *TLSResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(tr.ToJSON())
}

// MarshalJSON implements json.Marshaler for Certificate
func (c *Certificate) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.ToJSON())
}

// MarshalJSON implements json.Marshaler for CSRInfo
func (info *CSRInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(info.ToJSON())
}

// MarshalJSON implements json.Marshaler for VerificationResult
func (vr *VerificationResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(vr.ToJSON())
}

// Helper functions

func subjectToJSON(subject interface{}) JSONSubject {
	switch s := subject.(type) {
	case pkix.Name:
		return JSONSubject{
			CommonName:         s.CommonName,
			Organization:       s.Organization,
			OrganizationalUnit: s.OrganizationalUnit,
			Country:            s.Country,
			Province:           s.Province,
			Locality:           s.Locality,
			StreetAddress:      s.StreetAddress,
			PostalCode:         s.PostalCode,
		}
	default:
		return JSONSubject{}
	}
}

func getKeyUsageStrings(usage x509.KeyUsage) []string {
	var usages []string

	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "Digital Signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "Content Commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "Key Encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "Data Encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "Key Agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "Certificate Sign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRL Sign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "Encipher Only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "Decipher Only")
	}

	return usages
}

func getExtKeyUsageStrings(usage []x509.ExtKeyUsage) []string {
	var usages []string

	for _, u := range usage {
		switch u {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "Server Authentication")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "Client Authentication")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "Code Signing")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "Email Protection")
		case x509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "IPSec End System")
		case x509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "IPSec Tunnel")
		case x509.ExtKeyUsageIPSECUser:
			usages = append(usages, "IPSec User")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "Time Stamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSP Signing")
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			usages = append(usages, "Microsoft Server Gated Crypto")
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			usages = append(usages, "Netscape Server Gated Crypto")
		}
	}

	return usages
}
