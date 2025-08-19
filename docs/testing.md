# Testing Guide

## Running Tests

### Run all tests
```bash
make test
```

### Run tests with coverage
```bash
make test-coverage
```

### Generate HTML coverage report
```bash
make test-coverage-html
```

### Run specific package tests
```bash
go test -v certwiz/pkg/cert
go test -v certwiz/pkg/ui
go test -v certwiz/cmd
```

## Test Structure

The certwiz project has comprehensive test coverage across all major components:

### Unit Tests

#### pkg/cert (cert_test.go)
- `TestInspectFile` - Tests certificate file inspection for PEM/DER formats
- `TestParseCertificate` - Tests certificate parsing logic
- `TestGenerate` - Tests self-signed certificate generation
- `TestConvert` - Tests format conversion between PEM and DER
- `TestVerify` - Tests certificate verification
- `TestCertificateExpiry` - Tests expiry date calculations
- Benchmark tests for performance optimization

#### pkg/ui (ui_test.go)
- `TestFormatSubject` - Tests certificate subject formatting
- `TestFormatPublicKey` - Tests public key display formatting
- `TestFormatSANs` - Tests Subject Alternative Names formatting
- `TestFormatStatus` - Tests certificate status display
- `TestDisplayCertificate` - Tests full certificate display
- `TestDisplayCertificateChain` - Tests certificate chain display
- Various helper function tests

### Integration Tests

#### cmd package
- `TestRootCommand` - Tests root command and help
- `TestInspectCommand` - Tests inspect command with various inputs
- `TestGenerateCommand` - Tests certificate generation command
- Command flag validation tests

## Test Fixtures

Test certificates are generated using the `testdata/generate_test_certs.sh` script:

- `valid.pem/der` - Valid test certificate
- `expired.pem` - Expired certificate
- `many-sans.pem` - Certificate with multiple SANs
- `strong.pem` - 4096-bit RSA certificate
- `invalid.pem` - Corrupted certificate for error testing
- `fullchain.pem` - Complete certificate chain for chain testing

To regenerate test certificates:
```bash
make test-generate-certs
```

## Coverage Goals

Current test coverage:
- pkg/cert: ~62.5%
- pkg/ui: ~61.2%

Target coverage: 80% for all packages

## Continuous Integration

GitHub Actions workflows are configured for:
- Running tests on multiple OS (Linux, macOS, Windows)
- Multiple Go versions (1.20, 1.21)
- Code coverage reporting via Codecov
- Linting with golangci-lint

See `.github/workflows/test.yml` for CI configuration.

## Writing Tests

When adding new features:
1. Write unit tests for the core logic
2. Add integration tests for commands
3. Update test fixtures if needed
4. Ensure tests pass locally before pushing
5. Check coverage with `make test-coverage`

### Test Best Practices
- Use table-driven tests for multiple scenarios
- Mock external dependencies (network, filesystem where appropriate)
- Test both success and error cases
- Include benchmark tests for performance-critical code
- Keep tests focused and independent