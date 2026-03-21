package controller

import (
	"poolx/internal/model"
	"poolx/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type proxyNodeStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type proxyNodeTestResultsQuery struct {
	ProxyNodeID int `form:"proxy_node_id"`
	Limit       int `form:"limit"`
}

// GetProxyNodes godoc
// @Summary List proxy nodes with paging and filters
// @Tags ProxyNode
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/proxy-nodes [get]
func GetProxyNodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("p"))
	sourceConfigID, _ := strconv.Atoi(c.Query("source_config_id"))

	var enabled *bool
	if value := strings.TrimSpace(c.Query("enabled")); value != "" {
		parsed := value == "true" || value == "1"
		enabled = &parsed
	}

	nodes, err := service.ListProxyNodes(service.ProxyNodeListInput{
		Page:           page,
		Keyword:        c.Query("keyword"),
		SourceConfigID: sourceConfigID,
		Enabled:        enabled,
	})
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, nodes)
}

// UpdateProxyNodeStatus godoc
// @Summary Enable or disable a proxy node
// @Tags ProxyNode
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/proxy-nodes/{id}/status [post]
func UpdateProxyNodeStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "无效的参数")
		return
	}

	var request proxyNodeStatusRequest
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}

	if err := service.SetProxyNodeEnabled(id, request.Enabled); err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelWarn, "proxy node status update failed | username="+c.GetString("username")+" | node_id="+strconv.Itoa(id)+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	respondSuccessMessage(c, "")
}

// TestProxyNodes godoc
// @Summary Test selected proxy nodes and persist the result
// @Tags ProxyNode
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/proxy-nodes/test [post]
func TestProxyNodes(c *gin.Context) {
	var request service.NodeTestInput
	if err := decodeJSONBody(c.Request.Body, &request); err != nil {
		respondBadRequest(c, "无效的参数")
		return
	}

	results, err := service.ExecuteNodeTests(request)
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelWarn, "proxy node test failed | username="+c.GetString("username")+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}

	respondSuccess(c, results)
}

// GetNodeTestResults godoc
// @Summary List recent test results for a proxy node
// @Tags ProxyNode
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/node-test-results [get]
func GetNodeTestResults(c *gin.Context) {
	proxyNodeID, _ := strconv.Atoi(c.Query("proxy_node_id"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	results, err := service.GetNodeTestResults(proxyNodeID, limit)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, results)
}
