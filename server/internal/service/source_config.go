package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"poolx/internal/model"
	proxypkg "poolx/internal/pkg/proxy"
	"poolx/internal/pkg/sourcefetch"
	"strings"
	"time"

	"gorm.io/gorm"
)

const sourcePreviewLimit = 100

type SourceUploadInput struct {
	Filename     string
	UploadedBy   string
	UploadedByID int
	Content      []byte
}

type SourceSubscriptionInput struct {
	URL          string
	UploadedBy   string
	UploadedByID int
}

type sourceParseInput struct {
	SourceType   string
	Filename     string
	SourceURL    string
	ContentType  string
	FetchedAt    *time.Time
	UploadedBy   string
	UploadedByID int
	Content      []byte
}

var sourceConfigURLFetcher = sourcefetch.NewFetcher().FetchYAML

type ParsedNodePreview struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Server         string `json:"server"`
	Port           int    `json:"port"`
	Fingerprint    string `json:"fingerprint"`
	DuplicateScope string `json:"duplicate_scope"`
}

type ParseSummary struct {
	TotalNodes      int `json:"total_nodes"`
	ValidNodes      int `json:"valid_nodes"`
	InvalidNodes    int `json:"invalid_nodes"`
	DuplicateNodes  int `json:"duplicate_nodes"`
	ImportableNodes int `json:"importable_nodes"`
}

type SourceParseResponse struct {
	SourceConfig model.SourceConfig    `json:"source_config"`
	Summary      ParseSummary          `json:"summary"`
	Nodes        []ParsedNodePreview   `json:"nodes"`
	Errors       []proxypkg.ParseIssue `json:"errors"`
}

type SourceImportResult struct {
	SourceConfigID int `json:"source_config_id"`
	ImportedNodes  int `json:"imported_nodes"`
	SkippedNodes   int `json:"skipped_nodes"`
}

func ParseAndStoreSourceConfig(input SourceUploadInput) (*SourceParseResponse, error) {
	return parseAndStoreSourceConfig(sourceParseInput{
		SourceType:   model.SourceConfigSourceTypeUpload,
		Filename:     input.Filename,
		UploadedBy:   input.UploadedBy,
		UploadedByID: input.UploadedByID,
		Content:      input.Content,
	})
}

func ParseAndStoreSourceConfigFromURL(ctx context.Context, input SourceSubscriptionInput) (*SourceParseResponse, error) {
	fetched, err := sourceConfigURLFetcher(ctx, input.URL)
	if err != nil {
		return nil, err
	}
	if fetched == nil {
		return nil, fmt.Errorf("拉取订阅内容失败")
	}

	return parseAndStoreSourceConfig(sourceParseInput{
		SourceType:   model.SourceConfigSourceTypeSubscriptionURL,
		Filename:     fetched.DisplayName,
		SourceURL:    strings.TrimSpace(input.URL),
		ContentType:  fetched.ContentType,
		FetchedAt:    &fetched.FetchedAt,
		UploadedBy:   input.UploadedBy,
		UploadedByID: input.UploadedByID,
		Content:      fetched.Content,
	})
}

func SourceConfigURLFetcherForTest() func(context.Context, string) (*sourcefetch.FetchResult, error) {
	return sourceConfigURLFetcher
}

func SetSourceConfigURLFetcherForTest(fetcher func(context.Context, string) (*sourcefetch.FetchResult, error)) {
	if fetcher == nil {
		sourceConfigURLFetcher = sourcefetch.NewFetcher().FetchYAML
		return
	}
	sourceConfigURLFetcher = fetcher
}

