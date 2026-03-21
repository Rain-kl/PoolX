package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"poolx/internal/model"
	proxypkg "poolx/internal/pkg/proxy"
	"strings"

	"gorm.io/gorm"
)

const sourcePreviewLimit = 100

type SourceUploadInput struct {
	Filename     string
	UploadedBy   string
	UploadedByID int
	Content      []byte
}

type ParsedNodePreview struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Server         string `json:"server"`
	Port           int    `json:"port"`
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
		Filename:       filename,
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

func ImportSourceConfig(sourceConfigID int) (*SourceImportResult, error) {
	sourceConfig, err := model.GetSourceConfigByID(sourceConfigID)
	if err != nil {
		return nil, fmt.Errorf("导入记录不存在")
	}

	parseResult, err := proxypkg.ParseYAML([]byte(sourceConfig.RawContent))
	if err != nil {
		return nil, err
	}

	existingFingerprints, err := model.FindExistingNodeFingerprints(parsedFingerprints(parseResult.Nodes))
	if err != nil {
		return nil, err
	}

	batchFingerprints := make(map[string]struct{})
	toCreate := make([]model.ProxyNode, 0, len(parseResult.Nodes))
	skippedCount := 0

	for _, item := range parseResult.Nodes {
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
