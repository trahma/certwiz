package ui

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"strings"
	"time"

	"certwiz/internal/config"
	env "certwiz/internal/environ"
	"certwiz/pkg/cert"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	// Color palette
	green  = lipgloss.Color("#00ff00")
	red    = lipgloss.Color("#ff0000")
	yellow = lipgloss.Color("#ffff00")
	cyan   = lipgloss.Color("#00ffff")
	blue   = lipgloss.Color("#0066cc")
	white  = lipgloss.Color("#ffffff")
)

// uiConfig holds the current UI configuration
var uiConfig *config.Config

// SetConfig sets the UI configuration
func SetConfig(cfg *config.Config) {
	uiConfig = cfg
}

// getConfig returns the current config, loading default if not set
func getConfig() *config.Config {
	if uiConfig == nil {
		uiConfig = config.DefaultConfig()
	}
	return uiConfig
}

// Style getters that respect config
func getTitleStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if getConfig().ShouldShowColors() {
		style = style.Foreground(cyan)
	}
	return style
}

func getHeaderStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if getConfig().ShouldShowColors() {
		style = style.Foreground(blue)
	}
	return style
}

func getSuccessStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if getConfig().ShouldShowColors() {
		style = style.Foreground(green)
	}
	return style
}

func getErrorStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if getConfig().ShouldShowColors() {
		style = style.Foreground(red)
	}
	return style
}

func getWarningStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if getConfig().ShouldShowColors() {
		style = style.Foreground(yellow)
	}
	return style
}

func getKeyStyle() lipgloss.Style {
	style := lipgloss.NewStyle()
	if getConfig().ShouldShowColors() {
		style = style.Foreground(cyan)
	}
	return style
}

func getValueStyle() lipgloss.Style {
	style := lipgloss.NewStyle()
	if getConfig().ShouldShowColors() {
		style = style.Foreground(white)
	}
	return style
}

// getEmoji returns emoji or ASCII based on config and environment
func getEmoji(emoji, ascii string) string {
	if !getConfig().ShouldShowEmojis() || env.IsCI() {
		return ascii
	}
	return emoji
}

// getPanelStyle returns the appropriate panel style based on environment and config
func getPanelStyle() lipgloss.Style {
	cfg := getConfig()

	// If borders are disabled, return a simple style with just padding
	if !cfg.ShouldShowBorders() {
		return lipgloss.NewStyle().Padding(0, 0)
	}

	// Check if we're in a CI environment or terminal doesn't support Unicode
	if env.IsCI() || !env.SupportsUnicode() {
		// Use ASCII borders for CI environments
		style := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1, 2)
		if cfg.ShouldShowColors() {
			style = style.BorderForeground(cyan)
		}
		return style
	}

	// Use rounded borders for regular terminals
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)
	if cfg.ShouldShowColors() {
		style = style.BorderForeground(cyan)
	}
	return style
}

// isCI/supportsUnicode logic moved to internal/env

// DisplayCertificate shows certificate information in a formatted table
func DisplayCertificate(cert *cert.Certificate, showFull bool) {
	title := "Certificate Information"
	if cert.Source != "" {
		if strings.HasPrefix(cert.Source, "http") {
			title = fmt.Sprintf("Certificate for %s", cert.Source)
		} else {
			title = fmt.Sprintf("Certificate from %s", cert.Source)
		}
	}

	fmt.Println(getTitleStyle().Render(title))
	fmt.Println()

	// Basic information table
	table := [][]string{
		{"Subject", formatSubject(cert.Subject)},
		{"Issuer", formatSubject(cert.Issuer)},
		{"Serial Number", fmt.Sprintf("%x", cert.SerialNumber)},
		{"Valid From", formatDate(cert.NotBefore)},
		{"Valid To", formatDate(cert.NotAfter)},
		{"Status", formatStatus(cert)},
		{"Public Key", formatPublicKey(cert.PublicKey)},
		{"Signature Algorithm", cert.SignatureAlgorithm.String()},
	}

	// Add TLS connection info if available (URL inspection only)
	if cert.TLSVersion != 0 {
		table = append(table, []string{"TLS Version", tls.VersionName(cert.TLSVersion)})
	}
	if cert.CipherSuite != 0 {
		table = append(table, []string{"Cipher Suite", tls.CipherSuiteName(cert.CipherSuite)})
	}

	// Add SANs if present
	if len(cert.DNSNames) > 0 || len(cert.IPAddresses) > 0 {
		sans := []string{}
		sans = append(sans, cert.DNSNames...)
		for _, ip := range cert.IPAddresses {
			sans = append(sans, ip.String())
		}

		// Format SANs with word wrapping
		sanText := formatSANs(sans)
		// Add count in parentheses if there are many SANs
		if len(sans) > 10 {
			sanText = fmt.Sprintf("(%d total)\n                      %s", len(sans), sanText)
		}
		table = append(table, []string{"SANs", sanText})
	}

	// Display table
	content := formatTable(table)

	var borderColor lipgloss.Color
	if cert.IsExpired {
		borderColor = red
	} else if cert.DaysUntilExpiry < 30 {
		borderColor = yellow
	} else {
		borderColor = green
	}

	// Get terminal width to constrain the panel
	width, _, err := term.GetSize(0)
	if err != nil || width <= 0 {
		width = 80 // default fallback
	}

	// Constrain panel to terminal width
	// The panel adds borders and padding, so we need to account for that
	panel := getPanelStyle().
		BorderForeground(borderColor).
		Width(width - 4) // Account for terminal margins
	fmt.Println(panel.Render(content))

	if showFull {
		displayExtensions(cert.Certificate)
	}
}

