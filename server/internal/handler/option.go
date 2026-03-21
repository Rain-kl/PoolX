package controller

import (
	"ginnexttemplate/internal/model"
	"ginnexttemplate/internal/service"

	"github.com/gin-gonic/gin"
)

type geoIPPreviewRequest struct {
	Provider string `json:"provider"`
	IP       string `json:"ip"`
}

// GetOptions godoc
// @Summary List editable options
// @Tags Options
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/option/ [get]
func GetOptions(c *gin.Context) {
	respondSuccess(c, service.ListEditableOptions())
}

// UpdateOption godoc
// @Summary Update option
// @Tags Options
// @Accept json
// @Produce json
// @Param payload body model.Option true "Option payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/option/update [post]
func UpdateOption(c *gin.Context) {
	var option model.Option
	if err := decodeJSONBody(c.Request.Body, &option); err != nil {
		respondBadRequest(c, "invalid request payload")
		return
	}
	if err := service.UpdateEditableOption(option); err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelError, "option update failed | key="+option.Key+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}
	_ = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelInfo, "option updated | key="+option.Key)
	respondSuccessMessage(c, "")
}

func PreviewGeoIP(c *gin.Context) {
	var request geoIPPreviewRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "invalid request payload")
		return
	}

	preview, err := service.PreviewGeoIPLookup(request.Provider, request.IP)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}

	respondSuccess(c, preview)
}
