package ui

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"os"
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

const (
	// defaultTerminalWidth is used when the terminal size cannot be determined.
	defaultTerminalWidth = 80
	// panelMargin is subtracted from the terminal width when sizing a panel.
	panelMargin = 4
	// panelPadding is the total horizontal padding inside a panel (2 each side).
	panelPadding = 4
	// minValueWidth is the narrowest column a wrapped value will be laid out in.
	minValueWidth = 24
)

// uiConfig holds the current UI configuration
var uiConfig *config.Config

// textStyles holds the render-only styles, built once per configuration.
// These styles are only ever used via Render, so sharing them is safe.
type textStyles struct {
	title, header, success, errorS, warning, key, value lipgloss.Style
}

var styles *textStyles

// SetConfig replaces the UI configuration and invalidates the cached styles
func SetConfig(cfg *config.Config) {
	uiConfig = cfg
	styles = nil
}

// getConfig returns the current config, loading default if not set
func getConfig() *config.Config {
	if uiConfig == nil {
		uiConfig = config.DefaultConfig()
	}
	return uiConfig
}

// getStyles returns the cached text styles, building them on first use
func getStyles() *textStyles {
	if styles == nil {
		styles = buildStyles(getConfig())
	}
	return styles
}

func buildStyles(cfg *config.Config) *textStyles {
	color := cfg.ShouldShowColors()
	mk := func(bold bool, c lipgloss.Color, padded bool) lipgloss.Style {
		s := lipgloss.NewStyle().Bold(bold)
		if padded {
			s = s.Padding(0, 1)
		}
		if color {
			s = s.Foreground(c)
		}
		return s
	}
	return &textStyles{
		title:   mk(true, cyan, true),
		header:  mk(true, blue, false),
		success: mk(true, green, false),
		errorS:  mk(true, red, false),
		warning: mk(true, yellow, false),
		key:     mk(false, cyan, false),
		value:   mk(false, white, false),
	}
}

// Style accessors that respect config
func getTitleStyle() lipgloss.Style   { return getStyles().title }
func getHeaderStyle() lipgloss.Style  { return getStyles().header }
func getSuccessStyle() lipgloss.Style { return getStyles().success }
func getErrorStyle() lipgloss.Style   { return getStyles().errorS }
func getWarningStyle() lipgloss.Style { return getStyles().warning }
func getKeyStyle() lipgloss.Style     { return getStyles().key }
func getValueStyle() lipgloss.Style   { return getStyles().value }

// Emoji returns emoji or its ASCII equivalent based on config and environment
func Emoji(emoji, ascii string) string {
	if !getConfig().ShouldShowEmojis() || env.IsCI() {
		return ascii
	}
	return emoji
}

// getPanelStyle returns the appropriate panel style based on environment and config.
// It always builds a fresh style because callers modify it (e.g. BorderForeground).
func getPanelStyle() lipgloss.Style {
	cfg := getConfig()

	// If borders are disabled, return a simple style with just padding
	if !cfg.ShouldShowBorders() {
		return lipgloss.NewStyle().Padding(0, 0)
	}

	// ASCII borders for CI environments or terminals without Unicode support
	border := lipgloss.RoundedBorder()
	if env.IsCI() || !env.SupportsUnicode() {
		border = lipgloss.NormalBorder()
	}

	style := lipgloss.NewStyle().
		Border(border).
		Padding(1, 2)
	if cfg.ShouldShowColors() {
		style = style.BorderForeground(cyan)
	}
	return style
}

// terminalWidth returns the width of the terminal attached to stdout,
// falling back to a default when stdout is not a terminal.
func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return defaultTerminalWidth
	}
	return width
}

// renderPanel renders content inside a panel sized to the terminal width
func renderPanel(content string, termWidth int, borderColor lipgloss.Color) string {
	return getPanelStyle().
		BorderForeground(borderColor).
		Width(termWidth - panelMargin).
		Render(content)
}