// DisplayGenerationResult shows the result of certificate generation
func DisplayGenerationResult(certPath, keyPath string) {
	checkmark := getEmoji("✓", "[OK]")
	fmt.Println(getSuccessStyle().Render(fmt.Sprintf("%s Certificate generated successfully!", checkmark)))
	fmt.Println()

	table := [][]string{
		{"Certificate", certPath},
		{"Private Key", keyPath},
	}

	content := formatTable(table)
	fmt.Println(getPanelStyle().Render(content))
}

// DisplayConversionResult shows the result of certificate conversion
func DisplayConversionResult(inputPath, outputPath, fromFormat, toFormat string) {
	checkmark := getEmoji("✓", "[OK]")
	fmt.Println(getSuccessStyle().Render(fmt.Sprintf("%s Converted from %s to %s", checkmark, strings.ToUpper(fromFormat), strings.ToUpper(toFormat))))
	fmt.Println()

	table := [][]string{
		{"Input", inputPath},
		{"Output", outputPath},
	}

	content := formatTable(table)
	fmt.Println(getPanelStyle().Render(content))
}

// DisplayVerificationResult shows certificate verification results
func DisplayVerificationResult(result *cert.VerificationResult) {
	title := "Verification Results"
	fmt.Println(getTitleStyle().Render(title))
	fmt.Println()

	// Overall status
	checkmark := getEmoji("✓", "[OK]")
	crossMark := getEmoji("✗", "[FAIL]")
	if result.IsValid {
		fmt.Println(getSuccessStyle().Render(fmt.Sprintf("%s Certificate is valid", checkmark)))
	} else {
		fmt.Println(getErrorStyle().Render(fmt.Sprintf("%s Certificate validation failed", crossMark)))
	}
	fmt.Println()

	// Show errors
	if len(result.Errors) > 0 {
		errMark := getEmoji("✗", "[X]")
		fmt.Println(getErrorStyle().Render("Errors:"))
		for _, err := range result.Errors {
			fmt.Printf("  %s %s\n", getErrorStyle().Render(errMark), err)
		}
		fmt.Println()
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		warnSymbol := getEmoji("⚠", "[!]")
		fmt.Println(getWarningStyle().Render("Warnings:"))
		for _, warning := range result.Warnings {
			fmt.Printf("  %s %s\n", getWarningStyle().Render(warnSymbol), warning)
		}
		fmt.Println()
	}

	// Show basic checks
	now := time.Now()
	cert := result.Certificate.Certificate

	checks := [][]string{}

	// Date checks
	checkmark2 := getEmoji("✓", "[OK]")
	crossMark2 := getEmoji("✗", "[X]")
	if cert.NotBefore.After(now) {
		checks = append(checks, []string{crossMark2, "Not yet valid", getErrorStyle().Render("FAIL")})
	} else if cert.NotAfter.Before(now) {
		checks = append(checks, []string{crossMark2, "Expired", getErrorStyle().Render("FAIL")})
	} else {
		checks = append(checks, []string{checkmark2, "Date validity", getSuccessStyle().Render("PASS")})
	}

	if len(checks) > 0 {
		fmt.Println(getHeaderStyle().Render("Validation Checks:"))
		for _, check := range checks {
			fmt.Printf("  %s %s: %s\n", check[0], check[1], check[2])
		}
	}
}

