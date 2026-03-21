package controller

import (
	"poolx/internal/service"

	"github.com/gin-gonic/gin"
)

func WeChatAuth(c *gin.Context) {
	user, err := service.HandleWeChatOAuth(c.Query("code"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	setupLogin(user, c)
}

func WeChatBind(c *gin.Context) {
	if err := service.BindWeChatAccount(c.GetInt("id"), c.Query("code")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}
