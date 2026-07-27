package adminauth

import (
	"errors"
	"io"
	"net/http"
	"time"

	adminapp "github.com/Rain-kl/Foam/backend/internal/application/adminauth"
	admindomain "github.com/Rain-kl/Foam/backend/internal/domain/admin"
	"github.com/Rain-kl/Foam/backend/internal/shared/response"
	"github.com/Rain-kl/Foam/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

const (
	refreshCookieName = "foam_admin_refresh"
	// Cookie scoped to user auth routes (Wavelet-style /api/v1/user/*).
	refreshCookiePath = "/api/v1/user"
)

type Handler struct {
	service       *adminapp.Service
	secureCookies bool
}

func NewHandler(service *adminapp.Service, secureCookies bool) *Handler {
	return &Handler{service: service, secureCookies: secureCookies}
}

// RegisterPublic mounts unauthenticated auth routes under /api/v1/user.
func (h *Handler) RegisterPublic(router *gin.RouterGroup) {
	router.POST("/login", h.login)
	router.POST("/refresh", h.refresh)
	router.GET("/logout", h.logout)
	router.POST("/logout", h.logout) // allow POST for clients that prefer it
}

// RegisterAuthenticated mounts authenticated user routes under /api/v1/user.
func (h *Handler) RegisterAuthenticated(router *gin.RouterGroup) {
	router.GET("/self", h.me)
	router.POST("/change-password", h.changePassword)
}

// RegisterUserInfoAliases mounts Wavelet-compatible me aliases on the v1 group.
func (h *Handler) RegisterUserInfoAliases(v1 *gin.RouterGroup) {
	v1.GET("/user-info", h.me)
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type tokenResponse struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  string `json:"access_token_expires_at"`
	RefreshTokenExpiresAt string `json:"refresh_token_expires_at"`
}

type userResponse struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

type loginResponse struct {
	userResponse
	Tokens tokenResponse `json:"tokens"`
}

func (h *Handler) login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	adminValue, tokens, err := h.service.Login(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		if errors.Is(err, adminapp.ErrRuntimeUnavailable) {
			response.Error(c, http.StatusServiceUnavailable, "authRuntimeUnavailable", "管理员认证服务暂不可用")
			return
		}
		response.Error(c, http.StatusUnauthorized, "invalidCredentials", "管理员账号或密码错误")
		return
	}
	h.setRefreshCookie(c, tokens)
	h.setAccessCookie(c, tokens)
	response.Success(c, http.StatusOK, loginResponse{
		userResponse: newUserResponse(adminValue),
		Tokens:       newTokenResponse(tokens),
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var request refreshRequest
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	if request.RefreshToken == "" {
		request.RefreshToken, _ = c.Cookie(refreshCookieName)
	}
	if request.RefreshToken == "" {
		response.Error(c, http.StatusUnauthorized, "invalidRefreshToken", "刷新会话无效")
		return
	}
	tokens, err := h.service.Refresh(c.Request.Context(), request.RefreshToken)
	if err != nil {
		if errors.Is(err, adminapp.ErrRuntimeUnavailable) {
			response.Error(c, http.StatusServiceUnavailable, "authRuntimeUnavailable", "管理员认证服务暂不可用")
			return
		}
		response.Error(c, http.StatusUnauthorized, "invalidRefreshToken", "刷新会话无效")
		return
	}
	h.setRefreshCookie(c, tokens)
	h.setAccessCookie(c, tokens)
	response.Success(c, http.StatusOK, newTokenResponse(tokens))
}

func (h *Handler) logout(c *gin.Context) {
	var request refreshRequest
	// GET logout: body optional; prefer cookie.
	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
			response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
			return
		}
	}
	if request.RefreshToken == "" {
		request.RefreshToken, _ = c.Cookie(refreshCookieName)
	}
	if err := h.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "authRuntimeUnavailable", "管理员认证服务暂不可用")
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, "", -1, refreshCookiePath, "", h.secureCookies || c.Request.TLS != nil, true)
	c.SetCookie(middleware.AccessCookieName, "", -1, middleware.AccessCookiePath, "", h.secureCookies || c.Request.TLS != nil, true)
	// Wavelet returns empty string / null; keep a small explicit payload for scaffold clients.
	response.Success(c, http.StatusOK, "")
}

func (h *Handler) me(c *gin.Context) {
	value, ok := c.Get(middleware.AdminKey)
	adminValue, valid := value.(admindomain.Admin)
	if !ok || !valid {
		response.Error(c, http.StatusUnauthorized, "adminUnauthorized", "未登录")
		return
	}
	response.Success(c, http.StatusOK, newUserResponse(adminValue))
}

func (h *Handler) changePassword(c *gin.Context) {
	var request changePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalidRequest", "请求参数无效")
		return
	}
	value, ok := c.Get(middleware.AdminKey)
	adminValue, valid := value.(admindomain.Admin)
	if !ok || !valid {
		response.Error(c, http.StatusUnauthorized, "adminUnauthorized", "未登录")
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), adminValue.ID, request.OldPassword, request.NewPassword); err != nil {
		if errors.Is(err, adminapp.ErrInvalidCredentials) || errors.Is(err, adminapp.ErrInvalidPassword) {
			response.Error(c, http.StatusBadRequest, "passwordChangeFailed", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "passwordChangeFailed", "修改管理员密码失败")
		return
	}
	response.Success(c, http.StatusOK, "密码修改成功")
}

func newUserResponse(value admindomain.Admin) userResponse {
	return userResponse{ID: value.ID, Username: value.Username, IsAdmin: true}
}

func newTokenResponse(value adminapp.Tokens) tokenResponse {
	return tokenResponse{
		AccessToken:           value.AccessToken,
		AccessTokenExpiresAt:  value.AccessTokenExpiresAt.Format(time.RFC3339),
		RefreshTokenExpiresAt: value.RefreshTokenExpiresAt.Format(time.RFC3339),
	}
}

func (h *Handler) setRefreshCookie(c *gin.Context, value adminapp.Tokens) {
	maxAge := int(time.Until(value.RefreshTokenExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, value.RefreshToken, maxAge, refreshCookiePath, "", h.secureCookies || c.Request.TLS != nil, true)
}

func (h *Handler) setAccessCookie(c *gin.Context, value adminapp.Tokens) {
	maxAge := int(time.Until(value.AccessTokenExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(middleware.AccessCookieName, value.AccessToken, maxAge, middleware.AccessCookiePath, "", h.secureCookies || c.Request.TLS != nil, true)
}
