package controller

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"poolx/internal/pkg/common"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProxyZashboardClash(c *gin.Context) {
	controllerAddress := strings.TrimSpace(common.ClashExternalController)
	if controllerAddress == "" {
		controllerAddress = common.DefaultClashExternalController
	}

	target, err := url.Parse("http://" + controllerAddress)
	if err != nil {
		respondFailure(c, "Clash 控制地址无效")
		return
	}

	secret := strings.TrimSpace(common.ClashSecret)
	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = path
		req.URL.RawPath = path
		req.Host = target.Host
		if secret != "" {
			req.Header.Set("Authorization", "Bearer "+secret)
			query := req.URL.Query()
			query.Set("token", secret)
			req.URL.RawQuery = query.Encode()
		}
	}
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, proxyErr error) {
		c.JSON(502, gin.H{
			"success": false,
			"message": "连接 Clash 控制接口失败: " + proxyErr.Error(),
		})
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
