package service

import (
	"errors"
	"fmt"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"poolx/internal/pkg/runtimeconfig"
	"sort"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultPortProfileTestURL   = "https://cp.cloudflare.com/generate_204"
	defaultPortProfileGroupName = "POOLX"
)

type PortProfilePayload struct {
	Name                string `json:"name"`
	ListenHost          string `json:"listen_host"`
	MixedPort           int    `json:"mixed_port"`
	SocksPort           int    `json:"socks_port"`
	HTTPPort            int    `json:"http_port"`
	StrategyType        string `json:"strategy_type"`
	StrategyGroupName   string `json:"strategy_group_name"`
	TestURL             string `json:"test_url"`
	TestIntervalSeconds int    `json:"test_interval_seconds"`
	IncludeInRuntime    bool   `json:"include_in_runtime"`
	NodeIDs             []int  `json:"node_ids"`
}

type PortProfilePreview struct {
	Profile    model.PortProfile  `json:"profile"`
	NodeIDs    []int              `json:"node_ids"`
	Nodes      []*model.ProxyNode `json:"nodes"`
	KernelType string             `json:"kernel_type"`
	Checksum   string             `json:"checksum"`
	Content    string             `json:"content"`
}

func ListPortProfiles() ([]*model.PortProfileWithNodes, error) {
	profiles, err := model.ListPortProfiles()
	if err != nil {
		return nil, err
	}

	result := make([]*model.PortProfileWithNodes, 0, len(profiles))
	for _, profile := range profiles {
		item, err := buildPortProfileView(profile)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func GetPortProfile(id int) (*model.PortProfileWithNodes, error) {
	profile, err := model.GetPortProfileByID(id)
	if err != nil {
		return nil, fmt.Errorf("端口配置不存在")
	}
	return buildPortProfileView(profile)
}

func CreatePortProfile(payload PortProfilePayload) (*model.PortProfileWithNodes, error) {
	normalized, err := normalizePortProfilePayload(payload)
	if err != nil {
		return nil, err
	}

	profile := &model.PortProfile{
		Name:                normalized.Name,
		ListenHost:          normalized.ListenHost,
		MixedPort:           normalized.MixedPort,
		SocksPort:           normalized.SocksPort,
		HTTPPort:            normalized.HTTPPort,
		StrategyType:        normalized.StrategyType,
		StrategyGroupName:   normalized.StrategyGroupName,
		TestURL:             normalized.TestURL,
		TestIntervalSeconds: normalized.TestIntervalSeconds,
		IncludeInRuntime:    normalized.IncludeInRuntime,
		KernelType:          common.KernelType,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(profile).Error; err != nil {
			return err
		}
		return replacePortProfileNodes(tx, profile.ID, normalized.NodeIDs)
	}); err != nil {
		return nil, err
	}

	return GetPortProfile(profile.ID)
}

func UpdatePortProfile(id int, payload PortProfilePayload) (*model.PortProfileWithNodes, error) {
	normalized, err := normalizePortProfilePayload(payload)
	if err != nil {
		return nil, err
	}
	profile, err := model.GetPortProfileByID(id)
	if err != nil {
		return nil, fmt.Errorf("端口配置不存在")
	}

	profile.Name = normalized.Name
	profile.ListenHost = normalized.ListenHost
	profile.MixedPort = normalized.MixedPort
	profile.SocksPort = normalized.SocksPort
	profile.HTTPPort = normalized.HTTPPort
	profile.StrategyType = normalized.StrategyType
	profile.StrategyGroupName = normalized.StrategyGroupName
	profile.TestURL = normalized.TestURL
	profile.TestIntervalSeconds = normalized.TestIntervalSeconds
	profile.IncludeInRuntime = normalized.IncludeInRuntime
	profile.KernelType = common.KernelType

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(profile).Error; err != nil {
			return err
		}
		if err := replacePortProfileNodes(tx, profile.ID, normalized.NodeIDs); err != nil {
			return err
		}
		return tx.Where("port_profile_id = ?", profile.ID).Delete(&model.RuntimeConfig{}).Error
	}); err != nil {
		return nil, err
	}

	return GetPortProfile(profile.ID)
}

