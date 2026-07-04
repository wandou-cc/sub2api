package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	LicenseStatusUnused   = "unused"
	LicenseStatusActive   = "active"
	LicenseStatusDisabled = "disabled"
	LicenseStatusExpired  = "expired"
	LicenseStatusRevoked  = "revoked"
	LicenseStatusRefunded = "refunded"
)

var (
	ErrLicenseCodeNotFound        = infraerrors.NotFound("LICENSE_CODE_NOT_FOUND", "activation code not found")
	ErrLicenseNotFound            = infraerrors.NotFound("LICENSE_NOT_FOUND", "license not found")
	ErrLicenseProductMismatch     = infraerrors.Forbidden("LICENSE_PRODUCT_MISMATCH", "license product mismatch")
	ErrLicenseNotActive           = infraerrors.Forbidden("LICENSE_NOT_ACTIVE", "license is not active")
	ErrLicenseFingerprintMismatch = infraerrors.Forbidden("LICENSE_FINGERPRINT_MISMATCH", "USB fingerprint mismatch")
	ErrLicensePrivateKeyMissing   = infraerrors.InternalServer("LICENSE_PRIVATE_KEY_MISSING", "license private key is not configured")
	ErrLicensePrivateKeyInvalid   = infraerrors.InternalServer("LICENSE_PRIVATE_KEY_INVALID", "license private key is invalid")
	ErrLicenseCodeConflict        = infraerrors.Conflict("LICENSE_CODE_CONFLICT", "activation code already exists")
)

// LicenseRepository defines persistence operations for the standalone license module.
type LicenseRepository interface {
	CreateCodes(ctx context.Context, codes []LicenseCode) error
	ListCodes(ctx context.Context) ([]LicenseCode, error)
	Activate(ctx context.Context, input LicenseActivateInput, licenseID string, now time.Time) (*LicenseCode, error)
	Verify(ctx context.Context, input LicenseVerifyInput, now time.Time) (*LicenseCode, error)
	Deactivate(ctx context.Context, input LicenseDeactivateInput, now time.Time) (*LicenseCode, error)
	SetCodeStatus(ctx context.Context, codeID, status string, now time.Time) (*LicenseCode, error)
	UpdateCodeFeatures(ctx context.Context, codeID string, features []string, now time.Time) (*LicenseCode, error)
	RevokeLicense(ctx context.Context, licenseID string, now time.Time) (*LicenseCode, error)
}