func parseAndStoreSourceConfig(input sourceParseInput) (*SourceParseResponse, error) {
	filename := strings.TrimSpace(input.Filename)
	if filename == "" {
		return nil, fmt.Errorf("请先选择 YAML 文件")
	}
	if len(input.Content) == 0 {
		return nil, fmt.Errorf("上传内容为空")
	}

	parseResult, err := proxypkg.ParseYAML(input.Content)
	if err != nil {
		return nil, err
	}

	existingFingerprints, err := model.FindExistingNodeFingerprints(parsedFingerprints(parseResult.Nodes))
	if err != nil {
		return nil, err
	}

	previewNodes, duplicateCount, importableCount := buildPreviewNodes(parseResult.Nodes, existingFingerprints)
	contentHash := checksum(input.Content)

	sourceConfig := &model.SourceConfig{
		SourceType:     input.SourceType,
		SourceURL:      strings.TrimSpace(input.SourceURL),
		Filename:       filename,
		ContentType:    strings.TrimSpace(input.ContentType),
		FetchedAt:      input.FetchedAt,
		ContentHash:    contentHash,
		RawContent:     string(input.Content),
		Status:         model.SourceConfigStatusParsed,
		TotalNodes:     len(parseResult.Nodes) + len(parseResult.Issues),
		ValidNodes:     len(parseResult.Nodes),
		InvalidNodes:   len(parseResult.Issues),
		DuplicateNodes: duplicateCount,
		ImportedNodes:  0,
		UploadedBy:     fallbackString(input.UploadedBy, "unknown"),
		UploadedByID:   input.UploadedByID,
	}
	if err := model.DB.Create(sourceConfig).Error; err != nil {
		return nil, err
	}

	return &SourceParseResponse{
		SourceConfig: *sourceConfig,
		Summary: ParseSummary{
			TotalNodes:      sourceConfig.TotalNodes,
			ValidNodes:      sourceConfig.ValidNodes,
			InvalidNodes:    sourceConfig.InvalidNodes,
			DuplicateNodes:  duplicateCount,
			ImportableNodes: importableCount,
		},
		Nodes:  previewNodes,
		Errors: parseResult.Issues,
	}, nil
}

func ImportSourceConfig(sourceConfigID int, fingerprints []string) (*SourceImportResult, error) {
	sourceConfig, err := model.GetSourceConfigByID(sourceConfigID)
	if err != nil {
		return nil, fmt.Errorf("导入记录不存在")
	}

	parseResult, err := proxypkg.ParseYAML([]byte(sourceConfig.RawContent))
	if err != nil {
		return nil, err
	}

	selectedNodes := filterParsedNodes(parseResult.Nodes, fingerprints)
	if len(selectedNodes) == 0 {
		return nil, fmt.Errorf("当前没有可导入的节点")
	}

	existingFingerprints, err := model.FindExistingNodeFingerprints(parsedFingerprints(parseResult.Nodes))
	if err != nil {
		return nil, err
	}

	batchFingerprints := make(map[string]struct{})
	toCreate := make([]model.ProxyNode, 0, len(selectedNodes))
	skippedCount := 0

	for _, item := range selectedNodes {
		if _, seen := batchFingerprints[item.Fingerprint]; seen {
			skippedCount++
			continue
		}
		batchFingerprints[item.Fingerprint] = struct{}{}

		if _, exists := existingFingerprints[item.Fingerprint]; exists {
			skippedCount++
			continue
		}

		toCreate = append(toCreate, model.ProxyNode{
			SourceConfigID:   sourceConfig.ID,
			SourceConfigName: sourceConfig.Filename,
			Name:             item.Name,
			Type:             item.Type,
			Server:           item.Server,
			Port:             item.Port,
			Fingerprint:      item.Fingerprint,
			MetadataJSON:     item.MetadataJSON,
			Enabled:          true,
			LastTestStatus:   model.NodeTestStatusUnknown,
		})
	}

	if len(toCreate) == 0 {
		sourceConfig.Status = model.SourceConfigStatusImported
		sourceConfig.ImportedNodes = 0
		sourceConfig.DuplicateNodes = skippedCount
		if err := model.DB.Save(sourceConfig).Error; err != nil {
			return nil, err
		}
		return &SourceImportResult{
			SourceConfigID: sourceConfig.ID,
			ImportedNodes:  0,
			SkippedNodes:   skippedCount,
		}, nil
	}

	return persistImportedNodes(sourceConfig, toCreate, skippedCount)
}

