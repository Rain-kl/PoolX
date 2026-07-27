package settings

import (
	"errors"
	"net/http"
	"time"

	settingsapp "github.com/Rain-kl/Foam/backend/internal/application/settings"
	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *settingsapp.Service
}

func NewHandler(service *settingsapp.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts settings under /api/v1/admin (admin auth required by parent group).
func (h *Handler) Register(router *gin.RouterGroup) {
	router.GET("/settings", h.get)
	router.PUT("/settings", h.update)
}

type settingsConfigDTO struct {
	App      appConfigDTO      `json:"app"`
	Frontend frontendConfigDTO `json:"frontend"`
	Clash    clashConfigDTO    `json:"clash"`
}

type appConfigDTO struct {
	DisplayName string `json:"display_name"`
}

type frontendConfigDTO struct {
	PublicAPIBaseURL string `json:"public_api_base_url"`
}

type clashConfigDTO struct {
	KernelType               string `json:"kernel_type"`
	MihomoBinaryPath         string `json:"mihomo_binary_path"`
	MihomoBinaryVersion      string `json:"mihomo_binary_version"`
	MihomoBinarySource       string `json:"mihomo_binary_source"`
	ClashExternalController  string `json:"clash_external_controller"`
	ClashMode                string `json:"clash_mode"`
	ClashSecret              string `json:"clash_secret"`
	ClashAllowLAN            bool   `json:"clash_allow_lan"`
	NodeTestDefaultURL       string `json:"node_test_default_url"`
	NodeTestDefaultTimeoutMS int    `json:"node_test_default_timeout_ms"`
}

type settingsSnapshotDTO struct {
	Config               settingsConfigDTO `json:"config"`
	Revision             uint64            `json:"revision"`
	UpdatedAt            string            `json:"updated_at,omitempty"`
	FilePublicAPIBaseURL string            `json:"file_public_api_base_url"`
	Effective            effectiveDTO      `json:"effective"`
}

type effectiveDTO struct {
	DisplayName      string `json:"display_name"`
	PublicAPIBaseURL string `json:"public_api_base_url"`
}

type updateRequest struct {
	Revision uint64            `json:"revision"`
	Config   settingsConfigDTO `json:"config" binding:"required"`
}

func (h *Handler) get(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "settingsUnavailable", "运行设置服务暂不可用")
		return
	}
	response.Success(c, http.StatusOK, toDTO(h.service))
}

func (h *Handler) update(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "settingsUnavailable", "运行设置服务暂不可用")
		return
	}
	var request updateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	snapshot, err := h.service.Update(c.Request.Context(), request.Revision, settingsapp.EditableConfig{
		App: settingsapp.AppConfigInput{
			DisplayName: request.Config.App.DisplayName,
		},
		Frontend: settingsapp.FrontendConfigInput{
			PublicAPIBaseURL: request.Config.Frontend.PublicAPIBaseURL,
		},
		Clash: settingsapp.ClashConfigInput{
			KernelType:               request.Config.Clash.KernelType,
			MihomoBinaryPath:         request.Config.Clash.MihomoBinaryPath,
			MihomoBinaryVersion:      request.Config.Clash.MihomoBinaryVersion,
			MihomoBinarySource:       request.Config.Clash.MihomoBinarySource,
			ClashExternalController:  request.Config.Clash.ClashExternalController,
			ClashMode:                request.Config.Clash.ClashMode,
			ClashSecret:              request.Config.Clash.ClashSecret,
			ClashAllowLAN:            request.Config.Clash.ClashAllowLAN,
			NodeTestDefaultURL:       request.Config.Clash.NodeTestDefaultURL,
			NodeTestDefaultTimeoutMS: request.Config.Clash.NodeTestDefaultTimeoutMS,
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, settingsapp.ErrInvalidInput):
			response.Error(c, http.StatusBadRequest, "invalidRequest", "运行设置参数无效")
		case errors.Is(err, settingsapp.ErrConflict):
			response.Error(c, http.StatusConflict, "settingsConflict", "运行设置已被其他会话更新，请刷新后重试")
		default:
			response.Error(c, http.StatusInternalServerError, "settingsUpdateFailed", "更新运行设置失败")
		}
		return
	}
	response.Success(c, http.StatusOK, snapshotToDTO(snapshot, h.service))
}

func toDTO(service *settingsapp.Service) settingsSnapshotDTO {
	return snapshotToDTO(service.Get(), service)
}

func snapshotToDTO(snapshot settingsapp.Snapshot, service *settingsapp.Service) settingsSnapshotDTO {
	updatedAt := ""
	if !snapshot.UpdatedAt.IsZero() {
		updatedAt = snapshot.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return settingsSnapshotDTO{
		Config: settingsConfigDTO{
			App: appConfigDTO{
				DisplayName: snapshot.Config.App.DisplayName,
			},
			Frontend: frontendConfigDTO{
				PublicAPIBaseURL: snapshot.Config.Frontend.PublicAPIBaseURL,
			},
			Clash: clashConfigDTO{
				KernelType:               snapshot.Config.Clash.KernelType,
				MihomoBinaryPath:         snapshot.Config.Clash.MihomoBinaryPath,
				MihomoBinaryVersion:      snapshot.Config.Clash.MihomoBinaryVersion,
				MihomoBinarySource:       snapshot.Config.Clash.MihomoBinarySource,
				ClashExternalController:  snapshot.Config.Clash.ClashExternalController,
				ClashMode:                snapshot.Config.Clash.ClashMode,
				ClashSecret:              snapshot.Config.Clash.ClashSecret,
				ClashAllowLAN:            snapshot.Config.Clash.ClashAllowLAN,
				NodeTestDefaultURL:       snapshot.Config.Clash.NodeTestDefaultURL,
				NodeTestDefaultTimeoutMS: snapshot.Config.Clash.NodeTestDefaultTimeoutMS,
			},
		},
		Revision:             snapshot.Revision,
		UpdatedAt:            updatedAt,
		FilePublicAPIBaseURL: snapshot.FilePublicAPIBaseURL,
		Effective: effectiveDTO{
			DisplayName:      service.DisplayName(),
			PublicAPIBaseURL: service.PublicAPIBaseURL(),
		},
	}
}
