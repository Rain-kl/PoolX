package service

import (
	"context"
	"fmt"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	kernelpkg "poolx/internal/pkg/kernel"
	"sort"
	"strings"
	"sync"
	"time"
)

const maxNodeTestParallelism = 4

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

type ProxyNodeDeleteInput struct {
	NodeIDs []int `json:"node_ids"`
}

type ProxyNodeTagsInput struct {
	NodeIDs []int  `json:"node_ids"`
	Tags    string `json:"tags"`
}

type NodeTestExecution struct {
	NodeID       int        `json:"node_id"`
	NodeName     string     `json:"node_name"`
	Status       string     `json:"status"`
	LatencyMS    *int       `json:"latency_ms,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	TestURL      string     `json:"test_url,omitempty"`
	DialAddress  string     `json:"dial_address"`
	Cached       bool       `json:"cached"`
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

func DeleteProxyNode(id int) error {
	if id <= 0 {
		return fmt.Errorf("无效的节点 ID")
	}
	if _, err := model.GetProxyNodeByID(id); err != nil {
		return fmt.Errorf("节点不存在")
	}
	return model.DeleteProxyNodeByID(id)
}

func DeleteProxyNodes(ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("请先选择要删除的节点")
	}

	deleted := 0
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		if err := DeleteProxyNode(id); err != nil {
			return deleted, err
		}
		deleted++
	}
	if deleted == 0 {
		return 0, fmt.Errorf("请先选择要删除的节点")
	}
	return deleted, nil
}

func UpdateProxyNodeTags(ids []int, tags string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("请先选择至少一个节点")
	}
	normalized := normalizeProxyNodeTags(tags)
	result := model.DB.Model(&model.ProxyNode{}).Where("id IN ?", ids).Update("tags", normalized)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, fmt.Errorf("未找到可更新的节点")
	}
	return int(result.RowsAffected), nil
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

	type indexedExecution struct {
		index int
		item  NodeTestExecution
	}
	jobs := make(chan int)
	resultsCh := make(chan indexedExecution, len(nodes))
	workerCount := minNodeTestWorkerCount(len(nodes), maxNodeTestParallelism)
	if workerCount <= 0 {
		workerCount = 1
	}
	var wg sync.WaitGroup
	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				node := nodes[index]
				execution := buildNodeTestExecution(ctx, node, timeout, testURL)
				resultsCh <- indexedExecution{index: index, item: execution}
			}
		}()
	}
	for index := range nodes {
		jobs <- index
	}
	close(jobs)
	wg.Wait()
	close(resultsCh)

	results := make([]NodeTestExecution, 0, len(nodes))
	ordered := make([]indexedExecution, 0, len(nodes))
	for item := range resultsCh {
		ordered = append(ordered, item)
	}
	sort.Slice(ordered, func(left int, right int) bool {
		return ordered[left].index < ordered[right].index
	})
	for _, item := range ordered {
		if err := persistNodeTestExecution(item.item.NodeID, item.item); err != nil {
			return nil, err
		}
		results = append(results, item.item)
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

func buildNodeTestExecution(ctx context.Context, node *model.ProxyNode, timeout time.Duration, testURL string) NodeTestExecution {
	return executeMetadataNodeTest(ctx, metadataNodeTestInput{
		NodeID:       node.ID,
		Name:         node.Name,
		Server:       node.Server,
		Port:         node.Port,
		MetadataJSON: node.MetadataJSON,
		Timeout:      timeout,
		TestURL:      testURL,
	})
}

func normalizeNodeTestTimeout(timeoutMS int) time.Duration {
	if timeoutMS <= 0 {
		timeoutMS = common.NodeTestDefaultTimeoutMS
	}
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
	value := strings.TrimSpace(testURL)
	if value == "" {
		value = strings.TrimSpace(common.NodeTestDefaultURL)
	}
	if value == "" {
		return common.DefaultNodeTestURL
	}
	return value
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

func normalizeProxyNodeTags(tags string) string {
	if strings.TrimSpace(tags) == "" {
		return ""
	}
	parts := strings.FieldsFunc(tags, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n'
	})
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return strings.Join(result, ", ")
}

func minNodeTestWorkerCount(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