// valueWidth returns the number of columns available for a table value
// inside a panel, given the terminal width and the longest key in the table.
// The panel is rendered with Width(termWidth-panelMargin), which in lipgloss
// includes padding but not the border.
func valueWidth(termWidth int, longestKey string) int {
	w := termWidth - panelMargin - panelPadding - len(longestKey) - len(": ")
	if w < minValueWidth {
		return minValueWidth
	}
	return w
}

// longestKey returns the longest key in a key/value table
func longestKey(table [][]string) string {
	longest := ""
	for _, row := range table {
		if len(row[0]) > len(longest) {
			longest = row[0]
		}
	}
	return longest
}

// expiryColor returns the border color reflecting a certificate's validity
func expiryColor(c *cert.Certificate) lipgloss.Color {
	switch {
	case c.IsExpired:
		return red
	case c.DaysUntilExpiry < 30:
		return yellow
	default:
		return green
	}
}

// DisplayCertificate shows certificate information in a formatted table
func DisplayCertificate(c *cert.Certificate, showFull bool) {
	title := "Certificate Information"
	if c.Source != "" {
		if strings.Contains(c.Source, "://") {
			title = fmt.Sprintf("Certificate for %s", c.Source)
		} else {
			title = fmt.Sprintf("Certificate from %s", c.Source)
		}
	}

	fmt.Println(getTitleStyle().Render(title))
	fmt.Println()

	width := terminalWidth()

	// Basic information table
	table := [][]string{
		{"Subject", formatSubject(c.Subject)},
		{"Issuer", formatSubject(c.Issuer)},
		{"Serial Number", fmt.Sprintf("%x", c.SerialNumber)},
		{"Valid From", formatDate(c.NotBefore)},
		{"Valid To", formatDate(c.NotAfter)},
		{"Status", formatStatus(c)},
		{"Public Key", formatPublicKey(c.PublicKey)},
		{"Signature Algorithm", c.SignatureAlgorithm.String()},
		{"SHA-256 Fingerprint", ""},
		{"SHA-1 Fingerprint", ""},
	}

	// Add TLS connection info if available (URL inspection only)
	if c.TLSVersion != 0 {
		table = append(table, []string{"TLS Version", cert.TLSVersionName(c.TLSVersion)})
	}
	if c.CipherSuite != 0 {
		table = append(table, []string{"Cipher Suite", tls.CipherSuiteName(c.CipherSuite)})
	}

	// Wrapped values depend on the key column width, so fill them in once
	// all keys are known.
	avail := valueWidth(width, longestKey(table))
	for _, row := range table {
		switch row[0] {
		case "SHA-256 Fingerprint":
			row[1] = wrapFingerprint(c.FingerprintSHA256(), avail)
		case "SHA-1 Fingerprint":
			row[1] = wrapFingerprint(c.FingerprintSHA1(), avail)
		}
	}

	// Add SANs if present
	sans := collectSANs(c.Certificate)
	if len(sans) > 0 {
		sanText := formatSANs(sans, avail)
		// Add count on its own line if there are many SANs
		if len(sans) > 10 {
			sanText = fmt.Sprintf("(%d total)\n%s", len(sans), sanText)
		}
		table = append(table, []string{"SANs", sanText})
	}

	fmt.Println(renderPanel(formatTable(table), width, expiryColor(c)))

	if showFull {
		displayExtensions(c.Certificate)
	}
}

// collectSANs gathers every Subject Alternative Name as a display string
func collectSANs(c *x509.Certificate) []string {
	sans := make([]string, 0, len(c.DNSNames)+len(c.IPAddresses)+len(c.EmailAddresses)+len(c.URIs))
	sans = append(sans, c.DNSNames...)
	for _, ip := range c.IPAddresses {
		sans = append(sans, ip.String())
	}
	sans = append(sans, c.EmailAddresses...)
	for _, u := range c.URIs {
		sans = append(sans, u.String())
	}
	return sans
}