// LicenseCode represents one activation code and its bound license record.
type LicenseCode struct {
	ID             int64
	CodeID         string
	Code           string
	LicenseID      string
	Product        string
	ProductBatch   string
	Features       []string
	Status         string
	USBFingerprint string
	ActivatedAt    *time.Time
	LastVerifiedAt *time.Time
	ExpiresAt      *time.Time
	RevokedAt      *time.Time
	RevokedReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// LicenseCreateCodesInput contains admin activation-code generation parameters.
type LicenseCreateCodesInput struct {
	Count        int
	Product      string
	ProductBatch string
	Features     []string
	Prefix       string
	ExpiresAt    *time.Time
}

// LicenseActivateInput contains client activation data.
type LicenseActivateInput struct {
	ActivationCode string
	Product        string
	ProductBatch   string
	USBFingerprint string
}

// LicenseVerifyInput contains client verification data.
type LicenseVerifyInput struct {
	LicenseID      string
	CodeID         string
	Product        string
	USBFingerprint string
}

// LicenseDeactivateInput contains client deactivation data.
type LicenseDeactivateInput struct {
	LicenseID string
	CodeID    string
}

// LicensePayload is the exact payload signed and returned to clients.
type LicensePayload struct {
	LicenseID      string   `json:"licenseId"`
	CodeID         string   `json:"codeId"`
	Product        string   `json:"product"`
	USBFingerprint string   `json:"usbFingerprint"`
	Features       []string `json:"features"`
	ActivatedAt    string   `json:"activatedAt"`
	LastVerifiedAt string   `json:"lastVerifiedAt"`
}

// SignedLicense is the activation response persisted by clients.
type SignedLicense struct {
	LicenseID      string   `json:"licenseId"`
	CodeID         string   `json:"codeId"`
	Product        string   `json:"product"`
	USBFingerprint string   `json:"usbFingerprint"`
	Features       []string `json:"features"`
	ActivatedAt    string   `json:"activatedAt"`
	LastVerifiedAt string   `json:"lastVerifiedAt"`
	Signature      string   `json:"signature"`
}

// LicenseVerifyResult is returned when an active license is verified.
type LicenseVerifyResult struct {
	OK               bool     `json:"ok"`
	Status           string   `json:"status"`
	Features         []string `json:"features"`
	SessionToken     string   `json:"sessionToken"`
	SessionExpiresAt string   `json:"sessionExpiresAt"`
	MinClientVersion string   `json:"minClientVersion"`
	LatestVersion    string   `json:"latestVersion"`
	UpdateURL        string   `json:"updateUrl"`
}

// LicenseService implements standalone product-license workflows.
type LicenseService struct {
	repo LicenseRepository
	cfg  *config.Config
}

// NewLicenseService creates a standalone license service.
func NewLicenseService(repo LicenseRepository, cfg *config.Config) *LicenseService {
	return &LicenseService{repo: repo, cfg: cfg}
}

// CreateCodes creates activation codes for the standalone license module.
func (s *LicenseService) CreateCodes(ctx context.Context, input LicenseCreateCodesInput) ([]LicenseCode, error) {
	if input.Count < 1 || input.Count > 1000 {
		return nil, infraerrors.BadRequest("LICENSE_CODE_COUNT_INVALID", "count must be 1-1000")
	}
	product := strings.TrimSpace(input.Product)
	if product == "" {
		product = s.cfg.License.ProductID
	}
	productBatch := strings.TrimSpace(input.ProductBatch)
	if productBatch == "" {
		productBatch = s.cfg.License.ProductBatch
	}
	features := input.Features
	if len(features) == 0 {
		features = s.cfg.License.DefaultFeatures
	}
	prefix := strings.TrimSpace(input.Prefix)
	if prefix == "" {
		prefix = "UCLAW"
	}

	codes := make([]LicenseCode, 0, input.Count)
	now := time.Now().UTC()
	for i := 0; i < input.Count; i++ {
		activationCode, err := newActivationCode(prefix)
		if err != nil {
			return nil, err
		}
		codeID, err := newLicenseID("code")
		if err != nil {
			return nil, err
		}
		codes = append(codes, LicenseCode{
			CodeID:       codeID,
			Code:         activationCode,
			Product:      product,
			ProductBatch: productBatch,
			Features:     features,
			Status:       LicenseStatusUnused,
			ExpiresAt:    input.ExpiresAt,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if err := s.repo.CreateCodes(ctx, codes); err != nil {
		return nil, err
	}
	return codes, nil
}

// ListCodes returns all standalone license activation codes.
func (s *LicenseService) ListCodes(ctx context.Context) ([]LicenseCode, error) {
	return s.repo.ListCodes(ctx)
}

// Activate activates an unused code or returns the existing binding for the same USB fingerprint.
func (s *LicenseService) Activate(ctx context.Context, input LicenseActivateInput) (*SignedLicense, error) {
	if err := requireLicenseActivationInput(input); err != nil {
		return nil, err
	}
	licenseID, err := newLicenseID("lic")
	if err != nil {
		return nil, err
	}
	code, err := s.repo.Activate(ctx, input, licenseID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return s.signedLicense(code)
}

// Verify checks license binding and returns a short-lived session token.
func (s *LicenseService) Verify(ctx context.Context, input LicenseVerifyInput) (*LicenseVerifyResult, error) {
	if err := requireLicenseVerifyInput(input); err != nil {
		return nil, err
	}
	code, err := s.repo.Verify(ctx, input, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	sessionToken, err := newSessionToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.License.SessionTTLSeconds) * time.Second)
	return &LicenseVerifyResult{
		OK:               true,
		Status:           LicenseStatusActive,
		Features:         code.Features,
		SessionToken:     sessionToken,
		SessionExpiresAt: licenseTime(expiresAt),
		MinClientVersion: s.cfg.License.MinClientVersion,
		LatestVersion:    s.cfg.License.LatestVersion,
		UpdateURL:        s.cfg.License.UpdateURL,
	}, nil
}

// Deactivate revokes the current license binding.
func (s *LicenseService) Deactivate(ctx context.Context, input LicenseDeactivateInput) (*LicenseCode, error) {
	if strings.TrimSpace(input.LicenseID) == "" || strings.TrimSpace(input.CodeID) == "" {
		return nil, infraerrors.BadRequest("LICENSE_DEACTIVATE_INVALID", "licenseId and codeId are required")
	}
	return s.repo.Deactivate(ctx, input, time.Now().UTC())
}

// SetCodeStatus changes an activation code status from the standalone admin API.
func (s *LicenseService) SetCodeStatus(ctx context.Context, codeID, status string) (*LicenseCode, error) {
	switch status {
	case LicenseStatusDisabled, LicenseStatusRefunded, LicenseStatusRevoked, "enable":
	default:
		return nil, infraerrors.BadRequest("LICENSE_STATUS_INVALID", "license status is invalid")
	}
	return s.repo.SetCodeStatus(ctx, codeID, status, time.Now().UTC())
}

// UpdateCodeFeatures replaces the feature set granted by one activation code.
func (s *LicenseService) UpdateCodeFeatures(ctx context.Context, codeID string, features []string) (*LicenseCode, error) {
	codeID = strings.TrimSpace(codeID)
	if codeID == "" {
		return nil, infraerrors.BadRequest("LICENSE_CODE_ID_REQUIRED", "codeId is required")
	}
	normalized := normalizeLicenseFeatures(features)
	if len(normalized) == 0 {
		return nil, infraerrors.BadRequest("LICENSE_FEATURES_REQUIRED", "at least one feature is required")
	}
	return s.repo.UpdateCodeFeatures(ctx, codeID, normalized, time.Now().UTC())
}

// normalizeLicenseFeatures trims, deduplicates, and drops empty feature strings.
func normalizeLicenseFeatures(features []string) []string {
	seen := make(map[string]struct{}, len(features))
	out := make([]string, 0, len(features))
	for _, feature := range features {
		feature = strings.TrimSpace(feature)
		if feature == "" {
			continue
		}
		if _, ok := seen[feature]; ok {
			continue
		}
		seen[feature] = struct{}{}
		out = append(out, feature)
	}
	return out
}

// RevokeLicense revokes one license from the standalone admin API.
func (s *LicenseService) RevokeLicense(ctx context.Context, licenseID string) (*LicenseCode, error) {
	if strings.TrimSpace(licenseID) == "" {
		return nil, infraerrors.BadRequest("LICENSE_ID_REQUIRED", "licenseId is required")
	}
	return s.repo.RevokeLicense(ctx, licenseID, time.Now().UTC())
}

// PublicKeyPEM returns the public key derived from the configured Ed25519 private key.
func (s *LicenseService) PublicKeyPEM() (string, error) {
	privateKey, err := s.privateKey()
	if err != nil {
		return "", err
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", ErrLicensePrivateKeyInvalid.WithCause(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
}

// signedLicense signs the client license payload with Ed25519.
func (s *LicenseService) signedLicense(code *LicenseCode) (*SignedLicense, error) {
	privateKey, err := s.privateKey()
	if err != nil {
		return nil, err
	}
	payload := LicensePayload{
		LicenseID:      code.LicenseID,
		CodeID:         code.CodeID,
		Product:        code.Product,
		USBFingerprint: code.USBFingerprint,
		Features:       code.Features,
		ActivatedAt:    licenseTimePtr(code.ActivatedAt),
		LastVerifiedAt: licenseTimePtr(code.LastVerifiedAt),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	signature := ed25519.Sign(privateKey, payloadBytes)
	return &SignedLicense{
		LicenseID:      payload.LicenseID,
		CodeID:         payload.CodeID,
		Product:        payload.Product,
		USBFingerprint: payload.USBFingerprint,
		Features:       payload.Features,
		ActivatedAt:    payload.ActivatedAt,
		LastVerifiedAt: payload.LastVerifiedAt,
		Signature:      base64.RawURLEncoding.EncodeToString(signature),
	}, nil
}

// privateKey loads and parses the configured Ed25519 private key.
func (s *LicenseService) privateKey() (ed25519.PrivateKey, error) {
	pemText := strings.TrimSpace(s.cfg.License.PrivateKeyPEM)
	if pemText == "" {
		if strings.TrimSpace(s.cfg.License.PrivateKeyFile) == "" {
			return nil, ErrLicensePrivateKeyMissing
		}
		data, err := os.ReadFile(s.cfg.License.PrivateKeyFile)
		if err != nil {
			return nil, ErrLicensePrivateKeyInvalid.WithCause(err)
		}
		pemText = string(data)
	}
	pemText = strings.ReplaceAll(pemText, `\n`, "\n")
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, ErrLicensePrivateKeyInvalid
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrLicensePrivateKeyInvalid.WithCause(err)
	}
	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrLicensePrivateKeyInvalid
	}
	return privateKey, nil
}

// requireLicenseActivationInput validates required activation fields.
func requireLicenseActivationInput(input LicenseActivateInput) error {
	if strings.TrimSpace(input.ActivationCode) == "" || strings.TrimSpace(input.Product) == "" ||
		strings.TrimSpace(input.ProductBatch) == "" || strings.TrimSpace(input.USBFingerprint) == "" {
		return infraerrors.BadRequest("LICENSE_ACTIVATE_INVALID", "activationCode, product, productBatch and usbFingerprint are required")
	}
	return nil
}

// requireLicenseVerifyInput validates required verification fields.
func requireLicenseVerifyInput(input LicenseVerifyInput) error {
	if strings.TrimSpace(input.LicenseID) == "" || strings.TrimSpace(input.CodeID) == "" ||
		strings.TrimSpace(input.Product) == "" || strings.TrimSpace(input.USBFingerprint) == "" {
		return infraerrors.BadRequest("LICENSE_VERIFY_INVALID", "licenseId, codeId, product and usbFingerprint are required")
	}
	return nil
}

// newLicenseID creates a prefixed random license identifier.
func newLicenseID(prefix string) (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate license id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(bytes), nil
}

// newActivationCode creates a user-facing activation code.
func newActivationCode(prefix string) (string, error) {
	left := make([]byte, 4)
	right := make([]byte, 4)
	if _, err := rand.Read(left); err != nil {
		return "", fmt.Errorf("generate activation code: %w", err)
	}
	if _, err := rand.Read(right); err != nil {
		return "", fmt.Errorf("generate activation code: %w", err)
	}
	return strings.ToUpper(prefix) + "-" + strings.ToUpper(hex.EncodeToString(left)) + "-" + strings.ToUpper(hex.EncodeToString(right)), nil
}

// newSessionToken creates a short-lived client session token.
func newSessionToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate license session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// licenseTimePtr formats a nullable time for client license payloads.
func licenseTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return licenseTime(*value)
}

// licenseTime formats time in the same millisecond ISO shape used by the standalone server.
func licenseTime(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.000Z")
}
