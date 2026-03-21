package runtimeconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"poolx/internal/model"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type MihomoRenderInput struct {
	Profile model.PortProfile
	Nodes   []*model.ProxyNode
}

type RenderResult struct {
	KernelType string `json:"kernel_type"`
	Checksum   string `json:"checksum"`
	Content    string `json:"content"`
}

func RenderMihomoConfig(input MihomoRenderInput) (*RenderResult, error) {
	if len(input.Nodes) == 0 {
		return nil, fmt.Errorf("请先选择至少一个节点")
	}

	proxies := make([]map[string]any, 0, len(input.Nodes))
	proxyNames := make([]string, 0, len(input.Nodes))
	for _, node := range input.Nodes {
		parsed, err := decodeNodeMetadata(node.MetadataJSON)
		if err != nil {
			return nil, fmt.Errorf("解析节点 %s 元数据失败: %v", node.Name, err)
		}
		parsed["name"] = node.Name
		proxies = append(proxies, parsed)
		proxyNames = append(proxyNames, node.Name)
	}

	groupName := strings.TrimSpace(input.Profile.StrategyGroupName)
	if groupName == "" {
		groupName = "POOLX"
	}

	config := map[string]any{
		"allow-lan":    false,
		"bind-address": fallbackString(strings.TrimSpace(input.Profile.ListenHost), "127.0.0.1"),
		"mode":         "rule",
		"log-level":    "info",
		"proxies":      proxies,
		"rules": []string{
			fmt.Sprintf("MATCH,%s", groupName),
		},
	}

	if input.Profile.MixedPort > 0 {
		config["mixed-port"] = input.Profile.MixedPort
	}
	if input.Profile.SocksPort > 0 {
		config["socks-port"] = input.Profile.SocksPort
	}
	if input.Profile.HTTPPort > 0 {
		config["port"] = input.Profile.HTTPPort
	}

	config["proxy-groups"] = []map[string]any{buildStrategyGroup(input.Profile, groupName, proxyNames)}

	contentBytes, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	content := string(contentBytes)
	sum := sha256.Sum256(contentBytes)
	return &RenderResult{
		KernelType: "mihomo",
		Checksum:   hex.EncodeToString(sum[:]),
		Content:    content,
	}, nil
}

func buildStrategyGroup(profile model.PortProfile, groupName string, proxyNames []string) map[string]any {
	result := map[string]any{
		"name":    groupName,
		"type":    profile.StrategyType,
		"proxies": proxyNames,
	}

	switch profile.StrategyType {
	case model.PortProfileStrategyURLTest, model.PortProfileStrategyFallback, model.PortProfileStrategyLoadBalance:
		result["url"] = fallbackString(strings.TrimSpace(profile.TestURL), "https://cp.cloudflare.com/generate_204")
		result["interval"] = normalizePositive(profile.TestIntervalSeconds, 300)
	default:
		result["type"] = model.PortProfileStrategySelect
	}

	if profile.StrategyType == model.PortProfileStrategyLoadBalance {
		result["strategy"] = "consistent-hashing"
	}

	return result
}

func decodeNodeMetadata(raw string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	if err == nil && len(result) > 0 {
		return result, nil
	}

	// Fall back to YAML decoding for any legacy payloads that were not stored as JSON.
	err = yaml.Unmarshal([]byte(raw), &result)
	return result, err
}

func fallbackString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizePositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func SortNodeIDs(ids []int) []int {
	result := append([]int(nil), ids...)
	sort.Ints(result)
	return result
}
