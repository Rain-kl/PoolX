package controller

import (
	"ginnexttemplate/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GitHubOAuth(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") != nil {
		GitHubBind(c)
		return
	}
	user, err := service.HandleGitHubOAuth(c.Query("code"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	setupLogin(user, c)
}

func GitHubBind(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		respondFailure(c, "未登录")
		return
	}
	if err := service.BindGitHubAccount(id.(int), c.Query("code")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "bind")
}
