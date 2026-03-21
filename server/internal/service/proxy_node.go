package service

import (
	"context"
	"fmt"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	kernelpkg "poolx/internal/pkg/kernel"
	"strings"
	"time"
)

const defaultNodeTestURL = "https://cp.cloudflare.com/generate_204"

var runNodeKernelTest = kernelpkg.TestNodeWithMihomo

type ProxyNodeListInput struct {
	Page           int
	Keyword        string
	SourceConfigID int
	Enabled        *bool
}

type NodeTestInput struct {
	NodeIDs          []int    `json:"node_ids"`
	NodeFingerprints []string `json:"node_fingerprints"`
	TimeoutMS        int      `json:"timeout_ms"`
	TestURL          string   `json:"test_url"`
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

func ExecuteNodeTests(ctx context.Context, input NodeTestInput) ([]NodeTestExecution, error) {
	if len(input.NodeIDs) == 0 {
		return nil, fmt.Errorf("请先选择要测试的节点")
	}
	if strings.TrimSpace(common.MihomoBinaryPath) == "" {
		return nil, fmt.Errorf("请先在系统设置中完成 Mihomo 二进制安装或路径校验")
	}
	timeout := normalizeNodeTestTimeout(input.TimeoutMS)
	testURL := normalizeNodeTestURL(input.TestURL)

	nodes, err := model.FindProxyNodesByIDs(input.NodeIDs)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到可测试的节点")
	}

	results := make([]NodeTestExecution, 0, len(nodes))
	for _, node := range nodes {
		execution := executeMetadataNodeTest(ctx, metadataNodeTestInput{
			NodeID:       node.ID,
			Name:         node.Name,
			Server:       node.Server,
			Port:         node.Port,
			MetadataJSON: node.MetadataJSON,
			Timeout:      timeout,
			TestURL:      testURL,
		})
		if err := persistNodeTestExecution(node.ID, execution); err != nil {
			return nil, err
		}
		results = append(results, execution)
	}

	return results, nil
}

type metadataNodeTestInput struct {
	NodeID       int
	Name         string
	Server       string
	Port         int
	MetadataJSON string
	Timeout      time.Duration
	TestURL      string
}

func executeMetadataNodeTest(ctx context.Context, input metadataNodeTestInput) NodeTestExecution {
	startedAt := time.Now()
	dialAddress := fmt.Sprintf("%s:%d", input.Server, input.Port)

	execution := NodeTestExecution{
		NodeID:      input.NodeID,
		NodeName:    input.Name,
		TestURL:     input.TestURL,
		DialAddress: dialAddress,
		StartedAt:   startedAt,
	}

	result, err := runNodeKernelTest(ctx, kernelpkg.MihomoNodeTestInput{
		BinaryPath:   common.MihomoBinaryPath,
		ProxyName:    input.Name,
		MetadataJSON: input.MetadataJSON,
		TestURL:      input.TestURL,
		Timeout:      input.Timeout,
	})
	finishedAt := time.Now()
	execution.FinishedAt = finishedAt

	if err != nil {
		execution.Status = model.NodeTestStatusFailed
		execution.ErrorMessage = err.Error()
		execution.LastTestedAt = &finishedAt
		return execution
	}

	latency := result.LatencyMS
	execution.Status = model.NodeTestStatusSuccess
	execution.LatencyMS = &latency
	execution.LastTestedAt = &finishedAt
	return execution
}

func normalizeNodeTestTimeout(timeoutMS int) time.Duration {
	timeout := time.Duration(timeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	if timeout > 60*time.Second {
		timeout = 60 * time.Second
	}
	return timeout
}

func normalizeNodeTestURL(testURL string) string {
	if strings.TrimSpace(testURL) == "" {
		return defaultNodeTestURL
	}
	return strings.TrimSpace(testURL)
}

func persistNodeTestExecution(nodeID int, execution NodeTestExecution) error {
	updates := map[string]any{
		"last_test_status": execution.Status,
		"last_latency_ms":  execution.LatencyMS,
		"last_test_error":  execution.ErrorMessage,
		"last_tested_at":   execution.LastTestedAt,
	}
	return model.DB.Model(&model.ProxyNode{}).Where("id = ?", nodeID).Updates(updates).Error
}
