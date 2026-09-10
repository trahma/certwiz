package cert

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"certwiz/internal/testutil"
)

func TestTLSVersionName(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0x0300, "0x0300"}, // SSL 3.0: not in the table, falls back to hex
		{0, "0x0000"},
	}
	for _, tt := range tests {
		if got := TLSVersionName(tt.version); got != tt.want {
			t.Errorf("TLSVersionName(0x%04x) = %q, want %q", tt.version, got, tt.want)
		}
	}
}

func TestTLSResultToJSON(t *testing.T) {
	tr := &TLSResult{
		Host: "example.test",
		Port: 8443,
		Versions: []TLSVersionInfo{
			{Version: TLSVersionTLS10, Name: "TLS 1.0", Supported: false, Error: "protocol version not supported"},
			{Version: TLSVersionTLS11, Name: "TLS 1.1", Supported: false, Error: "protocol version not supported"},
			{Version: TLSVersionTLS12, Name: "TLS 1.2", Supported: true, CipherSuite: tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
			{Version: TLSVersionTLS13, Name: "TLS 1.3", Supported: true, CipherSuite: tls.TLS_AES_128_GCM_SHA256},
		},
		MinSupported: TLSVersionTLS12,
		MaxSupported: TLSVersionTLS13,
	}

	j := tr.ToJSON()
	if j.Host != "example.test" || j.Port != 8443 {
		t.Errorf("host/port = %q/%d", j.Host, j.Port)
	}
	if len(j.Versions) != 4 {
		t.Fatalf("versions = %d, want 4", len(j.Versions))
	}

	wantVersions := []JSONTLSVersionInfo{
		{Version: "0x0301", Name: "TLS 1.0", Supported: false, Error: "protocol version not supported"},
		{Version: "0x0302", Name: "TLS 1.1", Supported: false, Error: "protocol version not supported"},
		{Version: "0x0303", Name: "TLS 1.2", Supported: true, CipherSuite: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		{Version: "0x0304", Name: "TLS 1.3", Supported: true, CipherSuite: "TLS_AES_128_GCM_SHA256"},
	}
	if !reflect.DeepEqual(j.Versions, wantVersions) {
		t.Errorf("Versions = %+v\nwant      %+v", j.Versions, wantVersions)
	}
	if j.MinSupported != "TLS 1.2" || j.MaxSupported != "TLS 1.3" {
		t.Errorf("min/max = %q/%q, want TLS 1.2/TLS 1.3", j.MinSupported, j.MaxSupported)
	}

	// An unsupported version with a stale cipher suite must not report one.
	stale := &TLSResult{Versions: []TLSVersionInfo{{Version: TLSVersionTLS12, Supported: false, CipherSuite: tls.TLS_AES_128_GCM_SHA256}}}
	if cs := stale.ToJSON().Versions[0].CipherSuite; cs != "" {
		t.Errorf("unsupported version reported cipher suite %q", cs)
	}

	// Nothing supported: min/max stay empty.
	none := &TLSResult{Host: "h", Port: 1}
	nj := none.ToJSON()
	if nj.MinSupported != "" || nj.MaxSupported != "" {
		t.Errorf("min/max should be empty when nothing is supported, got %q/%q", nj.MinSupported, nj.MaxSupported)
	}
	if nj.Versions == nil || len(nj.Versions) != 0 {
		t.Errorf("Versions should be an empty non-nil slice, got %#v", nj.Versions)
	}
}

func TestTLSResultMarshalJSON(t *testing.T) {
	tr := &TLSResult{
		Host:         "example.test",
		Port:         443,
		Versions:     []TLSVersionInfo{{Version: TLSVersionTLS13, Name: "TLS 1.3", Supported: true, CipherSuite: tls.TLS_AES_256_GCM_SHA384}},
		MinSupported: TLSVersionTLS13,
		MaxSupported: TLSVersionTLS13,
	}
	viaMarshaler, err := json.Marshal(tr)
	if err != nil {
		t.Fatal(err)
	}
	viaToJSON, err := json.Marshal(tr.ToJSON())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(viaMarshaler, viaToJSON) {
		t.Errorf("MarshalJSON output differs from ToJSON:\n%s\n%s", viaMarshaler, viaToJSON)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(viaMarshaler, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"host", "port", "versions", "min_supported", "max_supported"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("JSON missing key %q", key)
		}
	}
}

func TestVerificationResultToJSON(t *testing.T) {
	c, err := InspectFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatalf("InspectFile: %v", err)
	}

	t.Run("key not checked omits key_matches", func(t *testing.T) {
		vr := &VerificationResult{Certificate: c, IsValid: true, Errors: []string{}, Warnings: []string{"w1"}}
		j := vr.ToJSON()
		if j.KeyMatches != nil {
			t.Errorf("KeyMatches should be nil when KeyChecked is false, got %v", *j.KeyMatches)
		}
		if !j.IsValid {
			t.Error("IsValid not carried through")
		}
		if !reflect.DeepEqual(j.Warnings, []string{"w1"}) {
			t.Errorf("Warnings = %v", j.Warnings)
		}
		if j.Certificate.Subject.CommonName != c.Subject.CommonName {
			t.Errorf("embedded certificate CN = %q, want %q", j.Certificate.Subject.CommonName, c.Subject.CommonName)
		}

		raw, err := json.Marshal(vr)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte(`"key_matches"`)) {
			t.Error("key_matches should be omitted from JSON when the key was not checked")
		}
	})

	for _, matches := range []bool{true, false} {
		matches := matches
		t.Run(fmt.Sprintf("key checked matches=%v", matches), func(t *testing.T) {
			vr := &VerificationResult{
				Certificate: c,
				IsValid:     matches,
				Errors:      []string{"e1", "e2"},
				KeyChecked:  true,
				KeyMatches:  matches,
			}
			j := vr.ToJSON()
			if j.KeyMatches == nil {
				t.Fatal("KeyMatches should be set when KeyChecked is true")
			}
			if *j.KeyMatches != matches {
				t.Errorf("KeyMatches = %v, want %v", *j.KeyMatches, matches)
			}
			if !reflect.DeepEqual(j.Errors, []string{"e1", "e2"}) {
				t.Errorf("Errors = %v", j.Errors)
			}

			viaMarshaler, err := json.Marshal(vr)
			if err != nil {
				t.Fatal(err)
			}
			viaToJSON, err := json.Marshal(vr.ToJSON())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(viaMarshaler, viaToJSON) {
				t.Error("VerificationResult.MarshalJSON differs from ToJSON")
			}
			want := fmt.Sprintf(`"key_matches":%v`, matches)
			if !bytes.Contains(viaMarshaler, []byte(want)) {
				t.Errorf("JSON missing %s:\n%s", want, viaMarshaler)
			}
		})
	}
}

