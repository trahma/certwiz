package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"reflect"
	"testing"
	"time"

	"certwiz/internal/testutil"
)

func TestKeyUsageNamesAllFlagsInOrder(t *testing.T) {
	all := x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement | x509.KeyUsageCertSign | x509.KeyUsageCRLSign |
		x509.KeyUsageEncipherOnly | x509.KeyUsageDecipherOnly

	want := []string{
		"Digital Signature",
		"Content Commitment",
		"Key Encipherment",
		"Data Encipherment",
		"Key Agreement",
		"Certificate Sign",
		"CRL Sign",
		"Encipher Only",
		"Decipher Only",
	}
	if got := KeyUsageNames(all); !reflect.DeepEqual(got, want) {
		t.Errorf("KeyUsageNames(all) = %v, want %v", got, want)
	}

	// Subset keeps canonical order regardless of bit position
	got := KeyUsageNames(x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature)
	want = []string{"Digital Signature", "CRL Sign"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("KeyUsageNames(subset) = %v, want %v", got, want)
	}
}

func TestKeyUsageNamesEmpty(t *testing.T) {
	if got := KeyUsageNames(0); len(got) != 0 {
		t.Errorf("KeyUsageNames(0) = %v, want empty", got)
	}
}

func TestExtKeyUsageName(t *testing.T) {
	name, ok := ExtKeyUsageName(x509.ExtKeyUsageServerAuth)
	if !ok || name != "Server Authentication" {
		t.Errorf("ExtKeyUsageName(ServerAuth) = %q, %v", name, ok)
	}
	name, ok = ExtKeyUsageName(x509.ExtKeyUsageMicrosoftKernelCodeSigning)
	if !ok || name != "Microsoft Kernel Code Signing" {
		t.Errorf("ExtKeyUsageName(MicrosoftKernelCodeSigning) = %q, %v", name, ok)
	}
	if name, ok := ExtKeyUsageName(x509.ExtKeyUsage(9999)); ok || name != "" {
		t.Errorf("ExtKeyUsageName(unknown) = %q, %v, want \"\", false", name, ok)
	}
}

func TestExtKeyUsageNamesSkipsUnknownPreservesOrder(t *testing.T) {
	in := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
		x509.ExtKeyUsage(9999),
		x509.ExtKeyUsageServerAuth,
		x509.ExtKeyUsageAny,
	}
	want := []string{"Client Authentication", "Server Authentication", "Any"}
	if got := ExtKeyUsageNames(in); !reflect.DeepEqual(got, want) {
		t.Errorf("ExtKeyUsageNames = %v, want %v", got, want)
	}
	if got := ExtKeyUsageNames(nil); len(got) != 0 {
		t.Errorf("ExtKeyUsageNames(nil) = %v, want empty", got)
	}
}

func TestToJSONKeyUsageStrings(t *testing.T) {
	// valid.pem carries no key usage extension: JSON must report none
	c, err := InspectFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatalf("InspectFile: %v", err)
	}
	jc := c.ToJSON()
	if len(jc.KeyUsage) != 0 || len(jc.ExtKeyUsage) != 0 {
		t.Errorf("valid.pem key_usage=%v ext_key_usage=%v, want none", jc.KeyUsage, jc.ExtKeyUsage)
	}

	// A generated certificate with known usages must produce the canonical strings
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "usage-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	certs, err := InspectData(der, "generated")
	if err != nil {
		t.Fatalf("InspectData: %v", err)
	}
	jc = certs[0].ToJSON()
	wantKU := []string{"Digital Signature", "Key Encipherment"}
	wantEKU := []string{"Server Authentication", "Client Authentication"}
	if !reflect.DeepEqual(jc.KeyUsage, wantKU) {
		t.Errorf("key_usage = %v, want %v", jc.KeyUsage, wantKU)
	}
	if !reflect.DeepEqual(jc.ExtKeyUsage, wantEKU) {
		t.Errorf("ext_key_usage = %v, want %v", jc.ExtKeyUsage, wantEKU)
	}
}
