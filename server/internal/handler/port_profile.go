package controller

import (
	"poolx/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPortProfiles godoc
// @Summary List workspace port profiles
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles [get]
func GetPortProfiles(c *gin.Context) {
	items, err := service.ListPortProfiles()
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, items)
}

// GetPortProfile godoc
// @Summary Get one workspace port profile
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles/{id} [get]
func GetPortProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	item, err := service.GetPortProfile(id)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, item)
}

// CreatePortProfile godoc
// @Summary Create workspace port profile
// @Tags Workspace
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles [post]
func CreatePortProfile(c *gin.Context) {
	var payload service.PortProfilePayload
	if err := decodeJSONBody(c.Request.Body, &payload); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}
	item, err := service.CreatePortProfile(payload)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, item)
}

// UpdatePortProfile godoc
// @Summary Update workspace port profile
// @Tags Workspace
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles/{id} [post]
func UpdatePortProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	var payload service.PortProfilePayload
	if err := decodeJSONBody(c.Request.Body, &payload); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}
	item, err := service.UpdatePortProfile(id, payload)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, item)
}

// DeletePortProfile godoc
// @Summary Delete workspace port profile
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles/{id}/delete [post]
func DeletePortProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	if err := service.DeletePortProfile(id); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

// PreviewPortProfile godoc
// @Summary Generate mergeable workspace config fragment preview from payload
// @Tags Workspace
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles/preview [post]
func PreviewPortProfile(c *gin.Context) {
	var payload service.PortProfilePayload
	if err := decodeJSONBody(c.Request.Body, &payload); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}
	preview, err := service.PreviewPortProfile(payload)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, preview)
}

// PreviewSavedPortProfile godoc
// @Summary Generate mergeable fragment preview for saved workspace port profile
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles/{id}/preview [get]
func PreviewSavedPortProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	preview, err := service.PreviewSavedPortProfile(id)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, preview)
}

// SaveRuntimeConfig godoc
// @Summary Persist latest workspace fragment preview as snapshot
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profiles/{id}/runtime/save [post]
func SaveRuntimeConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	runtimeConfig, err := service.SaveRuntimePreview(id)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, runtimeConfig)
}

// GetProxyNodeOptions godoc
// @Summary List proxy node options for selector
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/proxy-nodes/options [get]
func GetProxyNodeOptions(c *gin.Context) {
	items, err := service.ListProxyNodeOptions(c.Query("keyword"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, items)
}