func TestSourceConfigNodes(ctx context.Context, sourceConfigID int, input NodeTestInput) ([]NodeTestExecution, error) {
	sourceConfig, err := model.GetSourceConfigByID(sourceConfigID)
	if err != nil {
		return nil, fmt.Errorf("导入记录不存在")
	}

	parseResult, err := proxypkg.ParseYAML([]byte(sourceConfig.RawContent))
	if err != nil {
		return nil, err
	}

	selectedNodes := filterParsedNodes(parseResult.Nodes, input.NodeFingerprints)
	if len(selectedNodes) == 0 {
		return nil, fmt.Errorf("请先选择要测速的节点")
	}

	timeout := normalizeNodeTestTimeout(input.TimeoutMS)
	testURL := normalizeNodeTestURL(input.TestURL)

	results := make([]NodeTestExecution, 0, len(selectedNodes))
	for _, item := range selectedNodes {
		execution := executeMetadataNodeTest(ctx, metadataNodeTestInput{
			Name:         item.Name,
			Server:       item.Server,
			Port:         item.Port,
			MetadataJSON: item.MetadataJSON,
			Timeout:      timeout,
			TestURL:      testURL,
		})
		results = append(results, execution)
	}
	return results, nil
}

func persistImportedNodes(sourceConfig *model.SourceConfig, nodes []model.ProxyNode, skippedCount int) (*SourceImportResult, error) {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&nodes).Error; err != nil {
			return err
		}
		sourceConfig.Status = model.SourceConfigStatusImported
		sourceConfig.ImportedNodes = len(nodes)
		sourceConfig.DuplicateNodes = skippedCount
		return tx.Save(sourceConfig).Error
	})
	if err != nil {
		return nil, err
	}

	return &SourceImportResult{
		SourceConfigID: sourceConfig.ID,
		ImportedNodes:  len(nodes),
		SkippedNodes:   skippedCount,
	}, nil
}

func buildPreviewNodes(nodes []proxypkg.ParsedNode, existingFingerprints map[string]struct{}) ([]ParsedNodePreview, int, int) {
	batchFingerprints := make(map[string]struct{})
	previewNodes := make([]ParsedNodePreview, 0, minInt(len(nodes), sourcePreviewLimit))
	duplicateCount := 0
	importableCount := 0

	for _, item := range nodes {
		scope := "none"
		if _, seen := batchFingerprints[item.Fingerprint]; seen {
			scope = "batch"
			duplicateCount++
		} else {
			batchFingerprints[item.Fingerprint] = struct{}{}
			if _, exists := existingFingerprints[item.Fingerprint]; exists {
				scope = "database"
				duplicateCount++
			} else {
				importableCount++
			}
		}

		if len(previewNodes) < sourcePreviewLimit {
			previewNodes = append(previewNodes, ParsedNodePreview{
				Name:           item.Name,
				Type:           item.Type,
				Server:         item.Server,
				Port:           item.Port,
				Fingerprint:    item.Fingerprint,
				DuplicateScope: scope,
			})
		}
	}

	return previewNodes, duplicateCount, importableCount
}

func parsedFingerprints(nodes []proxypkg.ParsedNode) []string {
	items := make([]string, 0, len(nodes))
	for _, item := range nodes {
		items = append(items, item.Fingerprint)
	}
	return items
}

func filterParsedNodes(nodes []proxypkg.ParsedNode, fingerprints []string) []proxypkg.ParsedNode {
	if len(fingerprints) == 0 {
		return nodes
	}
	selected := make(map[string]struct{}, len(fingerprints))
	for _, item := range fingerprints {
		if strings.TrimSpace(item) == "" {
			continue
		}
		selected[strings.TrimSpace(item)] = struct{}{}
	}
	if len(selected) == 0 {
		return nodes
	}

	filtered := make([]proxypkg.ParsedNode, 0, len(nodes))
	for _, item := range nodes {
		if _, ok := selected[item.Fingerprint]; ok {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func checksum(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
