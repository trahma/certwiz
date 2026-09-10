package cert

import "crypto/x509"

// keyUsageNames lists key usage flags with their display names, in the
// order they are reported.
var keyUsageNames = []struct {
	flag x509.KeyUsage
	name string
}{
	{x509.KeyUsageDigitalSignature, "Digital Signature"},
	{x509.KeyUsageContentCommitment, "Content Commitment"},
	{x509.KeyUsageKeyEncipherment, "Key Encipherment"},
	{x509.KeyUsageDataEncipherment, "Data Encipherment"},
	{x509.KeyUsageKeyAgreement, "Key Agreement"},
	{x509.KeyUsageCertSign, "Certificate Sign"},
	{x509.KeyUsageCRLSign, "CRL Sign"},
	{x509.KeyUsageEncipherOnly, "Encipher Only"},
	{x509.KeyUsageDecipherOnly, "Decipher Only"},
}

// extKeyUsageNames maps extended key usages to their display names.
var extKeyUsageNames = map[x509.ExtKeyUsage]string{
	x509.ExtKeyUsageAny:                            "Any",
	x509.ExtKeyUsageServerAuth:                     "Server Authentication",
	x509.ExtKeyUsageClientAuth:                     "Client Authentication",
	x509.ExtKeyUsageCodeSigning:                    "Code Signing",
	x509.ExtKeyUsageEmailProtection:                "Email Protection",
	x509.ExtKeyUsageIPSECEndSystem:                 "IPSec End System",
	x509.ExtKeyUsageIPSECTunnel:                    "IPSec Tunnel",
	x509.ExtKeyUsageIPSECUser:                      "IPSec User",
	x509.ExtKeyUsageTimeStamping:                   "Time Stamping",
	x509.ExtKeyUsageOCSPSigning:                    "OCSP Signing",
	x509.ExtKeyUsageMicrosoftServerGatedCrypto:     "Microsoft Server Gated Crypto",
	x509.ExtKeyUsageNetscapeServerGatedCrypto:      "Netscape Server Gated Crypto",
	x509.ExtKeyUsageMicrosoftCommercialCodeSigning: "Microsoft Commercial Code Signing",
	x509.ExtKeyUsageMicrosoftKernelCodeSigning:     "Microsoft Kernel Code Signing",
}

// KeyUsageNames returns the display names of every key usage flag set in
// usage, in a fixed canonical order. It returns nil when no flags are set.
func KeyUsageNames(usage x509.KeyUsage) []string {
	var names []string
	for _, u := range keyUsageNames {
		if usage&u.flag != 0 {
			names = append(names, u.name)
		}
	}
	return names
}

// ExtKeyUsageName returns the display name for an extended key usage and
// whether the usage is known.
func ExtKeyUsageName(usage x509.ExtKeyUsage) (string, bool) {
	name, ok := extKeyUsageNames[usage]
	return name, ok
}

// ExtKeyUsageNames returns the display names of the given extended key
// usages in input order, skipping any that are not known.
func ExtKeyUsageNames(usages []x509.ExtKeyUsage) []string {
	var names []string
	for _, u := range usages {
		if name, ok := extKeyUsageNames[u]; ok {
			names = append(names, name)
		}
	}
	return names
}