func DeletePortProfile(id int) error {
	if id <= 0 {
		return fmt.Errorf("无效的端口配置 ID")
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("port_profile_id = ?", id).Delete(&model.PortProfileNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("port_profile_id = ?", id).Delete(&model.RuntimeConfig{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.PortProfile{}, "id = ?", id).Error
	})
}

func PreviewPortProfile(payload PortProfilePayload) (*PortProfilePreview, error) {
	normalized, err := normalizePortProfilePayload(payload)
	if err != nil {
		return nil, err
	}
	nodes, err := resolveSelectedNodes(normalized.NodeIDs)
	if err != nil {
		return nil, err
	}
	profile := model.PortProfile{
		Name:                normalized.Name,
		ListenHost:          normalized.ListenHost,
		MixedPort:           normalized.MixedPort,
		SocksPort:           normalized.SocksPort,
		HTTPPort:            normalized.HTTPPort,
		StrategyType:        normalized.StrategyType,
		StrategyGroupName:   normalized.StrategyGroupName,
		TestURL:             normalized.TestURL,
		TestIntervalSeconds: normalized.TestIntervalSeconds,
		IncludeInRuntime:    normalized.IncludeInRuntime,
		KernelType:          common.KernelType,
	}

	rendered, err := runtimeconfig.RenderMihomoConfig(runtimeconfig.MihomoRenderInput{
		Profile: profile,
		Nodes:   nodes,
	})
	if err != nil {
		return nil, err
	}

	return &PortProfilePreview{
		Profile:    profile,
		NodeIDs:    normalized.NodeIDs,
		Nodes:      nodes,
		KernelType: rendered.KernelType,
		Checksum:   rendered.Checksum,
		Content:    rendered.Content,
	}, nil
}

func PreviewSavedPortProfile(id int) (*PortProfilePreview, error) {
	view, err := GetPortProfile(id)
	if err != nil {
		return nil, err
	}
	return PreviewPortProfile(PortProfilePayload{
		Name:                view.Profile.Name,
		ListenHost:          view.Profile.ListenHost,
		MixedPort:           view.Profile.MixedPort,
		SocksPort:           view.Profile.SocksPort,
		HTTPPort:            view.Profile.HTTPPort,
		StrategyType:        view.Profile.StrategyType,
		StrategyGroupName:   view.Profile.StrategyGroupName,
		TestURL:             view.Profile.TestURL,
		TestIntervalSeconds: view.Profile.TestIntervalSeconds,
		IncludeInRuntime:    view.Profile.IncludeInRuntime,
		NodeIDs:             view.NodeIDs,
	})
}

func SaveRuntimePreview(id int) (*model.RuntimeConfig, error) {
	preview, err := PreviewSavedPortProfile(id)
	if err != nil {
		return nil, err
	}

	runtimeConfig := &model.RuntimeConfig{
		PortProfileID:  id,
		KernelType:     preview.KernelType,
		Checksum:       preview.Checksum,
		RenderedConfig: preview.Content,
	}
	if err := model.DB.Where("port_profile_id = ?", id).
		Assign(runtimeConfig).
		FirstOrCreate(runtimeConfig).Error; err != nil {
		return nil, err
	}
	return runtimeConfig, nil
}

func ListProxyNodeOptions(keyword string) ([]*model.ProxyNode, error) {
	filter := model.ProxyNodeListFilter{
		Keyword: strings.TrimSpace(keyword),
	}
	items, err := model.ListProxyNodes(0, 200, filter)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left int, right int) bool {
		return strings.ToLower(items[left].Name) < strings.ToLower(items[right].Name)
	})
	return items, nil
}

func buildPortProfileView(profile *model.PortProfile) (*model.PortProfileWithNodes, error) {
	relations, err := model.ListPortProfileNodes(profile.ID)
	if err != nil {
		return nil, err
	}
	nodeIDs := make([]int, 0, len(relations))
	for _, item := range relations {
		nodeIDs = append(nodeIDs, item.ProxyNodeID)
	}

	nodes, err := model.FindProxyNodesByIDs(nodeIDs)
	if err != nil {
		return nil, err
	}

	var runtimeConfig *model.RuntimeConfig
	if existing, runtimeErr := model.GetRuntimeConfigByPortProfileID(profile.ID); runtimeErr == nil {
		runtimeConfig = existing
	} else if !errors.Is(runtimeErr, gorm.ErrRecordNotFound) {
		return nil, runtimeErr
	}

	return &model.PortProfileWithNodes{
		Profile: *profile,
		NodeIDs: nodeIDs,
		Nodes:   nodes,
		Runtime: runtimeConfig,
	}, nil
}

