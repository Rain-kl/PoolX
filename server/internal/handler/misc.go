package controller

import (
	"ginnexttemplate/internal/pkg/common"
	"ginnexttemplate/internal/service"

	"github.com/gin-gonic/gin"
)

// GetStatus godoc
// @Summary Get server status
// @Tags Public
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/status [get]
func GetStatus(c *gin.Context) {
	respondSuccess(c, gin.H{
		"version":            common.Version,
		"start_time":         common.StartTime,
		"email_verification": common.EmailVerificationEnabled,
		"github_oauth":       common.GitHubOAuthEnabled,
		"github_client_id":   common.GitHubClientId,
		"system_name":        common.SystemName,
		"home_page_link":     common.HomePageLink,
		"footer_html":        common.Footer,
		"wechat_qrcode":      common.WeChatAccountQRCodeImageURL,
		"wechat_login":       common.WeChatAuthEnabled,
		"server_address":     common.ServerAddress,
		"turnstile_check":    common.TurnstileCheckEnabled,
		"turnstile_site_key": common.TurnstileSiteKey,
	})
}

func GetNotice(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	respondSuccess(c, common.OptionMap["Notice"])
}

func GetAbout(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	respondSuccess(c, common.OptionMap["About"])
}

func SendEmailVerification(c *gin.Context) {
	if err := service.SendEmailVerification(c.Query("email")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

func SendPasswordResetEmail(c *gin.Context) {
	if err := service.SendPasswordResetEmail(c.Query("email")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

type PasswordResetRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func ResetPassword(c *gin.Context) {
	var req PasswordResetRequest
	if err := decodeJSONBody(c.Request.Body, &req); err != nil {
		respondFailure(c, "无效的参数")
		return
	}
	password, err := service.ResetPassword(service.PasswordResetInput(req))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, password)
}
