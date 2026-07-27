package runtimeconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Rain-kl/Foam/backend/internal/domain/clash"
	"gopkg.in/yaml.v3"
)

type MihomoRenderInput struct {
	Profile clash.PortProfile
	Nodes   []*clash.ProxyNode
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

	proxySettings := input.Profile.ProxySettings
	groupName := strings.TrimSpace(input.Profile.Name)
	if groupName == "" {
		groupName = "FOAM"
	}

	fragment := map[string]any{
		"fragment_kind": "port-profile",
		"profile": map[string]any{
			"id":                 input.Profile.ID,
			"name":               input.Profile.Name,
			"include_in_runtime": input.Profile.IncludeInRuntime,
			"kernel":             fallbackString(strings.TrimSpace(input.Profile.KernelType), "mihomo"),
			"listen_host":        fallbackString(strings.TrimSpace(input.Profile.ListenHost), "0.0.0.0"),
		},
		"listeners": buildListeners(input.Profile, proxySettings),
		"strategy": map[string]any{
			"group_name":               groupName,
			"type":                     normalizeStrategyType(proxySettings.StrategyType),
			"test_url":                 fallbackString(strings.TrimSpace(proxySettings.TestURL), "https://cp.cloudflare.com/generate_204"),
			"test_interval":            normalizePositive(proxySettings.TestIntervalSeconds, 300),
			"load_balance_mode":        normalizeLoadBalanceStrategy(proxySettings.LoadBalanceStrategy),
			"load_balance_lazy":        proxySettings.LoadBalanceLazy,
			"load_balance_disable_udp": proxySettings.LoadBalanceDisableUDP,
		},
		"proxies": proxies,
		"proxy-groups": []map[string]any{
			buildStrategyGroup(input.Profile, proxySettings, groupName, proxyNames),
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

func buildListeners(profile clash.PortProfile, proxySettings clash.PortProfileProxySettings) []map[string]any {
	host := fallbackString(strings.TrimSpace(profile.ListenHost), "0.0.0.0")
	listeners := make([]map[string]any, 0, 3)
	appendListener := func(kind string, port int) {
		if port <= 0 {
			return
		}
		listener := map[string]any{
			"name":   buildListenerName(profile.Name, kind, port),
			"type":   kind,
			"listen": host,
			"port":   port,
			"udp":    proxySettings.UDPEnabled,
		}
		if proxySettings.AuthEnabled {
			listener["users"] = []map[string]any{
				{
					"username": proxySettings.AuthUsername,
					"password": proxySettings.AuthPassword,
				},
			}
		}
		listeners = append(listeners, listener)
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

func buildStrategyGroup(_ clash.PortProfile, proxySettings clash.PortProfileProxySettings, groupName string, proxyNames []string) map[string]any {
	result := map[string]any{
		"name":    groupName,
		"type":    normalizeStrategyType(proxySettings.StrategyType),
		"proxies": proxyNames,
	}

	switch normalizeStrategyType(proxySettings.StrategyType) {
	case clash.PortProfileStrategyURLTest, clash.PortProfileStrategyFallback, clash.PortProfileStrategyLoadBalance:
		result["url"] = fallbackString(strings.TrimSpace(proxySettings.TestURL), "https://cp.cloudflare.com/generate_204")
		result["interval"] = normalizePositive(proxySettings.TestIntervalSeconds, 300)
	default:
		result["type"] = clash.PortProfileStrategySelect
	}

	if proxySettings.StrategyType == clash.PortProfileStrategyLoadBalance {
		result["strategy"] = normalizeLoadBalanceStrategy(proxySettings.LoadBalanceStrategy)
		if proxySettings.LoadBalanceLazy {
			result["lazy"] = true
		}
		if proxySettings.LoadBalanceDisableUDP {
			result["disable-udp"] = true
		}
	}

	return result
}

func normalizeLoadBalanceStrategy(value string) string {
	switch strings.TrimSpace(value) {
	case "round-robin":
		return "round-robin"
	default:
		return "consistent-hashing"
	}
}

func normalizeStrategyType(value string) string {
	switch strings.TrimSpace(value) {
	case clash.PortProfileStrategyURLTest, clash.PortProfileStrategyFallback, clash.PortProfileStrategyLoadBalance:
		return strings.TrimSpace(value)
	default:
		return clash.PortProfileStrategySelect
	}
}

func decodeNodeMetadata(raw string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	if err == nil && len(result) > 0 {
		return result, nil
	}

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
