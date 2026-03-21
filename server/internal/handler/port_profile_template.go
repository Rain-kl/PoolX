package controller

import (
	"poolx/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type savePortProfileTemplateRequest struct {
	Name    string                     `json:"name"`
	Payload service.PortProfilePayload `json:"payload"`
}

// GetPortProfileTemplates godoc
// @Summary List saved workspace templates
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profile-templates [get]
func GetPortProfileTemplates(c *gin.Context) {
	items, err := service.ListPortProfileTemplates()
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, items)
}

// SavePortProfileTemplate godoc
// @Summary Save current workspace payload as template
// @Tags Workspace
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profile-templates [post]
func SavePortProfileTemplate(c *gin.Context) {
	var request savePortProfileTemplateRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}
	item, err := service.SavePortProfileTemplate(request.Name, request.Payload)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, item)
}

// DeletePortProfileTemplate godoc
// @Summary Delete a workspace template
// @Tags Workspace
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/port-profile-templates/{id}/delete [post]
func DeletePortProfileTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	if err := service.DeletePortProfileTemplate(id); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}
