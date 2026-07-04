package handler

import (
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// LicenseHandler handles standalone license APIs that do not use system login.
type LicenseHandler struct {
	licenseService *service.LicenseService
}

// NewLicenseHandler creates a standalone license handler.
func NewLicenseHandler(licenseService *service.LicenseService) *LicenseHandler {
	return &LicenseHandler{licenseService: licenseService}
}

type licenseActivateRequest struct {
	ActivationCode string `json:"activationCode" binding:"required"`
	Product        string `json:"product" binding:"required"`
	ProductBatch   string `json:"productBatch" binding:"required"`
	USBFingerprint string `json:"usbFingerprint" binding:"required"`
}

type licenseVerifyRequest struct {
	LicenseID      string `json:"licenseId" binding:"required"`
	CodeID         string `json:"codeId" binding:"required"`
	Product        string `json:"product" binding:"required"`
	USBFingerprint string `json:"usbFingerprint" binding:"required"`
}

type licenseDeactivateRequest struct {
	LicenseID string `json:"licenseId" binding:"required"`
	CodeID    string `json:"codeId" binding:"required"`
}

type licenseCreateCodesRequest struct {
	Count        int      `json:"count" binding:"required"`
	Product      string   `json:"product"`
	ProductBatch string   `json:"productBatch"`
	Features     []string `json:"features"`
	Prefix       string   `json:"prefix"`
	ExpiresAt    string   `json:"expiresAt"`
}

type licenseUpdateFeaturesRequest struct {
	Features []string `json:"features"`
}

type licenseCodeResponse struct {
	ID             int64    `json:"id"`
	CodeID         string   `json:"codeId"`
	Code           string   `json:"code"`
	LicenseID      string   `json:"licenseId"`
	Product        string   `json:"product"`
	ProductBatch   string   `json:"productBatch"`
	Features       []string `json:"features"`
	Status         string   `json:"status"`
	USBFingerprint string   `json:"usbFingerprint"`
	ActivatedAt    string   `json:"activatedAt"`
	LastVerifiedAt string   `json:"lastVerifiedAt"`
	ExpiresAt      string   `json:"expiresAt"`
	RevokedAt      string   `json:"revokedAt"`
	RevokedReason  string   `json:"revokedReason"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

// Activate activates a standalone license code.
func (h *LicenseHandler) Activate(c *gin.Context) {
	var req licenseActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.licenseService.Activate(c.Request.Context(), service.LicenseActivateInput{
		ActivationCode: req.ActivationCode,
		Product:        req.Product,
		ProductBatch:   req.ProductBatch,
		USBFingerprint: req.USBFingerprint,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// Verify verifies a standalone license and returns a session token.
func (h *LicenseHandler) Verify(c *gin.Context) {
	var req licenseVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.licenseService.Verify(c.Request.Context(), service.LicenseVerifyInput{
		LicenseID:      req.LicenseID,
		CodeID:         req.CodeID,
		Product:        req.Product,
		USBFingerprint: req.USBFingerprint,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// Deactivate revokes a standalone license from the client side.
func (h *LicenseHandler) Deactivate(c *gin.Context) {
	var req licenseDeactivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.licenseService.Deactivate(c.Request.Context(), service.LicenseDeactivateInput{
		LicenseID: req.LicenseID,
		CodeID:    req.CodeID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, licenseCodeToResponse(*result))
}

// PublicKey returns the standalone license public key.
func (h *LicenseHandler) PublicKey(c *gin.Context) {
	publicKey, err := h.licenseService.PublicKeyPEM()
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"publicKeyPem": publicKey})
}

// CreateCodes creates standalone activation codes for system administrators.
func (h *LicenseHandler) CreateCodes(c *gin.Context) {
	var req licenseCreateCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	expiresAt, err := parseLicenseExpiresAt(req.ExpiresAt)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	codes, err := h.licenseService.CreateCodes(c.Request.Context(), service.LicenseCreateCodesInput{
		Count:        req.Count,
		Product:      req.Product,
		ProductBatch: req.ProductBatch,
		Features:     req.Features,
		Prefix:       req.Prefix,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"codes": licenseCodesToResponse(codes)})
}

// ListCodes lists standalone activation codes for system administrators.
func (h *LicenseHandler) ListCodes(c *gin.Context) {
	codes, err := h.licenseService.ListCodes(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"codes": licenseCodesToResponse(codes)})
}

// DisableCode disables one standalone activation code.
func (h *LicenseHandler) DisableCode(c *gin.Context) {
	h.setCodeStatus(c, service.LicenseStatusDisabled)
}

// EnableCode enables one standalone activation code.
func (h *LicenseHandler) EnableCode(c *gin.Context) {
	h.setCodeStatus(c, "enable")
}

// RefundCode marks one standalone activation code as refunded.
func (h *LicenseHandler) RefundCode(c *gin.Context) {
	h.setCodeStatus(c, service.LicenseStatusRefunded)
}

// RevokeLicense revokes one standalone license.
func (h *LicenseHandler) RevokeLicense(c *gin.Context) {
	code, err := h.licenseService.RevokeLicense(c.Request.Context(), c.Param("license_id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, licenseCodeToResponse(*code))
}

// UpdateFeatures replaces the feature set on one activation code.
func (h *LicenseHandler) UpdateFeatures(c *gin.Context) {
	var req licenseUpdateFeaturesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	code, err := h.licenseService.UpdateCodeFeatures(c.Request.Context(), c.Param("code_id"), req.Features)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, licenseCodeToResponse(*code))
}

// setCodeStatus updates one code status through the system admin route.
func (h *LicenseHandler) setCodeStatus(c *gin.Context, status string) {
	code, err := h.licenseService.SetCodeStatus(c.Request.Context(), c.Param("code_id"), status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, licenseCodeToResponse(*code))
}

// parseLicenseExpiresAt parses an optional RFC3339 expiration timestamp.
func parseLicenseExpiresAt(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// licenseCodesToResponse maps service license codes to JSON responses.
func licenseCodesToResponse(codes []service.LicenseCode) []licenseCodeResponse {
	out := make([]licenseCodeResponse, 0, len(codes))
	for i := range codes {
		out = append(out, licenseCodeToResponse(codes[i]))
	}
	return out
}

// licenseCodeToResponse maps one service license code to a JSON response.
func licenseCodeToResponse(code service.LicenseCode) licenseCodeResponse {
	return licenseCodeResponse{
		ID:             code.ID,
		CodeID:         code.CodeID,
		Code:           code.Code,
		LicenseID:      code.LicenseID,
		Product:        code.Product,
		ProductBatch:   code.ProductBatch,
		Features:       code.Features,
		Status:         code.Status,
		USBFingerprint: code.USBFingerprint,
		ActivatedAt:    licenseTimeForResponse(code.ActivatedAt),
		LastVerifiedAt: licenseTimeForResponse(code.LastVerifiedAt),
		ExpiresAt:      licenseTimeForResponse(code.ExpiresAt),
		RevokedAt:      licenseTimeForResponse(code.RevokedAt),
		RevokedReason:  code.RevokedReason,
		CreatedAt:      code.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      code.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// licenseTimeForResponse formats nullable license timestamps for JSON.
func licenseTimeForResponse(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
