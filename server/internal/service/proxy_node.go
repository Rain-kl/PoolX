package service

import (
	"fmt"
	"net"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ProxyNodeListInput struct {
	Page           int
	Keyword        string
	SourceConfigID int
	Enabled        *bool
}

type NodeTestInput struct {
	NodeIDs   []int  `json:"node_ids"`
	TimeoutMS int    `json:"timeout_ms"`
	TestURL   string `json:"test_url"`
}

type NodeTestExecution struct {
	NodeID       int        `json:"node_id"`
	NodeName     string     `json:"node_name"`
	Status       string     `json:"status"`
	LatencyMS    *int       `json:"latency_ms,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	TestURL      string     `json:"test_url,omitempty"`
	DialAddress  string     `json:"dial_address"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   time.Time  `json:"finished_at"`
	LastTestedAt *time.Time `json:"last_tested_at,omitempty"`
}

func ListProxyNodes(input ProxyNodeListInput) ([]*model.ProxyNode, error) {
	page := input.Page
	if page < 0 {
		page = 0
	}
	return model.ListProxyNodes(page*common.ItemsPerPage, common.ItemsPerPage, model.ProxyNodeListFilter{
		Keyword:        strings.TrimSpace(input.Keyword),
		SourceConfigID: input.SourceConfigID,
		Enabled:        input.Enabled,
	})
}

func SetProxyNodeEnabled(id int, enabled bool) error {
	node, err := model.GetProxyNodeByID(id)
	if err != nil {
		return fmt.Errorf("节点不存在")
	}
	node.Enabled = enabled
	return model.DB.Save(node).Error
}

func ExecuteNodeTests(input NodeTestInput) ([]NodeTestExecution, error) {
	if len(input.NodeIDs) == 0 {
		return nil, fmt.Errorf("请先选择要测试的节点")
	}
	timeout := time.Duration(input.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	if timeout > 30*time.Second {
		timeout = 30 * time.Second
	}

	nodes, err := model.FindProxyNodesByIDs(input.NodeIDs)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到可测试的节点")
	}

	results := make([]NodeTestExecution, 0, len(nodes))
	for _, node := range nodes {
		execution := testSingleNode(node, timeout, strings.TrimSpace(input.TestURL))
		if err := persistNodeTestExecution(node.ID, execution); err != nil {
			return nil, err
		}
		results = append(results, execution)
	}

	return results, nil
}

func GetNodeTestResults(proxyNodeID int, limit int) ([]*model.NodeTestResult, error) {
	if proxyNodeID <= 0 {
		return nil, fmt.Errorf("无效的节点 ID")
	}
	return model.ListNodeTestResults(proxyNodeID, limit)
}

func testSingleNode(node *model.ProxyNode, timeout time.Duration, testURL string) NodeTestExecution {
	startedAt := time.Now()
	dialAddress := net.JoinHostPort(node.Server, strconv.Itoa(node.Port))
	conn, err := net.DialTimeout("tcp", dialAddress, timeout)
	finishedAt := time.Now()

	execution := NodeTestExecution{
		NodeID:      node.ID,
		NodeName:    node.Name,
		TestURL:     testURL,
		DialAddress: dialAddress,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
	}

	if err != nil {
		execution.Status = model.NodeTestStatusFailed
		execution.ErrorMessage = err.Error()
		execution.LastTestedAt = &finishedAt
		return execution
	}
	_ = conn.Close()

	latency := int(finishedAt.Sub(startedAt).Milliseconds())
	execution.Status = model.NodeTestStatusSuccess
	execution.LatencyMS = &latency
	execution.LastTestedAt = &finishedAt
	return execution
}

func persistNodeTestExecution(nodeID int, execution NodeTestExecution) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		result := &model.NodeTestResult{
			ProxyNodeID:  nodeID,
			Status:       execution.Status,
			LatencyMS:    execution.LatencyMS,
			ErrorMessage: execution.ErrorMessage,
			TestURL:      execution.TestURL,
			DialAddress:  execution.DialAddress,
			StartedAt:    execution.StartedAt,
			FinishedAt:   execution.FinishedAt,
		}
		if err := tx.Create(result).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"last_test_status": execution.Status,
			"last_latency_ms":  execution.LatencyMS,
			"last_test_error":  execution.ErrorMessage,
			"last_tested_at":   execution.LastTestedAt,
		}
		return tx.Model(&model.ProxyNode{}).Where("id = ?", nodeID).Updates(updates).Error
	})
}
