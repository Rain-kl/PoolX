package controller

import (
	"poolx/internal/service"

	"github.com/gin-gonic/gin"
)

// GetKernelCapability godoc
// @Summary Get current kernel capability negotiation result
// @Tags Capability
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/capabilities [get]
func GetKernelCapability(c *gin.Context) {
	respondSuccess(c, service.GetKernelCapability())
}
