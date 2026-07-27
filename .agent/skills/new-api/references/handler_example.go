// Illustrative HTTP handler for the new-api skill.
//
// Real location: backend/internal/transport/http/widget/handler.go (package widget)
// Mount under admin group → final path /api/v1/admin/widgets
//
// Envelope (Wavelet-aligned):
//
//	success: { "error_msg": "", "data": ... }
//	error:   { "error_msg": "...", "data": null }
//
// Prefer response.Success / response.Error from
// github.com/Rain-kl/Foam/backend/internal/shared/response
package references

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler binds HTTP to the application service.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register mounts routes on a group that already has AdminAuth
// (typically the /api/v1/admin group from server.go).
func (h *Handler) Register(router *gin.RouterGroup) {
	router.GET("/widgets", h.list)
	router.GET("/widgets/:id", h.get)
	router.POST("/widgets", h.create)
	router.PUT("/widgets/:id", h.update)
	router.DELETE("/widgets/:id", h.delete)
}

type widgetRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type widgetResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *Handler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.service.List(c.Request.Context(), page, pageSize, c.Query("search"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "widgetListFailed", "查询资源失败")
		return
	}
	items := make([]widgetResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toResponse(item))
	}
	response.Success(c, http.StatusOK, gin.H{
		"items":     items,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	})
}

func (h *Handler) get(c *gin.Context) {
	value, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, toResponse(value))
}

func (h *Handler) create(c *gin.Context) {
	var request widgetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	value, err := h.service.Create(c.Request.Context(), CreateInput{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, toResponse(value))
}

func (h *Handler) update(c *gin.Context) {
	var request widgetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	value, err := h.service.Update(c.Request.Context(), c.Param("id"), UpdateInput{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, toResponse(value))
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, "widgetNotFound", "资源不存在")
	case errors.Is(err, ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
	default:
		response.Error(c, http.StatusInternalServerError, "widgetFailed", "资源操作失败")
	}
}

func toResponse(value Widget) widgetResponse {
	return widgetResponse{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
		CreatedAt:   value.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   value.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// Wiring reminder (not in this package):
//
//  app/application.go:
//    widgetRepo := relational.NewWidgetRepository(database)
//    widgetService := widgetapp.NewService(widgetRepo)
//    httpserver.Dependencies{ Widgets: widgetService, ... }
//
//  transport/http/server.go:
//    admin.Use(middleware.AdminAuth(...))
//    widgethttp.NewHandler(deps.Widgets).Register(admin)
