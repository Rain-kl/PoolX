package example

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	exampleapp "github.com/Rain-kl/Foam/backend/internal/application/example"
	exampledomain "github.com/Rain-kl/Foam/backend/internal/domain/example"
	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *exampleapp.Service
}

func NewHandler(service *exampleapp.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts example CRUD under /api/v1/admin (admin auth required by parent group).
func (h *Handler) Register(router *gin.RouterGroup) {
	router.GET("/examples", h.list)
	router.GET("/examples/:id", h.get)
	router.POST("/examples", h.create)
	router.PUT("/examples/:id", h.update)
	router.DELETE("/examples/:id", h.delete)
}

type exampleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type exampleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *Handler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("pageSize", "20")))
	result, err := h.service.List(c.Request.Context(), page, pageSize, c.Query("search"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "exampleListFailed", "查询示例资源失败")
		return
	}
	items := make([]exampleResponse, 0, len(result.Items))
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
	var request exampleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	value, err := h.service.Create(c.Request.Context(), exampleapp.CreateInput{
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
	var request exampleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	value, err := h.service.Update(c.Request.Context(), c.Param("id"), exampleapp.UpdateInput{
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
	case errors.Is(err, exampleapp.ErrNotFound):
		response.Error(c, http.StatusNotFound, "exampleNotFound", "示例资源不存在")
	case errors.Is(err, exampleapp.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
	default:
		response.Error(c, http.StatusInternalServerError, "exampleFailed", "示例资源操作失败")
	}
}

func toResponse(value exampledomain.Example) exampleResponse {
	return exampleResponse{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
		CreatedAt:   value.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   value.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
