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

	fragment := map[string]any{
		"fragment_kind": "port-profile",
		"profile": map[string]any{
			"id":          input.Profile.ID,
			"name":        input.Profile.Name,
			"enabled":     input.Profile.Enabled,
			"kernel":      fallbackString(strings.TrimSpace(input.Profile.KernelType), "mihomo"),
			"listen_host": fallbackString(strings.TrimSpace(input.Profile.ListenHost), "127.0.0.1"),
		},
		"listeners": buildListeners(input.Profile),
		"strategy": map[string]any{
			"group_name":    groupName,
			"type":          normalizeStrategyType(input.Profile.StrategyType),
			"test_url":      fallbackString(strings.TrimSpace(input.Profile.TestURL), "https://cp.cloudflare.com/generate_204"),
			"test_interval": normalizePositive(input.Profile.TestIntervalSeconds, 300),
		},
		"proxies": proxies,
		"proxy-groups": []map[string]any{
			buildStrategyGroup(input.Profile, groupName, proxyNames),
		},
		"rules": []string{
			fmt.Sprintf("MATCH,%s", groupName),
		},
	}

	contentBytes, err := yaml.Marshal(fragment)
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

func buildListeners(profile model.PortProfile) []map[string]any {
	host := fallbackString(strings.TrimSpace(profile.ListenHost), "127.0.0.1")
	listeners := make([]map[string]any, 0, 3)
	appendListener := func(kind string, port int) {
		if port <= 0 {
			return
		}
		listeners = append(listeners, map[string]any{
			"name":    buildListenerName(profile.Name, kind, port),
			"type":    kind,
			"host":    host,
			"port":    port,
			"enabled": profile.Enabled,
		})
	}
	appendListener("mixed", profile.MixedPort)
	appendListener("socks", profile.SocksPort)
	appendListener("http", profile.HTTPPort)
	return listeners
}

func buildListenerName(profileName string, kind string, port int) string {
	base := strings.TrimSpace(profileName)
	if base == "" {
		base = "port-profile"
	}
	return fmt.Sprintf("%s-%s-%d", base, kind, port)
}

func buildStrategyGroup(profile model.PortProfile, groupName string, proxyNames []string) map[string]any {
	result := map[string]any{
		"name":    groupName,
		"type":    normalizeStrategyType(profile.StrategyType),
		"proxies": proxyNames,
	}

	switch normalizeStrategyType(profile.StrategyType) {
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

func normalizeStrategyType(value string) string {
	switch strings.TrimSpace(value) {
	case model.PortProfileStrategyURLTest, model.PortProfileStrategyFallback, model.PortProfileStrategyLoadBalance:
		return strings.TrimSpace(value)
	default:
		return model.PortProfileStrategySelect
	}
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
