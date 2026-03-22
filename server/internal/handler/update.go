package controller

import (
	"errors"
	"io"
	"net/http"
	"poolx/internal/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type confirmManualUpgradeRequest struct {
	UploadToken string `json:"upload_token"`
}

type serverUpgradeRequest struct {
	Channel string `json:"channel"`
}

// GetLatestRelease godoc
// @Summary Get latest GitHub release
// @Tags Update
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/update/latest-release [get]
func GetLatestRelease(c *gin.Context) {
	release, err := service.GetLatestServerRelease(c.Request.Context(), c.Query("channel"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    release,
	})
}

// UpgradeServer godoc
// @Summary Upgrade server binary from latest GitHub release
// @Tags Update
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/update/upgrade [post]
func UpgradeServer(c *gin.Context) {
	var request serverUpgradeRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的参数",
			})
			return
		}
	}
	release, err := service.ScheduleServerUpgrade(request.Channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务升级任务已启动，下载完成后将自动重启。",
		"data":    release,
	})
}

// StreamServerUpgradeLogs godoc
// @Summary Stream server upgrade logs over websocket
// @Tags Update
// @Router /api/update/logs/ws [get]
func StreamServerUpgradeLogs(c *gin.Context) {
	websocket.Handler(func(conn *websocket.Conn) {
		defer func() {
			_ = conn.Close()
		}()

		updates, unsubscribe := service.SubscribeServerUpgradeStream()
		defer unsubscribe()

		heartbeatTicker := time.NewTicker(15 * time.Second)
		defer heartbeatTicker.Stop()

		for {
			select {
			case snapshot, ok := <-updates:
				if !ok {
					return
				}
				if err := websocket.JSON.Send(conn, snapshot); err != nil {
					return
				}
			case <-heartbeatTicker.C:
				if err := websocket.JSON.Send(conn, service.ServerUpgradeStreamSnapshot{}); err != nil {
					return
				}
			case <-c.Request.Context().Done():
				return
			}
		}
	}).ServeHTTP(c.Writer, c.Request)
}

// UploadManualServerBinary godoc
// @Summary Upload server binary and inspect version before upgrade
// @Tags Update
// @Accept mpfd
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/update/manual-upload [post]
func UploadManualServerBinary(c *gin.Context) {
	fileHeader, err := c.FormFile("binary")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先选择要上传的服务端二进制文件。",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "读取上传文件失败。",
		})
		return
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := service.UploadManualServerBinary(c.Request.Context(), fileHeader.Filename, file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	message := strings.TrimSpace(info.ComparisonMessage)
	if message == "" {
		message = "已完成上传并检查升级包版本。"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    info,
	})
}

// ConfirmManualServerUpgrade godoc
// @Summary Confirm upgrade with previously uploaded server binary
// @Tags Update
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/update/manual-upgrade [post]
func ConfirmManualServerUpgrade(c *gin.Context) {
	var request confirmManualUpgradeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "升级确认参数无效。",
		})
		return
	}

	info, err := service.ConfirmManualServerUpgrade(request.UploadToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务升级任务已启动，确认无误后将自动重启。",
		"data":    info,
	})
}