// ShowError displays an error message
func ShowError(message string) {
	fmt.Println(getErrorStyle().Render(fmt.Sprintf("Error: %s", message)))
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	fmt.Println(getSuccessStyle().Render(message))
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	fmt.Println(getKeyStyle().Render(message))
}

// formatTable creates a formatted table from key-value pairs
func formatTable(data [][]string) string {
	var result strings.Builder

	// Find the maximum key length for alignment
	maxKeyLen := 0
	for _, row := range data {
		if len(row[0]) > maxKeyLen {
			maxKeyLen = len(row[0])
		}
	}

	for _, row := range data {
		key := fmt.Sprintf("%-*s", maxKeyLen, row[0])
		result.WriteString(fmt.Sprintf("%s: %s\n",
			getKeyStyle().Render(key),
			getValueStyle().Render(row[1])))
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// formatSubject formats certificate subject/issuer
func formatSubject(subject pkix.Name) string {
	parts := []string{}
	if subject.CommonName != "" {
		parts = append(parts, fmt.Sprintf("CN=%s", subject.CommonName))
	}
	if len(subject.Organization) > 0 {
		parts = append(parts, fmt.Sprintf("O=%s", strings.Join(subject.Organization, ", ")))
	}
	if len(subject.OrganizationalUnit) > 0 {
		parts = append(parts, fmt.Sprintf("OU=%s", strings.Join(subject.OrganizationalUnit, ", ")))
	}
	if len(subject.Country) > 0 {
		parts = append(parts, fmt.Sprintf("C=%s", strings.Join(subject.Country, ", ")))
	}

	if len(parts) == 0 {
		return "Unknown"
	}

	return strings.Join(parts, ", ")
}

// formatDate formats a time with color based on validity
func formatDate(t time.Time) string {
	formatted := t.Format("2006-01-02 15:04:05 UTC")
	now := time.Now()

	if t.Before(now) && t.After(now.AddDate(0, 0, -1)) {
		return getWarningStyle().Render(formatted)
	} else if t.After(now) {
		return getSuccessStyle().Render(formatted)
	}

	return formatted
}

// formatStatus formats certificate status with appropriate colors
func formatStatus(cert *cert.Certificate) string {
	if cert.IsExpired {
		return getErrorStyle().Render(fmt.Sprintf("EXPIRED (%d days ago)", -cert.DaysUntilExpiry))
	} else if cert.DaysUntilExpiry < 30 {
		return getWarningStyle().Render(fmt.Sprintf("EXPIRING SOON (%d days remaining)", cert.DaysUntilExpiry))
	} else {
		return getSuccessStyle().Render(fmt.Sprintf("Valid (%d days remaining)", cert.DaysUntilExpiry))
	}
}

// formatPublicKey formats public key information
func formatPublicKey(pubKey interface{}) string {
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA %d bits", key.Size()*8)
	case *ecdsa.PublicKey:
		return fmt.Sprintf("ECDSA %s", key.Curve.Params().Name)
	default:
		return "Unknown"
	}
}

// formatSANs formats SANs with word wrapping based on terminal width
func formatSANs(sans []string) string {
	// Get terminal width
	width, _, err := term.GetSize(0)
	if err != nil || width <= 0 {
		width = 80 // default fallback
	}

	// Calculate actual available width for SANs value
	// Looking at the actual rendering and your screenshot:
	// - Box has "│" on both sides with padding
	// - Key column is about 20 chars ("Signature Algorithm")
	// - Separator ": " is 2 chars
	// - Need extra margin for safety
	// Being VERY conservative: subtract 45 to ensure we never overflow
	availableWidth := width - 45
	if availableWidth < 30 {
		availableWidth = 30 // minimum width for readability
	}

	// Word wrap the SANs individually
	var lines []string
	var currentLine []string
	currentLineLength := 0

	for _, san := range sans {
		// Calculate what the line would be with this SAN
		addLength := len(san)
		if len(currentLine) > 0 {
			addLength += 2 // for ", "
		}

		// Check if adding this SAN would exceed available width
		if currentLineLength+addLength > availableWidth && len(currentLine) > 0 {
			// Save current line and start a new one
			lines = append(lines, strings.Join(currentLine, ", "))
			currentLine = []string{san}
			currentLineLength = len(san)
		} else {
			// Add to current line
			currentLine = append(currentLine, san)
			currentLineLength += addLength
		}
	}

	// Don't forget the last line
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, ", "))
	}

	// Join lines with newline and proper indentation
	// The indentation should align with where the value starts
	// Key column (20) + ": " (2) = 22 spaces
	if len(lines) > 1 {
		result := lines[0]
		for i := 1; i < len(lines); i++ {
			result += "\n" + strings.Repeat(" ", 22) + lines[i]
		}
		return result
	}

	if len(lines) > 0 {
		return lines[0]
	}

	return strings.Join(sans, ", ") // fallback
}

