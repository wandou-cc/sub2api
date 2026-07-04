package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterLicenseRoutes registers standalone license APIs without system login middleware.
func RegisterLicenseRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	license := v1.Group("/license")
	{
		license.GET("/public-key", h.License.PublicKey)
		license.POST("/activate", h.License.Activate)
		license.POST("/verify", h.License.Verify)
		license.POST("/deactivate", h.License.Deactivate)
	}
}

// RegisterAdminLicenseRoutes registers license management APIs under admin auth.
func RegisterAdminLicenseRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	license := admin.Group("/license")
	{
		license.POST("/codes", h.License.CreateCodes)
		license.GET("/codes", h.License.ListCodes)
		license.POST("/codes/:code_id/disable", h.License.DisableCode)
		license.POST("/codes/:code_id/enable", h.License.EnableCode)
		license.POST("/codes/:code_id/refund", h.License.RefundCode)
		license.PUT("/codes/:code_id/features", h.License.UpdateFeatures)
		license.POST("/licenses/:license_id/revoke", h.License.RevokeLicense)
	}
}
