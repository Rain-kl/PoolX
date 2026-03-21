package controller

import (
	"context"
	"poolx/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRuntimeStatus godoc
// @Summary Get Mihomo runtime status
// @Tags Runtime
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/runtime/status [get]
func GetRuntimeStatus(c *gin.Context) {
	status, err := service.GetRuntimeStatus(c.Request.Context())
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, status)
}

// StartRuntime godoc
// @Summary Start Mihomo runtime
// @Tags Runtime
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/runtime/start [post]
func StartRuntime(c *gin.Context) {
	status, err := service.StartRuntime(context.Background())
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, status)
}

// StopRuntime godoc
// @Summary Stop Mihomo runtime
// @Tags Runtime
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/runtime/stop [post]
func StopRuntime(c *gin.Context) {
	status, err := service.StopRuntime(context.Background())
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, status)
}

// ReloadRuntime godoc
// @Summary Reload Mihomo runtime
// @Tags Runtime
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/runtime/reload [post]
func ReloadRuntime(c *gin.Context) {
	status, err := service.ReloadRuntime(context.Background())
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, status)
}

// GetRuntimeLogs godoc
// @Summary Get runtime log stream
// @Tags Runtime
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/runtime/logs [get]
func GetRuntimeLogs(c *gin.Context) {
	afterSeq, _ := strconv.ParseInt(c.Query("after_seq"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := service.GetRuntimeLogs(afterSeq, limit)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, items)
}
