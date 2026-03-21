package controller

import (
	"poolx/internal/model"
	"poolx/internal/service"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ManageRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

func Login(c *gin.Context) {
	var loginRequest LoginRequest
	if err := decodeJSONBody(c.Request.Body, &loginRequest); err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSecurity, model.AppLogLevelWarn, "login rejected | invalid request payload")
		respondFailure(c, "无效的参数")
		return
	}
	user, err := service.AuthenticateUser(service.LoginInput(loginRequest))
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSecurity, model.AppLogLevelWarn, "login failed | username="+loginRequest.Username+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}
	_ = service.AppLog.Push(model.AppLogClassificationSecurity, model.AppLogLevelInfo, "login succeeded | username="+user.Username)
	setupLogin(user, c)
}

func setupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	if err := session.Save(); err != nil {
		respondFailure(c, "无法保存会话信息，请重试")
		return
	}
	cleanUser := model.User{
		Id:          user.Id,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
	}
	respondSuccess(c, cleanUser)
}

func Logout(c *gin.Context) {
	username := c.GetString("username")
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		respondFailure(c, err.Error())
		return
	}
	_ = service.AppLog.Push(model.AppLogClassificationSecurity, model.AppLogLevelInfo, "logout succeeded | username="+username)
	respondSuccessMessage(c, "")
}

func Register(c *gin.Context) {
	var user model.User
	if err := decodeJSONBody(c.Request.Body, &user); err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSecurity, model.AppLogLevelWarn, "register rejected | invalid request payload")
		respondFailure(c, "无效的参数")
		return
	}
	if err := service.RegisterUser(user); err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationSecurity, model.AppLogLevelWarn, "register failed | username="+user.Username+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}
	_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "register succeeded | username="+user.Username)
	respondSuccessMessage(c, "")
}

func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("p"))
	users, err := service.ListUsers(page)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, users)
}

func SearchUsers(c *gin.Context) {
	users, err := service.SearchUsers(c.Query("keyword"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, users)
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	user, err := service.GetUserForRole(id, c.GetInt("role"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, user)
}

func GenerateToken(c *gin.Context) {
	token, err := service.GenerateUserToken(c.GetInt("id"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, token)
}

func GetSelf(c *gin.Context) {
	user, err := service.GetSelf(c.GetInt("id"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, user)
}

func UpdateUser(c *gin.Context) {
	var updatedUser model.User
	if err := decodeJSONBody(c.Request.Body, &updatedUser); err != nil || updatedUser.Id == 0 {
		respondFailure(c, "无效的参数")
		return
	}
	if err := service.UpdateUserAsAdmin(updatedUser, c.GetInt("role")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

func UpdateSelf(c *gin.Context) {
	var user model.User
	if err := decodeJSONBody(c.Request.Body, &user); err != nil {
		respondFailure(c, "无效的参数")
		return
	}
	if err := service.UpdateSelf(c.GetInt("id"), user); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	if err := service.DeleteUserAsAdmin(id, c.GetInt("role")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

func DeleteSelf(c *gin.Context) {
	if err := service.DeleteSelf(c.GetInt("id")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

func CreateUser(c *gin.Context) {
	var user model.User
	if err := decodeJSONBody(c.Request.Body, &user); err != nil {
		respondFailure(c, "无效的参数")
		return
	}
	if err := service.CreateUserAsAdmin(user, c.GetInt("role")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}

func ManageUser(c *gin.Context) {
	var req ManageRequest
	if err := decodeJSONBody(c.Request.Body, &req); err != nil {
		respondFailure(c, "无效的参数")
		return
	}
	user, err := service.ManageUser(service.ManageUserInput(req), c.GetInt("role"))
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, user)
}

func EmailBind(c *gin.Context) {
	if err := service.BindEmail(c.GetInt("id"), c.Query("email"), c.Query("code")); err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccessMessage(c, "")
}