// DisplayCertificateChain shows the certificate chain
func DisplayCertificateChain(chain []*cert.Certificate) {
	if len(chain) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(getTitleStyle().Render("Certificate Chain"))
	fmt.Println()

	for i, c := range chain {
		// Create a summary view for chain certificates
		table := [][]string{
			{"Position", fmt.Sprintf("Chain[%d]", i+1)},
			{"Subject", formatSubject(c.Subject)},
			{"Issuer", formatSubject(c.Issuer)},
			{"Valid From", c.NotBefore.Format("2006-01-02")},
			{"Valid To", c.NotAfter.Format("2006-01-02")},
		}

		// Determine border color based on validity
		var borderColor lipgloss.Color
		if c.IsExpired {
			borderColor = red
			table = append(table, []string{"Status", getErrorStyle().Render("EXPIRED")})
		} else if c.DaysUntilExpiry < 30 {
			borderColor = yellow
			table = append(table, []string{"Status", getWarningStyle().Render(fmt.Sprintf("Expiring in %d days", c.DaysUntilExpiry))})
		} else {
			borderColor = green
			table = append(table, []string{"Status", getSuccessStyle().Render("Valid")})
		}

		content := formatTable(table)

		// Get terminal width to constrain the panel
		width, _, err := term.GetSize(0)
		if err != nil || width <= 0 {
			width = 80
		}

		panel := getPanelStyle().
			BorderForeground(borderColor).
			Width(width - 4)
		fmt.Println(panel.Render(content))

		if i < len(chain)-1 {
			fmt.Println() // Space between chain certificates
		}
	}
}

// displayExtensions shows certificate extensions (for --full output)
func displayExtensions(cert *x509.Certificate) {
	if len(cert.Extensions) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(getHeaderStyle().Render("Certificate Extensions"))
	fmt.Println()

	// Display parsed extensions with details
	displayParsedExtensions(cert)

	// Display any remaining unparsed extensions
	displayUnparsedExtensions(cert)
}

