package controller

import (
	"poolx/internal/model"
	"poolx/internal/service"

	"github.com/gin-gonic/gin"
)

type mihomoDownloadRequest struct {
	InstallPath string `json:"install_path"`
}

type mihomoInspectRequest struct {
	InstallPath string `json:"install_path"`
}

// InspectMihomoBinary godoc
// @Summary Inspect existing Mihomo binary path and verify version
// @Tags Kernel
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/kernel/mihomo/inspect [post]
func InspectMihomoBinary(c *gin.Context) {
	var request mihomoInspectRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "invalid request payload")
		return
	}

	result, err := service.InspectMihomoBinary(c.Request.Context(), request.InstallPath)
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelError, "mihomo binary inspect failed | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelInfo, "mihomo binary inspected | path="+result.InstallPath+" | version="+result.DetectedVersion)
	c.JSON(200, gin.H{
		"success": true,
		"message": "Mihomo 二进制检查通过。",
		"data":    result,
	})
}

// UploadMihomoBinary godoc
// @Summary Upload Mihomo binary, install it to target path and verify version
// @Tags Kernel
// @Accept mpfd
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/kernel/mihomo/upload [post]
func UploadMihomoBinary(c *gin.Context) {
	fileHeader, err := c.FormFile("binary")
	if err != nil {
		respondFailure(c, "请先选择要上传的 Mihomo 二进制文件。")
		return
	}

	installPath := c.PostForm("install_path")
	file, err := fileHeader.Open()
	if err != nil {
		respondFailure(c, "读取上传文件失败。")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	result, err := service.InstallUploadedMihomoBinary(c.Request.Context(), fileHeader.Filename, installPath, file)
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelError, "mihomo binary upload failed | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelInfo, "mihomo binary uploaded | path="+result.InstallPath+" | version="+result.DetectedVersion)
	c.JSON(200, gin.H{
		"success": true,
		"message": "Mihomo 二进制已上传并完成版本校验。",
		"data":    result,
	})
}

// DownloadMihomoBinary godoc
// @Summary Download latest official Mihomo binary for current platform and verify version
// @Tags Kernel
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/kernel/mihomo/download [post]
func DownloadMihomoBinary(c *gin.Context) {
	var request mihomoDownloadRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "invalid request payload")
		return
	}

	result, err := service.DownloadAndInstallMihomoBinary(c.Request.Context(), request.InstallPath)
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelError, "mihomo binary download failed | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelInfo, "mihomo binary downloaded | path="+result.InstallPath+" | version="+result.DetectedVersion)
	c.JSON(200, gin.H{
		"success": true,
		"message": "已从官方仓库下载 Mihomo 并完成版本校验。",
		"data":    result,
	})
}