// DisplayGenerationResult shows the result of certificate generation
func DisplayGenerationResult(certPath, keyPath string) {
	checkmark := Emoji("✓", "[OK]")
	fmt.Println(getSuccessStyle().Render(fmt.Sprintf("%s Certificate generated successfully!", checkmark)))
	fmt.Println()

	table := [][]string{
		{"Certificate", certPath},
		{"Private Key", keyPath},
	}

	fmt.Println(getPanelStyle().Render(formatTable(table)))
}

// DisplayConversionResult shows the result of certificate conversion
func DisplayConversionResult(inputPath, outputPath, fromFormat, toFormat string) {
	checkmark := Emoji("✓", "[OK]")
	fmt.Println(getSuccessStyle().Render(fmt.Sprintf("%s Converted from %s to %s", checkmark, strings.ToUpper(fromFormat), strings.ToUpper(toFormat))))
	fmt.Println()

	table := [][]string{
		{"Input", inputPath},
		{"Output", outputPath},
	}

	fmt.Println(getPanelStyle().Render(formatTable(table)))
}

// DisplayVerificationResult shows certificate verification results
func DisplayVerificationResult(result *cert.VerificationResult) {
	fmt.Println(getTitleStyle().Render("Verification Results"))
	fmt.Println()

	checkmark := Emoji("✓", "[OK]")
	crossMark := Emoji("✗", "[X]")

	// Overall status
	if result.IsValid {
		fmt.Println(getSuccessStyle().Render(fmt.Sprintf("%s Certificate is valid", checkmark)))
	} else {
		fmt.Println(getErrorStyle().Render(fmt.Sprintf("%s Certificate validation failed", Emoji("✗", "[FAIL]"))))
	}
	fmt.Println()

	// Show errors
	if len(result.Errors) > 0 {
		fmt.Println(getErrorStyle().Render("Errors:"))
		for _, err := range result.Errors {
			fmt.Printf("  %s %s\n", getErrorStyle().Render(crossMark), err)
		}
		fmt.Println()
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		warnSymbol := Emoji("⚠", "[!]")
		fmt.Println(getWarningStyle().Render("Warnings:"))
		for _, warning := range result.Warnings {
			fmt.Printf("  %s %s\n", getWarningStyle().Render(warnSymbol), warning)
		}
		fmt.Println()
	}

	// Show basic checks
	now := time.Now()
	c := result.Certificate.Certificate
	pass := getSuccessStyle().Render("PASS")
	fail := getErrorStyle().Render("FAIL")

	var checks [][]string

	// Date checks
	switch {
	case c.NotBefore.After(now):
		checks = append(checks, []string{crossMark, "Not yet valid", fail})
	case c.NotAfter.Before(now):
		checks = append(checks, []string{crossMark, "Expired", fail})
	default:
		checks = append(checks, []string{checkmark, "Date validity", pass})
	}

	// Private key match check
	if result.KeyChecked {
		if result.KeyMatches {
			checks = append(checks, []string{checkmark, "Private key match", pass})
		} else {
			checks = append(checks, []string{crossMark, "Private key match", fail})
		}
	}

	if len(checks) > 0 {
		fmt.Println(getHeaderStyle().Render("Validation Checks:"))
		for _, check := range checks {
			fmt.Printf("  %s %s: %s\n", check[0], check[1], check[2])
		}
	}
}

// ShowError displays an error message on stderr
func ShowError(message string) {
	ShowErrorTo(os.Stderr, message)
}

// ShowErrorTo writes a styled error message to w. Errors go to stderr by
// default so they survive stdout redirection.
func ShowErrorTo(w io.Writer, message string) {
	_, _ = fmt.Fprintln(w, getErrorStyle().Render(fmt.Sprintf("Error: %s", message)))
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	fmt.Println(getSuccessStyle().Render(message))
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	fmt.Println(getKeyStyle().Render(message))
}

