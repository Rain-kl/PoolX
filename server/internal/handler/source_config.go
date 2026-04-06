package controller

import (
	"io"
	"net/url"
	"poolx/internal/model"
	"poolx/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type sourceImportRequest struct {
	SourceConfigID int      `json:"source_config_id"`
	Fingerprints   []string `json:"fingerprints"`
}

type sourceNodeTestRequest struct {
	SourceConfigID int      `json:"source_config_id"`
	Fingerprints   []string `json:"fingerprints"`
	TimeoutMS      int      `json:"timeout_ms"`
	TestURL        string   `json:"test_url"`
}

type sourceParseURLRequest struct {
	URL string `json:"url" binding:"required"`
}

func isValidSourceSubscriptionURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return strings.TrimSpace(parsed.Hostname()) != ""
}

// ParseSourceConfig godoc
// @Summary Upload YAML source and return parsed node preview
// @Tags SourceImport
// @Accept mpfd
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/source-configs/parse [post]
func ParseSourceConfig(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondFailure(c, "请先选择 YAML 文件")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		respondFailure(c, "读取上传文件失败")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		respondFailure(c, "读取上传内容失败")
		return
	}

	result, err := service.ParseAndStoreSourceConfig(service.SourceUploadInput{
		Filename:     fileHeader.Filename,
		UploadedBy:   c.GetString("username"),
		UploadedByID: c.GetInt("id"),
		Content:      content,
	})
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelWarn, "source config parse failed | username="+c.GetString("username")+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "source config parsed | username="+c.GetString("username")+" | source_config_id="+strconv.Itoa(result.SourceConfig.ID))
	respondSuccess(c, result)
}

// ParseSourceConfigURL godoc
// @Summary Fetch YAML source by URL and return parsed node preview
// @Tags SourceImport
// @Accept json
// @Produce json
// @Param payload body sourceParseURLRequest true "Subscription URL payload"
// @Failure 400 {object} map[string]interface{}
// @Success 200 {object} map[string]interface{}
// @Router /api/source-configs/parse-url [post]
func ParseSourceConfigURL(c *gin.Context) {
	var request sourceParseURLRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "请输入有效的 http/https 订阅地址")
		return
	}
	request.URL = strings.TrimSpace(request.URL)
	if request.URL == "" || !isValidSourceSubscriptionURL(request.URL) {
		respondBadRequest(c, "请输入有效的 http/https 订阅地址")
		return
	}

	result, err := service.ParseAndStoreSourceConfigFromURL(c.Request.Context(), service.SourceSubscriptionInput{
		URL:          request.URL,
		UploadedBy:   c.GetString("username"),
		UploadedByID: c.GetInt("id"),
	})
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelWarn, "source config parse url failed | username="+c.GetString("username")+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "source config parsed from url | username="+c.GetString("username")+" | source_config_id="+strconv.Itoa(result.SourceConfig.ID))
	respondSuccess(c, result)
}

// ImportSourceConfig godoc
// @Summary Import parsed nodes into proxy node pool
// @Tags SourceImport
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/source-configs/import [post]
func ImportSourceConfig(c *gin.Context) {
	var request sourceImportRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil || request.SourceConfigID <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}

	result, err := service.ImportSourceConfig(request.SourceConfigID, request.Fingerprints)
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelError, "source config import failed | username="+c.GetString("username")+" | source_config_id="+strconv.Itoa(request.SourceConfigID)+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "source config imported | username="+c.GetString("username")+" | source_config_id="+strconv.Itoa(request.SourceConfigID)+" | imported_nodes="+strconv.Itoa(result.ImportedNodes))
	respondSuccess(c, result)
}

// TestSourceConfigNodes godoc
// @Summary Test selected parsed nodes before importing into node pool
// @Tags SourceImport
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/source-configs/test [post]
func TestSourceConfigNodes(c *gin.Context) {
	var request sourceNodeTestRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil || request.SourceConfigID <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}

	results, err := service.TestSourceConfigNodes(c.Request.Context(), request.SourceConfigID, service.NodeTestInput{
		NodeFingerprints: request.Fingerprints,
		TimeoutMS:        request.TimeoutMS,
		TestURL:          request.TestURL,
	})
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelWarn, "source config node test failed | username="+c.GetString("username")+" | source_config_id="+strconv.Itoa(request.SourceConfigID)+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	respondSuccess(c, results)
}