// displayParsedExtensions shows well-known extensions with their values
func displayParsedExtensions(cert *x509.Certificate) {
	// Key Usage
	if cert.KeyUsage != 0 {
		fmt.Println(getKeyStyle().Render("Key Usage") + getCriticalLabel(isExtensionCritical(cert, "2.5.29.15")))
		displayKeyUsage(cert.KeyUsage)
		fmt.Println()
	}

	// Extended Key Usage
	if len(cert.ExtKeyUsage) > 0 || len(cert.UnknownExtKeyUsage) > 0 {
		fmt.Println(getKeyStyle().Render("Extended Key Usage") + getCriticalLabel(isExtensionCritical(cert, "2.5.29.37")))
		displayExtendedKeyUsage(cert)
		fmt.Println()
	}

	// Basic Constraints
	if cert.BasicConstraintsValid {
		fmt.Println(getKeyStyle().Render("Basic Constraints") + getCriticalLabel(isExtensionCritical(cert, "2.5.29.19")))
		checkmark := getEmoji("✓", "[OK]")
		crossMark := getEmoji("✗", "[X]")
		arrow := getEmoji("→", "->")
		if cert.IsCA {
			fmt.Printf("  %s Certificate Authority: %s\n", getSuccessStyle().Render(checkmark), getSuccessStyle().Render("Yes"))
			if cert.MaxPathLen >= 0 {
				fmt.Printf("  %s Max Path Length: %d\n", getValueStyle().Render(arrow), cert.MaxPathLen)
			} else if cert.MaxPathLenZero {
				fmt.Printf("  %s Max Path Length: %d\n", getValueStyle().Render(arrow), 0)
			}
		} else {
			fmt.Printf("  %s Certificate Authority: %s\n", getValueStyle().Render(crossMark), getValueStyle().Render("No"))
		}
		fmt.Println()
	}

	// Subject Alternative Names (skip if already shown in main display)
	// We show a summary here since full list is in main display
	if len(cert.DNSNames) > 0 || len(cert.IPAddresses) > 0 || len(cert.EmailAddresses) > 0 || len(cert.URIs) > 0 {
		arrow := getEmoji("→", "->")
		fmt.Println(getKeyStyle().Render("Subject Alternative Name") + getCriticalLabel(isExtensionCritical(cert, "2.5.29.17")))
		sanCount := len(cert.DNSNames) + len(cert.IPAddresses) + len(cert.EmailAddresses) + len(cert.URIs)
		fmt.Printf("  %s %d SANs (", getValueStyle().Render(arrow), sanCount)
		parts := []string{}
		if len(cert.DNSNames) > 0 {
			parts = append(parts, fmt.Sprintf("%d DNS", len(cert.DNSNames)))
		}
		if len(cert.IPAddresses) > 0 {
			parts = append(parts, fmt.Sprintf("%d IP", len(cert.IPAddresses)))
		}
		if len(cert.EmailAddresses) > 0 {
			parts = append(parts, fmt.Sprintf("%d Email", len(cert.EmailAddresses)))
		}
		if len(cert.URIs) > 0 {
			parts = append(parts, fmt.Sprintf("%d URI", len(cert.URIs)))
		}
		fmt.Printf("%s)\n", strings.Join(parts, ", "))
		fmt.Println()
	}

	// Authority Info Access
	if len(cert.OCSPServer) > 0 || len(cert.IssuingCertificateURL) > 0 {
		arrow := getEmoji("→", "->")
		link := getEmoji("🔗", "[URL]")
		fmt.Println(getKeyStyle().Render("Authority Info Access"))
		if len(cert.OCSPServer) > 0 {
			fmt.Printf("  %s OCSP:\n", getValueStyle().Render(arrow))
			for _, url := range cert.OCSPServer {
				fmt.Printf("    %s %s\n", getKeyStyle().Render(link), url)
			}
		}
		if len(cert.IssuingCertificateURL) > 0 {
			fmt.Printf("  %s CA Issuers:\n", getValueStyle().Render(arrow))
			for _, url := range cert.IssuingCertificateURL {
				fmt.Printf("    %s %s\n", getKeyStyle().Render(link), url)
			}
		}
		fmt.Println()
	}

	// CRL Distribution Points
	if len(cert.CRLDistributionPoints) > 0 {
		link := getEmoji("🔗", "[URL]")
		fmt.Println(getKeyStyle().Render("CRL Distribution Points"))
		for _, url := range cert.CRLDistributionPoints {
			fmt.Printf("  %s %s\n", getKeyStyle().Render(link), url)
		}
		fmt.Println()
	}

	// Certificate Policies
	if len(cert.PolicyIdentifiers) > 0 {
		arrow := getEmoji("→", "->")
		fmt.Println(getKeyStyle().Render("Certificate Policies"))
		for _, oid := range cert.PolicyIdentifiers {
			policyName := getPolicyName(oid.String())
			fmt.Printf("  %s %s\n", getValueStyle().Render(arrow), policyName)
		}
		fmt.Println()
	}
}

// displayKeyUsage shows the key usage flags
func displayKeyUsage(usage x509.KeyUsage) {
	checkmark := getEmoji("✓", "[OK]")
	usages := []struct {
		flag x509.KeyUsage
		name string
	}{
		{x509.KeyUsageDigitalSignature, "Digital Signature"},
		{x509.KeyUsageContentCommitment, "Content Commitment"},
		{x509.KeyUsageKeyEncipherment, "Key Encipherment"},
		{x509.KeyUsageDataEncipherment, "Data Encipherment"},
		{x509.KeyUsageKeyAgreement, "Key Agreement"},
		{x509.KeyUsageCertSign, "Certificate Signing"},
		{x509.KeyUsageCRLSign, "CRL Signing"},
		{x509.KeyUsageEncipherOnly, "Encipher Only"},
		{x509.KeyUsageDecipherOnly, "Decipher Only"},
	}

	for _, u := range usages {
		if usage&u.flag != 0 {
			fmt.Printf("  %s %s\n", getSuccessStyle().Render(checkmark), u.name)
		}
	}
}