// formatTable creates a formatted table from key-value pairs. Values may
// span multiple lines; continuation lines are indented to the value column.
func formatTable(data [][]string) string {
	maxKeyLen := len(longestKey(data))
	indent := strings.Repeat(" ", maxKeyLen+len(": "))

	var result strings.Builder
	for i, row := range data {
		if i > 0 {
			result.WriteByte('\n')
		}
		key := fmt.Sprintf("%-*s", maxKeyLen, row[0])
		result.WriteString(getKeyStyle().Render(key))
		result.WriteString(": ")

		lines := strings.Split(row[1], "\n")
		for j, line := range lines {
			if j > 0 {
				result.WriteByte('\n')
				result.WriteString(indent)
			}
			result.WriteString(getValueStyle().Render(line))
		}
	}

	return result.String()
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
func formatStatus(c *cert.Certificate) string {
	switch {
	case c.IsExpired:
		return getErrorStyle().Render(fmt.Sprintf("EXPIRED (%d days ago)", -c.DaysUntilExpiry))
	case c.DaysUntilExpiry < 30:
		return getWarningStyle().Render(fmt.Sprintf("EXPIRING SOON (%d days remaining)", c.DaysUntilExpiry))
	default:
		return getSuccessStyle().Render(fmt.Sprintf("Valid (%d days remaining)", c.DaysUntilExpiry))
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

// wrapFingerprint wraps a colon-separated fingerprint on byte boundaries so
// each line fits within width columns. Lines are joined with "\n" and carry
// no indentation; formatTable aligns them to the value column.
func wrapFingerprint(fp string, width int) string {
	if width < minValueWidth {
		width = minValueWidth
	}
	if len(fp) <= width {
		return fp
	}

	// Each byte occupies 3 characters ("AB:"); break between bytes
	perLine := (width / 3) * 3
	var lines []string
	for start := 0; start < len(fp); start += perLine {
		end := start + perLine
		if end > len(fp) {
			end = len(fp)
		}
		lines = append(lines, strings.TrimSuffix(fp[start:end], ":"))
	}
	return strings.Join(lines, "\n")
}

// formatSANs joins SANs with ", ", wrapping so each line fits within width
// columns. Lines are joined with "\n" and carry no indentation; formatTable
// aligns them to the value column.
func formatSANs(sans []string, width int) string {
	if width < minValueWidth {
		width = minValueWidth
	}

	var lines []string
	var current []string
	currentLen := 0

	for _, san := range sans {
		addLen := len(san)
		if len(current) > 0 {
			addLen += len(", ")
		}

		if currentLen+addLen > width && len(current) > 0 {
			lines = append(lines, strings.Join(current, ", "))
			current = []string{san}
			currentLen = len(san)
		} else {
			current = append(current, san)
			currentLen += addLen
		}
	}

	if len(current) > 0 {
		lines = append(lines, strings.Join(current, ", "))
	}

	return strings.Join(lines, "\n")
}

// DisplayCertificateChain shows the certificate chain
func DisplayCertificateChain(chain []*cert.Certificate) {
	if len(chain) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(getTitleStyle().Render("Certificate Chain"))
	fmt.Println()

	width := terminalWidth()

	for i, c := range chain {
		// Create a summary view for chain certificates
		table := [][]string{
			{"Position", fmt.Sprintf("Chain[%d]", i+1)},
			{"Subject", formatSubject(c.Subject)},
			{"Issuer", formatSubject(c.Issuer)},
			{"Valid From", c.NotBefore.Format("2006-01-02")},
			{"Valid To", c.NotAfter.Format("2006-01-02")},
		}

		switch {
		case c.IsExpired:
			table = append(table, []string{"Status", getErrorStyle().Render("EXPIRED")})
		case c.DaysUntilExpiry < 30:
			table = append(table, []string{"Status", getWarningStyle().Render(fmt.Sprintf("Expiring in %d days", c.DaysUntilExpiry))})
		default:
			table = append(table, []string{"Status", getSuccessStyle().Render("Valid")})
		}

		fmt.Println(renderPanel(formatTable(table), width, expiryColor(c)))

		if i < len(chain)-1 {
			fmt.Println() // Space between chain certificates
		}
	}
}

// displayExtensions shows certificate extensions (for --full output)
func displayExtensions(c *x509.Certificate) {
	if len(c.Extensions) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(getHeaderStyle().Render("Certificate Extensions"))
	fmt.Println()

	// Display parsed extensions with details
	displayParsedExtensions(c)

	// Display any remaining unparsed extensions
	displayUnparsedExtensions(c)
}

// Well-known extension OIDs
const (
	oidKeyUsage         = "2.5.29.15"
	oidSubjectAltName   = "2.5.29.17"
	oidBasicConstraints = "2.5.29.19"
	oidExtKeyUsage      = "2.5.29.37"
	oidCRLDistPoints    = "2.5.29.31"
	oidCertPolicies     = "2.5.29.32"
	oidAuthorityInfo    = "1.3.6.1.5.5.7.1.1"
)

// criticalExtensions returns the set of extension OIDs marked critical
func criticalExtensions(c *x509.Certificate) map[string]bool {
	critical := make(map[string]bool, len(c.Extensions))
	for _, ext := range c.Extensions {
		if ext.Critical {
			critical[ext.Id.String()] = true
		}
	}
	return critical
}

// displayParsedExtensions shows well-known extensions with their values
func displayParsedExtensions(c *x509.Certificate) {
	critical := criticalExtensions(c)
	arrow := Emoji("→", "->")
	link := Emoji("🔗", "[URL]")

	// Key Usage
	if c.KeyUsage != 0 {
		fmt.Println(getKeyStyle().Render("Key Usage") + getCriticalLabel(critical[oidKeyUsage]))
		displayKeyUsage(c.KeyUsage)
		fmt.Println()
	}

	// Extended Key Usage
	if len(c.ExtKeyUsage) > 0 || len(c.UnknownExtKeyUsage) > 0 {
		fmt.Println(getKeyStyle().Render("Extended Key Usage") + getCriticalLabel(critical[oidExtKeyUsage]))
		displayExtendedKeyUsage(c)
		fmt.Println()
	}

	// Basic Constraints
	if c.BasicConstraintsValid {
		fmt.Println(getKeyStyle().Render("Basic Constraints") + getCriticalLabel(critical[oidBasicConstraints]))
		if c.IsCA {
			fmt.Printf("  %s Certificate Authority: %s\n", getSuccessStyle().Render(Emoji("✓", "[OK]")), getSuccessStyle().Render("Yes"))
			if c.MaxPathLen >= 0 {
				fmt.Printf("  %s Max Path Length: %d\n", getValueStyle().Render(arrow), c.MaxPathLen)
			} else if c.MaxPathLenZero {
				fmt.Printf("  %s Max Path Length: %d\n", getValueStyle().Render(arrow), 0)
			}
		} else {
			fmt.Printf("  %s Certificate Authority: %s\n", getValueStyle().Render(Emoji("✗", "[X]")), getValueStyle().Render("No"))
		}
		fmt.Println()
	}

	// Subject Alternative Names: summary only, the full list is in the main display
	sanCount := len(c.DNSNames) + len(c.IPAddresses) + len(c.EmailAddresses) + len(c.URIs)
	if sanCount > 0 {
		fmt.Println(getKeyStyle().Render("Subject Alternative Name") + getCriticalLabel(critical[oidSubjectAltName]))
		parts := []string{}
		if len(c.DNSNames) > 0 {
			parts = append(parts, fmt.Sprintf("%d DNS", len(c.DNSNames)))
		}
		if len(c.IPAddresses) > 0 {
			parts = append(parts, fmt.Sprintf("%d IP", len(c.IPAddresses)))
		}
		if len(c.EmailAddresses) > 0 {
			parts = append(parts, fmt.Sprintf("%d Email", len(c.EmailAddresses)))
		}
		if len(c.URIs) > 0 {
			parts = append(parts, fmt.Sprintf("%d URI", len(c.URIs)))
		}
		fmt.Printf("  %s %d SANs (%s)\n", getValueStyle().Render(arrow), sanCount, strings.Join(parts, ", "))
		fmt.Println()
	}

	// Authority Info Access
	if len(c.OCSPServer) > 0 || len(c.IssuingCertificateURL) > 0 {
		fmt.Println(getKeyStyle().Render("Authority Info Access"))
		if len(c.OCSPServer) > 0 {
			fmt.Printf("  %s OCSP:\n", getValueStyle().Render(arrow))
			for _, url := range c.OCSPServer {
				fmt.Printf("    %s %s\n", getKeyStyle().Render(link), url)
			}
		}
		if len(c.IssuingCertificateURL) > 0 {
			fmt.Printf("  %s CA Issuers:\n", getValueStyle().Render(arrow))
			for _, url := range c.IssuingCertificateURL {
				fmt.Printf("    %s %s\n", getKeyStyle().Render(link), url)
			}
		}
		fmt.Println()
	}

	// CRL Distribution Points
	if len(c.CRLDistributionPoints) > 0 {
		fmt.Println(getKeyStyle().Render("CRL Distribution Points"))
		for _, url := range c.CRLDistributionPoints {
			fmt.Printf("  %s %s\n", getKeyStyle().Render(link), url)
		}
		fmt.Println()
	}

	// Certificate Policies
	if len(c.PolicyIdentifiers) > 0 {
		fmt.Println(getKeyStyle().Render("Certificate Policies"))
		for _, oid := range c.PolicyIdentifiers {
			fmt.Printf("  %s %s\n", getValueStyle().Render(arrow), getPolicyName(oid.String()))
		}
		fmt.Println()
	}
}

// displayKeyUsage shows the key usage flags
func displayKeyUsage(usage x509.KeyUsage) {
	checkmark := getSuccessStyle().Render(Emoji("✓", "[OK]"))
	for _, name := range cert.KeyUsageNames(usage) {
		fmt.Printf("  %s %s\n", checkmark, name)
	}
}

// displayExtendedKeyUsage shows extended key usage
func displayExtendedKeyUsage(c *x509.Certificate) {
	checkmark := getSuccessStyle().Render(Emoji("✓", "[OK]"))
	arrow := getValueStyle().Render(Emoji("→", "->"))
	for _, usage := range c.ExtKeyUsage {
		if name, ok := cert.ExtKeyUsageName(usage); ok {
			fmt.Printf("  %s %s\n", checkmark, name)
		}
	}

	for _, oid := range c.UnknownExtKeyUsage {
		fmt.Printf("  %s %s\n", arrow, oid.String())
	}
}

// extensionOIDNames maps OIDs to names for extensions not parsed in detail
var extensionOIDNames = map[string]string{
	"2.5.29.14":               "Subject Key Identifier",
	"2.5.29.35":               "Authority Key Identifier",
	oidCRLDistPoints:          "CRL Distribution Points",
	oidCertPolicies:           "Certificate Policies",
	oidAuthorityInfo:          "Authority Info Access",
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

// displayedExtensions is the set of extensions already shown by displayParsedExtensions
var displayedExtensions = map[string]bool{
	oidKeyUsage:         true,
	oidSubjectAltName:   true,
	oidBasicConstraints: true,
	oidExtKeyUsage:      true,
	oidCRLDistPoints:    true,
	oidCertPolicies:     true,
	oidAuthorityInfo:    true,
}

// displayUnparsedExtensions shows extensions we haven't parsed
func displayUnparsedExtensions(c *x509.Certificate) {
	var otherExts []pkix.Extension
	for _, ext := range c.Extensions {
		if !displayedExtensions[ext.Id.String()] {
			otherExts = append(otherExts, ext)
		}
	}

	if len(otherExts) == 0 {
		return
	}

	fmt.Println(getKeyStyle().Render("Other Extensions"))
	arrow := getValueStyle().Render(Emoji("→", "->"))
	for _, ext := range otherExts {
		name := ext.Id.String()
		if n, ok := extensionOIDNames[name]; ok {
			name = n
		}
		fmt.Printf("  %s %s%s\n", arrow, name, getCriticalLabel(ext.Critical))
	}
}

// getCriticalLabel returns a formatted critical label if critical
func getCriticalLabel(critical bool) string {
	if critical {
		return getErrorStyle().Render(" [CRITICAL]")
	}
	return ""
}

// policyNames maps common certificate policy OIDs to human-readable names
var policyNames = map[string]string{
	"2.5.29.32.0":                "Any Policy",
	"2.23.140.1.2.1":             "Domain Validated",
	"2.23.140.1.2.2":             "Organization Validated",
	"2.23.140.1.2.3":             "Individual Validated",
	"2.23.140.1.1":               "Extended Validation",
	"1.3.6.1.4.1.6449.1.2.1.3.1": "StartCom Domain Validated",
	"1.3.6.1.4.1.6449.1.2.1.5.1": "StartCom Organization Validated",
	"1.3.6.1.4.1.6449.1.2.1.6.1": "StartCom Extended Validation",
}

// getPolicyName returns a human-readable name for common policy OIDs
func getPolicyName(oid string) string {
	if name, ok := policyNames[oid]; ok {
		return fmt.Sprintf("%s (%s)", name, oid)
	}
	return oid
}

// DisplayCSRInfo displays Certificate Signing Request information
func DisplayCSRInfo(info *cert.CSRInfo) {
	width := terminalWidth()

	table := [][]string{
		{"Subject", formatSubject(info.Subject)},
		{"Signature Algorithm", info.SignatureAlgorithm},
		{"Public Key", fmt.Sprintf("%s %d bits", info.PublicKeyAlgorithm, info.KeySize)},
	}

	// Add SANs if present
	if len(info.SANs) > 0 {
		table = append(table, []string{"Subject Alt Names", formatSANs(info.SANs, valueWidth(width, longestKey(table)))})
	}

	fmt.Println(renderPanel(formatTable(table), width, cyan))
}

// DisplayTLSVersionResults shows TLS version test results
func DisplayTLSVersionResults(result *cert.TLSResult) {
	title := fmt.Sprintf("TLS Version Support for %s:%d", result.Host, result.Port)
	fmt.Println(getTitleStyle().Render(title))
	fmt.Println()

	checkmark := Emoji("✓", "[OK]")
	crossMark := Emoji("✗", "[X]")

	// Create a table for version results
	table := make([][]string, 0, len(result.Versions))
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

	fmt.Println(renderPanel(formatTable(table), terminalWidth(), cyan))

	// Show summary
	fmt.Println()
	fmt.Println(getHeaderStyle().Render("Summary"))
	fmt.Println()

	arrow := Emoji("→", "->")
	if result.MinSupported != 0 {
		minName := cert.TLSVersionName(uint16(result.MinSupported))
		fmt.Printf("  %s Minimum supported version: %s\n", getKeyStyle().Render(arrow), getSuccessStyle().Render(minName))
	}
	if result.MaxSupported != 0 {
		maxName := cert.TLSVersionName(uint16(result.MaxSupported))
		fmt.Printf("  %s Maximum supported version: %s\n", getKeyStyle().Render(arrow), getSuccessStyle().Render(maxName))
	}

	// Security recommendations
	fmt.Println()
	var recommendations []string
	for _, v := range result.Versions {
		if v.Supported && (v.Version == cert.TLSVersionTLS10 || v.Version == cert.TLSVersionTLS11) {
			recommendations = append(recommendations, fmt.Sprintf(" %s is enabled but deprecated", v.Name))
		}
	}

	if len(recommendations) > 0 {
		warnSymbol := Emoji("⚠", "[!]")
		fmt.Println(getWarningStyle().Render(fmt.Sprintf("%s Security Warning:", warnSymbol)))
		for _, rec := range recommendations {
			fmt.Printf("  %s%s\n", getWarningStyle().Render(arrow), rec)
		}
		fmt.Println()
		fmt.Println(getKeyStyle().Render("Recommendation: Consider disabling TLS 1.0 and TLS 1.1 for improved security."))
	}
}