func TestCertificateAndCSRMarshalJSON(t *testing.T) {
	c, err := InspectFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatalf("InspectFile: %v", err)
	}
	viaMarshaler, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	viaToJSON, err := json.Marshal(c.ToJSON())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(viaMarshaler, viaToJSON) {
		t.Error("Certificate.MarshalJSON differs from ToJSON")
	}
	for _, key := range []string{`"subject"`, `"serial_number"`, `"fingerprint_sha256"`, `"fingerprint_sha1"`, `"source"`, `"format"`} {
		if !bytes.Contains(viaMarshaler, []byte(key)) {
			t.Errorf("certificate JSON missing %s", key)
		}
	}

	info := &CSRInfo{SignatureAlgorithm: "SHA256-RSA", PublicKeyAlgorithm: "RSA", KeySize: 2048, SANs: []string{"a.test", "IP:10.0.0.1"}}
	viaMarshaler, err = json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	viaToJSON, err = json.Marshal(info.ToJSON())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(viaMarshaler, viaToJSON) {
		t.Error("CSRInfo.MarshalJSON differs from ToJSON")
	}
	if !bytes.Contains(viaMarshaler, []byte(`"ip_addresses":["10.0.0.1"]`)) {
		t.Errorf("CSR JSON missing routed IP SAN:\n%s", viaMarshaler)
	}
}

func TestCertificateToJSONRemoteFields(t *testing.T) {
	host, port := startTLSServer(t, tls.VersionTLS12, tls.VersionTLS13)
	c, _, err := InspectURLWithOptions(host, port, "", testDialTimeout, "auto")
	if err != nil {
		t.Fatalf("InspectURLWithOptions: %v", err)
	}

	j := c.ToJSON()
	if j.TLSVersion != "TLS 1.3" {
		t.Errorf("tls_version = %q, want TLS 1.3", j.TLSVersion)
	}
	if j.CipherSuite == "" {
		t.Error("cipher_suite should be set for a remote certificate")
	}
	if !reflect.DeepEqual(j.IPAddresses, []string{"127.0.0.1"}) {
		t.Errorf("ip_addresses = %v, want [127.0.0.1]", j.IPAddresses)
	}
	if !reflect.DeepEqual(j.DNSNames, []string{"localhost"}) {
		t.Errorf("dns_names = %v, want [localhost]", j.DNSNames)
	}
	if j.Format != FormatDER || j.Source != "https://"+host {
		t.Errorf("format/source = %q/%q", j.Format, j.Source)
	}
	if !reflect.DeepEqual(j.KeyUsage, []string{"Digital Signature", "Key Encipherment"}) {
		t.Errorf("key_usage = %v", j.KeyUsage)
	}
	if !reflect.DeepEqual(j.ExtKeyUsage, []string{"Server Authentication"}) {
		t.Errorf("ext_key_usage = %v", j.ExtKeyUsage)
	}

	// A file-based certificate must not carry TLS fields.
	fileCert, err := InspectFile(testutil.TestdataPath("valid.pem"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(fileCert)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(`"tls_version"`)) || bytes.Contains(raw, []byte(`"cipher_suite"`)) {
		t.Error("file certificate JSON should omit tls_version and cipher_suite")
	}
}
