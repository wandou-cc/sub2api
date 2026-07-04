package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type licenseRepositoryStub struct {
	activate func(ctx context.Context, input LicenseActivateInput, licenseID string, now time.Time) (*LicenseCode, error)
}

// CreateCodes satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) CreateCodes(ctx context.Context, codes []LicenseCode) error {
	return nil
}

// ListCodes satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) ListCodes(ctx context.Context) ([]LicenseCode, error) {
	return nil, nil
}

// Activate delegates activation behavior for service tests.
func (r licenseRepositoryStub) Activate(ctx context.Context, input LicenseActivateInput, licenseID string, now time.Time) (*LicenseCode, error) {
	return r.activate(ctx, input, licenseID, now)
}

// Verify satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) Verify(ctx context.Context, input LicenseVerifyInput, now time.Time) (*LicenseCode, error) {
	return nil, nil
}

// Deactivate satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) Deactivate(ctx context.Context, input LicenseDeactivateInput, now time.Time) (*LicenseCode, error) {
	return nil, nil
}

// SetCodeStatus satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) SetCodeStatus(ctx context.Context, codeID, status string, now time.Time) (*LicenseCode, error) {
	return nil, nil
}

// UpdateCodeFeatures satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) UpdateCodeFeatures(ctx context.Context, codeID string, features []string, now time.Time) (*LicenseCode, error) {
	return nil, nil
}

// RevokeLicense satisfies LicenseRepository for service tests.
func (r licenseRepositoryStub) RevokeLicense(ctx context.Context, licenseID string, now time.Time) (*LicenseCode, error) {
	return nil, nil
}

// TestLicenseServiceActivateSignsPayload verifies the client license signature contract.
func TestLicenseServiceActivateSignsPayload(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privateKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	activatedAt := time.Date(2026, 7, 3, 10, 11, 12, 0, time.UTC)

	svc := NewLicenseService(licenseRepositoryStub{
		activate: func(ctx context.Context, input LicenseActivateInput, licenseID string, now time.Time) (*LicenseCode, error) {
			require.Equal(t, "UCLAW-ABCD-EFGH", input.ActivationCode)
			require.Equal(t, "uclaw-usb", input.Product)
			require.Equal(t, "dev-2026-06", input.ProductBatch)
			require.Equal(t, "fp-one", input.USBFingerprint)
			require.NotEmpty(t, licenseID)
			return &LicenseCode{
				CodeID:         "code_test",
				LicenseID:      licenseID,
				Product:        input.Product,
				USBFingerprint: input.USBFingerprint,
				Features:       []string{"openmontage"},
				ActivatedAt:    &activatedAt,
			}, nil
		},
	}, &config.Config{
		License: config.LicenseConfig{PrivateKeyPEM: privateKeyPEM},
	})

	license, err := svc.Activate(context.Background(), LicenseActivateInput{
		ActivationCode: "UCLAW-ABCD-EFGH",
		Product:        "uclaw-usb",
		ProductBatch:   "dev-2026-06",
		USBFingerprint: "fp-one",
	})
	require.NoError(t, err)

	payload := LicensePayload{
		LicenseID:      license.LicenseID,
		CodeID:         license.CodeID,
		Product:        license.Product,
		USBFingerprint: license.USBFingerprint,
		Features:       license.Features,
		ActivatedAt:    license.ActivatedAt,
		LastVerifiedAt: license.LastVerifiedAt,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	signature, err := base64.RawURLEncoding.DecodeString(license.Signature)
	require.NoError(t, err)
	require.True(t, ed25519.Verify(publicKey, payloadBytes, signature))
}

// featuresCapturingStub records the features passed to UpdateCodeFeatures.
type featuresCapturingStub struct {
	licenseRepositoryStub
	capturedCodeID string
	captured       []string
}

func (r *featuresCapturingStub) UpdateCodeFeatures(_ context.Context, codeID string, features []string, _ time.Time) (*LicenseCode, error) {
	r.capturedCodeID = codeID
	r.captured = features
	return &LicenseCode{CodeID: codeID, Features: features}, nil
}

// TestLicenseServiceUpdateCodeFeatures verifies validation and normalization.
func TestLicenseServiceUpdateCodeFeatures(t *testing.T) {
	stub := &featuresCapturingStub{}
	svc := NewLicenseService(stub, &config.Config{})

	if _, err := svc.UpdateCodeFeatures(context.Background(), "   ", []string{"openmontage"}); err == nil {
		t.Fatal("expected error for empty codeId")
	}
	if _, err := svc.UpdateCodeFeatures(context.Background(), "code_1", []string{"  "}); err == nil {
		t.Fatal("expected error for empty features")
	}

	code, err := svc.UpdateCodeFeatures(context.Background(), " code_1 ", []string{" openmontage ", "", "video-use", "openmontage"})
	require.NoError(t, err)
	require.Equal(t, "code_1", code.CodeID)
	require.Equal(t, []string{"openmontage", "video-use"}, code.Features)
	require.Equal(t, "code_1", stub.capturedCodeID)
	require.Equal(t, []string{"openmontage", "video-use"}, stub.captured)
}