// displayExtendedKeyUsage shows extended key usage
func displayExtendedKeyUsage(cert *x509.Certificate) {
	usageNames := map[x509.ExtKeyUsage]string{
		x509.ExtKeyUsageAny:                            "Any Extended Key Usage",
		x509.ExtKeyUsageServerAuth:                     "TLS Web Server Authentication",
		x509.ExtKeyUsageClientAuth:                     "TLS Web Client Authentication",
		x509.ExtKeyUsageCodeSigning:                    "Code Signing",
		x509.ExtKeyUsageEmailProtection:                "E-mail Protection",
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

	checkmark := getEmoji("✓", "[OK]")
	arrow := getEmoji("→", "->")
	for _, usage := range cert.ExtKeyUsage {
		if name, ok := usageNames[usage]; ok {
			fmt.Printf("  %s %s\n", getSuccessStyle().Render(checkmark), name)
		}
	}

	for _, oid := range cert.UnknownExtKeyUsage {
		fmt.Printf("  %s %s\n", getValueStyle().Render(arrow), oid.String())
	}
}

// displayUnparsedExtensions shows extensions we haven't parsed
func displayUnparsedExtensions(cert *x509.Certificate) {
	// Map of OIDs to names for extensions we don't parse above
	oidNames := map[string]string{
		"2.5.29.14":               "Subject Key Identifier",
		"2.5.29.35":               "Authority Key Identifier",
		"2.5.29.31":               "CRL Distribution Points",
		"2.5.29.32":               "Certificate Policies",
		"1.3.6.1.5.5.7.1.1":       "Authority Info Access",
		"1.3.6.1.4.1.11129.2.4.2": "Certificate Transparency SCT",
		"1.3.6.1.5.5.7.1.12":      "Logo Type",
		"2.5.29.9":                "Subject Directory Attributes",
		"2.5.29.16":               "Private Key Usage Period",
		"2.5.29.20":               "CRL Number",
		"2.5.29.28":               "Issuing Distribution Point",
		"2.5.29.30":               "Name Constraints",
		"2.5.29.33":               "Policy Mappings",
		"2.5.29.36":               "Policy Constraints",
		"2.5.29.54":               "Inhibit Any Policy",
	}

	displayed := map[string]bool{
		"2.5.29.15":         true, // Key Usage
		"2.5.29.17":         true, // SAN
		"2.5.29.19":         true, // Basic Constraints
		"2.5.29.37":         true, // Extended Key Usage
		"2.5.29.31":         true, // CRL Distribution Points
		"2.5.29.32":         true, // Certificate Policies
		"1.3.6.1.5.5.7.1.1": true, // Authority Info Access
	}

	var otherExts []pkix.Extension
	for _, ext := range cert.Extensions {
		if !displayed[ext.Id.String()] {
			otherExts = append(otherExts, ext)
		}
	}

	if len(otherExts) > 0 {
		fmt.Println(getKeyStyle().Render("Other Extensions"))
		arrow := getEmoji("→", "->")
		for _, ext := range otherExts {
			name := ext.Id.String()
			if n, ok := oidNames[name]; ok {
				name = n
			}
			critical := ""
			if ext.Critical {
				critical = getErrorStyle().Render(" [CRITICAL]")
			}
			fmt.Printf("  %s %s%s\n", getValueStyle().Render(arrow), name, critical)
		}
	}
}

// isExtensionCritical checks if an extension is marked as critical
func isExtensionCritical(cert *x509.Certificate, oid string) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.String() == oid {
			return ext.Critical
		}
	}
	return false
}

// getCriticalLabel returns a formatted critical label if critical
func getCriticalLabel(critical bool) string {
	if critical {
		return getErrorStyle().Render(" [CRITICAL]")
	}
	return ""
}

// getPolicyName returns a human-readable name for common policy OIDs
func getPolicyName(oid string) string {
	policies := map[string]string{
		"2.5.29.32.0":                "Any Policy",
		"2.23.140.1.2.1":             "Domain Validated",
		"2.23.140.1.2.2":             "Organization Validated",
		"2.23.140.1.2.3":             "Individual Validated",
		"2.23.140.1.1":               "Extended Validation",
		"1.3.6.1.4.1.6449.1.2.1.3.1": "StartCom Domain Validated",
		"1.3.6.1.4.1.6449.1.2.1.5.1": "StartCom Organization Validated",
		"1.3.6.1.4.1.6449.1.2.1.6.1": "StartCom Extended Validation",
	}

	if name, ok := policies[oid]; ok {
		return fmt.Sprintf("%s (%s)", name, oid)
	}
	return oid
}

