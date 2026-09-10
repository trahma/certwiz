package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"certwiz/internal/testutil"
)

func TestParseCSRRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	csrPath := filepath.Join(tmp, "req.csr")
	keyPath := filepath.Join(tmp, "req.key")

	opts := CSROptions{
		CommonName:         "www.example.test",
		Organization:       "Example Org",
		OrganizationalUnit: "Platform",
		Country:            "US",
		Province:           "California",
		Locality:           "San Francisco",
		EmailAddress:       "admin@example.test",
		SANs: []string{
			"www.example.test",
			"api.example.test",
			"IP:192.0.2.10",
			"email:ops@example.test",
			"uri:https://example.test/id",
		},
		KeySize: 2048,
	}
	if err := GenerateCSR(opts, csrPath, keyPath); err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}

	data, err := os.ReadFile(csrPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := ParseCSR(data)
	if err != nil {
		t.Fatalf("ParseCSR: %v", err)
	}

	subj := info.Subject
	if subj.CommonName != opts.CommonName {
		t.Errorf("CommonName = %q, want %q", subj.CommonName, opts.CommonName)
	}
	want := map[string][]string{
		"Organization":       {opts.Organization},
		"OrganizationalUnit": {opts.OrganizationalUnit},
		"Country":            {opts.Country},
		"Province":           {opts.Province},
		"Locality":           {opts.Locality},
	}
	got := map[string][]string{
		"Organization":       subj.Organization,
		"OrganizationalUnit": subj.OrganizationalUnit,
		"Country":            subj.Country,
		"Province":           subj.Province,
		"Locality":           subj.Locality,
	}
	for field, w := range want {
		if !reflect.DeepEqual(got[field], w) {
			t.Errorf("%s = %v, want %v", field, got[field], w)
		}
	}

	// SANs come back DNS first, then IP, email, URI with their prefixes.
	// The subject email is emitted as an email SAN too.
	wantSANs := []string{
		"www.example.test",
		"api.example.test",
		"IP:192.0.2.10",
		"email:admin@example.test",
		"email:ops@example.test",
		"uri:https://example.test/id",
	}
	if !reflect.DeepEqual(info.SANs, wantSANs) {
		t.Errorf("SANs = %v, want %v", info.SANs, wantSANs)
	}

	if info.SignatureAlgorithm != x509.SHA256WithRSA.String() {
		t.Errorf("SignatureAlgorithm = %q, want %q", info.SignatureAlgorithm, x509.SHA256WithRSA.String())
	}
	if info.PublicKeyAlgorithm != "RSA" {
		t.Errorf("PublicKeyAlgorithm = %q, want RSA", info.PublicKeyAlgorithm)
	}
	if info.KeySize != 2048 {
		t.Errorf("KeySize = %d, want 2048", info.KeySize)
	}
}

func TestParseCSRECDSA(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: "ec.example.test"},
		DNSNames: []string{"ec.example.test"},
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, &tmpl, key)
	if err != nil {
		t.Fatal(err)
	}
	data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})

	info, err := ParseCSR(data)
	if err != nil {
		t.Fatalf("ParseCSR: %v", err)
	}
	if info.PublicKeyAlgorithm != "ECDSA" {
		t.Errorf("PublicKeyAlgorithm = %q, want ECDSA", info.PublicKeyAlgorithm)
	}
	if info.KeySize != 256 {
		t.Errorf("KeySize = %d, want 256", info.KeySize)
	}
	if info.SignatureAlgorithm != x509.ECDSAWithSHA256.String() {
		t.Errorf("SignatureAlgorithm = %q, want %q", info.SignatureAlgorithm, x509.ECDSAWithSHA256.String())
	}
	if !reflect.DeepEqual(info.SANs, []string{"ec.example.test"}) {
		t.Errorf("SANs = %v, want [ec.example.test]", info.SANs)
	}
}

func TestParseCSRErrors(t *testing.T) {
	if _, err := ParseCSR([]byte("not a pem block")); err == nil {
		t.Error("expected error for non-PEM data")
	}

	// A valid PEM block whose payload is a certificate, not a CSR.
	certPEM, err := os.ReadFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseCSR(certPEM); err == nil {
		t.Error("expected error for a PEM block that is not a CSR")
	}
}

func TestCSRInfoToJSONRoutesSANs(t *testing.T) {
	info := &CSRInfo{
		Subject:            pkix.Name{CommonName: "x", Organization: []string{"O"}},
		SignatureAlgorithm: "SHA256-RSA",
		PublicKeyAlgorithm: "RSA",
		KeySize:            2048,
		SANs: []string{
			"a.example.test",
			"IP:10.0.0.1",
			"email:me@example.test",
			"uri:spiffe://example.test/svc",
			"b.example.test",
		},
	}
	j := info.ToJSON()

	if j.Subject.CommonName != "x" || !reflect.DeepEqual(j.Subject.Organization, []string{"O"}) {
		t.Errorf("subject not carried through: %+v", j.Subject)
	}
	if j.SignatureAlgorithm != "SHA256-RSA" || j.PublicKeyAlgorithm != "RSA" || j.PublicKeySize != 2048 {
		t.Errorf("algorithm fields not carried through: %+v", j)
	}
	if !reflect.DeepEqual(j.DNSNames, []string{"a.example.test", "b.example.test"}) {
		t.Errorf("DNSNames = %v", j.DNSNames)
	}
	if !reflect.DeepEqual(j.IPAddresses, []string{"10.0.0.1"}) {
		t.Errorf("IPAddresses = %v", j.IPAddresses)
	}
	if !reflect.DeepEqual(j.EmailAddresses, []string{"me@example.test"}) {
		t.Errorf("EmailAddresses = %v", j.EmailAddresses)
	}
	if !reflect.DeepEqual(j.URIs, []string{"spiffe://example.test/svc"}) {
		t.Errorf("URIs = %v", j.URIs)
	}
}