func normalizePortProfilePayload(payload PortProfilePayload) (*PortProfilePayload, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if payload.MixedPort <= 0 {
		return nil, fmt.Errorf("Mixed 端口必须大于 0")
	}
	if payload.SocksPort < 0 || payload.HTTPPort < 0 {
		return nil, fmt.Errorf("端口不能为负数")
	}
	if len(payload.NodeIDs) == 0 {
		return nil, fmt.Errorf("请先选择至少一个节点")
	}

	strategy := strings.TrimSpace(payload.StrategyType)
	switch strategy {
	case model.PortProfileStrategySelect, model.PortProfileStrategyURLTest, model.PortProfileStrategyFallback, model.PortProfileStrategyLoadBalance:
	default:
		strategy = model.PortProfileStrategySelect
	}

	normalized := &PortProfilePayload{
		Name:                name,
		ListenHost:          fallbackPortProfileString(strings.TrimSpace(payload.ListenHost), "127.0.0.1"),
		MixedPort:           payload.MixedPort,
		SocksPort:           payload.SocksPort,
		HTTPPort:            payload.HTTPPort,
		StrategyType:        strategy,
		StrategyGroupName:   fallbackPortProfileString(strings.TrimSpace(payload.StrategyGroupName), defaultPortProfileGroupName),
		TestURL:             fallbackPortProfileString(strings.TrimSpace(payload.TestURL), defaultPortProfileTestURL),
		TestIntervalSeconds: normalizePortProfilePositive(payload.TestIntervalSeconds, 300),
		IncludeInRuntime:    payload.IncludeInRuntime,
		NodeIDs:             deduplicateNodeIDs(payload.NodeIDs),
	}
	if len(normalized.NodeIDs) == 0 {
		return nil, fmt.Errorf("请先选择至少一个节点")
	}
	if err := validatePortProfilePorts(normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func validatePortProfilePorts(payload *PortProfilePayload) error {
	ports := make(map[int]string)
	addPort := func(port int, label string) error {
		if port == 0 {
			return nil
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("%s 端口无效", label)
		}
		if existing, ok := ports[port]; ok {
			return fmt.Errorf("%s 与 %s 端口冲突", label, existing)
		}
		ports[port] = label
		return nil
	}
	if err := addPort(payload.MixedPort, "Mixed"); err != nil {
		return err
	}
	if err := addPort(payload.SocksPort, "SOCKS"); err != nil {
		return err
	}
	if err := addPort(payload.HTTPPort, "HTTP"); err != nil {
		return err
	}
	return nil
}

func resolveSelectedNodes(nodeIDs []int) ([]*model.ProxyNode, error) {
	nodes, err := model.FindProxyNodesByIDs(nodeIDs)
	if err != nil {
		return nil, err
	}
	if len(nodes) != len(nodeIDs) {
		return nil, fmt.Errorf("部分节点不存在，请重新选择")
	}
	nodeMap := make(map[int]*model.ProxyNode, len(nodes))
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}
	ordered := make([]*model.ProxyNode, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		ordered = append(ordered, nodeMap[id])
	}
	return ordered, nil
}

func replacePortProfileNodes(tx *gorm.DB, profileID int, nodeIDs []int) error {
	if err := tx.Where("port_profile_id = ?", profileID).Delete(&model.PortProfileNode{}).Error; err != nil {
		return err
	}
	items := make([]model.PortProfileNode, 0, len(nodeIDs))
	for index, nodeID := range nodeIDs {
		items = append(items, model.PortProfileNode{
			PortProfileID: profileID,
			ProxyNodeID:   nodeID,
			SortOrder:     index,
		})
	}
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}

func deduplicateNodeIDs(nodeIDs []int) []int {
	seen := make(map[int]struct{}, len(nodeIDs))
	result := make([]int, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func fallbackPortProfileString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizePortProfilePositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