// DisplayCSRInfo displays Certificate Signing Request information
func DisplayCSRInfo(info *cert.CSRInfo) {
	// Create a table with CSR information
	table := [][]string{
		{"Subject", formatSubject(info.Subject)},
		{"Signature Algorithm", info.SignatureAlgorithm},
		{"Public Key", fmt.Sprintf("%s %d bits", info.PublicKeyAlgorithm, info.KeySize)},
	}

	// Add SANs if present
	if len(info.SANs) > 0 {
		sanText := formatSANs(info.SANs)
		table = append(table, []string{"Subject Alt Names", sanText})
	}

	// Display the table
	content := formatTable(table)

	// Get terminal width to constrain the panel
	width, _, err := term.GetSize(0)
	if err != nil || width <= 0 {
		width = 80
	}

	panel := getPanelStyle().
		BorderForeground(cyan).
		Width(width - 4)

	fmt.Println(panel.Render(content))
}

// DisplayTLSVersionResults shows TLS version test results
func DisplayTLSVersionResults(result *cert.TLSResult) {
	title := fmt.Sprintf("TLS Version Support for %s:%d", result.Host, result.Port)
	fmt.Println(getTitleStyle().Render(title))
	fmt.Println()

	// Determine symbols based on config and environment
	checkmark := getEmoji("✓", "[OK]")
	crossMark := getEmoji("✗", "[X]")

	// Create a table for version results
	table := [][]string{}
	for _, v := range result.Versions {
		var status string
		if v.Supported {
			cipherInfo := ""
			if v.CipherSuite != 0 {
				cipherInfo = fmt.Sprintf(" (%s)", tls.CipherSuiteName(v.CipherSuite))
			}
			status = fmt.Sprintf("%s %s%s", getSuccessStyle().Render(checkmark), getSuccessStyle().Render("Supported"), cipherInfo)
		} else {
			status = fmt.Sprintf("%s %s", getErrorStyle().Render(crossMark), getErrorStyle().Render("Not Supported"))
		}
		table = append(table, []string{v.Name, status})
	}

	// Display the table
	content := formatTable(table)

	// Get terminal width
	width, _, err := term.GetSize(0)
	if err != nil || width <= 0 {
		width = 80
	}

	panel := getPanelStyle().
		BorderForeground(cyan).
		Width(width - 4)
	fmt.Println(panel.Render(content))

	// Show summary
	fmt.Println()
	fmt.Println(getHeaderStyle().Render("Summary"))
	fmt.Println()

	if result.MinSupported != 0 {
		minName := tlsVersionNames(result.MinSupported)
		fmt.Printf("  %s Minimum supported version: %s\n", getKeyStyle().Render("→"), getSuccessStyle().Render(minName))
	}
	if result.MaxSupported != 0 {
		maxName := tlsVersionNames(result.MaxSupported)
		fmt.Printf("  %s Maximum supported version: %s\n", getKeyStyle().Render("→"), getSuccessStyle().Render(maxName))
	}

	// Security recommendations
	fmt.Println()
	recommendations := []string{}
	for _, v := range result.Versions {
		if v.Supported && (v.Version == cert.TLSVersionTLS10 || v.Version == cert.TLSVersionTLS11) {
			recommendations = append(recommendations, fmt.Sprintf(" %s is enabled but deprecated", v.Name))
		}
	}

	if len(recommendations) > 0 {
		warnSymbol := getEmoji("⚠", "[!]")
		arrow := getEmoji("→", "->")
		fmt.Println(getWarningStyle().Render(fmt.Sprintf("%s Security Warning:", warnSymbol)))
		for _, rec := range recommendations {
			fmt.Printf("  %s%s\n", getWarningStyle().Render(arrow), rec)
		}
		fmt.Println()
		fmt.Println(getKeyStyle().Render("Recommendation: Consider disabling TLS 1.0 and TLS 1.1 for improved security."))
	}
}

// tlsVersionNames is a helper to get version names
func tlsVersionNames(v cert.TLSVersion) string {
	switch v {
	case cert.TLSVersionTLS10:
		return "TLS 1.0"
	case cert.TLSVersionTLS11:
		return "TLS 1.1"
	case cert.TLSVersionTLS12:
		return "TLS 1.2"
	case cert.TLSVersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}
