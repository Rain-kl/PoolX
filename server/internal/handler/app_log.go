package controller

import (
	"ginnexttemplate/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAppLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	afterID, _ := strconv.Atoi(c.DefaultQuery("after_id", "0"))

	logs, err := service.ListAppLogs(limit, afterID, c.Query("classification"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, logs)
}

func PushAppLog(c *gin.Context) {
	var input service.AppLogPushInput
	if err := decodeJSONBody(c.Request.Body, &input); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}
	if err := service.PushFrontendAppLog(input); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}
