package runtimeconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"poolx/internal/model"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type AggregatedMihomoInput struct {
	Profiles          []*model.PortProfileWithNodes
	ControllerAddress string
	ControllerSecret  string
}

type RuntimeListener struct {
	ProfileID      int    `json:"profile_id"`
	ProfileName    string `json:"profile_name"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Listen         string `json:"listen"`
	Port           int    `json:"port"`
	ProxyGroupName string `json:"proxy_group_name"`
}

type FinalRenderResult struct {
	KernelType    string            `json:"kernel_type"`
	Checksum      string            `json:"checksum"`
	Content       string            `json:"content"`
	ProfileCount  int               `json:"profile_count"`
	ListenerCount int               `json:"listener_count"`
	Listeners     []RuntimeListener `json:"listeners"`
}

func RenderFinalMihomoConfig(input AggregatedMihomoInput) (*FinalRenderResult, error) {
	if len(input.Profiles) == 0 {
		return nil, fmt.Errorf("当前没有可启动的启用端口配置")
	}

	type proxyBinding struct {
		Name string
		Node *model.ProxyNode
	}

	proxies := make([]map[string]any, 0)
	proxyBindings := make(map[int]proxyBinding)
	usedProxyNames := make(map[string]int)
	usedGroupNames := make(map[string]int)
	usedListenerKeys := make(map[string]struct{})
	listeners := make([]map[string]any, 0)
	listenerViews := make([]RuntimeListener, 0)
	groups := make([]map[string]any, 0)

	for _, profile := range input.Profiles {
		if profile == nil || !profile.Profile.Enabled {
			continue
		}
		if len(profile.Nodes) == 0 {
			return nil, fmt.Errorf("端口配置 %s 未绑定可用节点", profile.Profile.Name)
		}

		groupName := reserveUniqueName(
			fallbackString(strings.TrimSpace(profile.Profile.StrategyGroupName), fallbackString(strings.TrimSpace(profile.Profile.Name), "POOLX")),
			usedGroupNames,
		)

		proxyNames := make([]string, 0, len(profile.Nodes))
		for _, node := range profile.Nodes {
			binding, ok := proxyBindings[node.ID]
			if !ok {
				parsed, err := decodeNodeMetadata(node.MetadataJSON)
				if err != nil {
					return nil, fmt.Errorf("解析节点 %s 元数据失败: %v", node.Name, err)
				}
				proxyName := reserveUniqueName(node.Name, usedProxyNames)
				parsed["name"] = proxyName
				proxies = append(proxies, parsed)
				binding = proxyBinding{
					Name: proxyName,
					Node: node,
				}
				proxyBindings[node.ID] = binding
			}
			proxyNames = append(proxyNames, binding.Name)
		}

		groups = append(groups, buildStrategyGroup(profile.Profile, groupName, proxyNames))
		for _, listener := range buildListeners(profile.Profile) {
			key := fmt.Sprintf("%s:%d", listener["listen"], listener["port"])
			if _, exists := usedListenerKeys[key]; exists {
				return nil, fmt.Errorf("监听地址冲突: %s", key)
			}
			usedListenerKeys[key] = struct{}{}
			listener["proxy"] = groupName
			listeners = append(listeners, listener)
			listenerViews = append(listenerViews, RuntimeListener{
				ProfileID:      profile.Profile.ID,
				ProfileName:    profile.Profile.Name,
				Name:           listener["name"].(string),
				Type:           listener["type"].(string),
				Listen:         listener["listen"].(string),
				Port:           listener["port"].(int),
				ProxyGroupName: groupName,
			})
		}
	}

	sort.Slice(listenerViews, func(left int, right int) bool {
		if listenerViews[left].Port == listenerViews[right].Port {
			return listenerViews[left].Name < listenerViews[right].Name
		}
		return listenerViews[left].Port < listenerViews[right].Port
	})

	config := map[string]any{
		"allow-lan":           false,
		"mode":                "rule",
		"log-level":           "info",
		"external-controller": input.ControllerAddress,
		"secret":              input.ControllerSecret,
		"listeners":           listeners,
		"proxies":             proxies,
		"proxy-groups":        groups,
		"rules": []string{
			"MATCH,DIRECT",
		},
	}

	contentBytes, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(contentBytes)
	return &FinalRenderResult{
		KernelType:    "mihomo",
		Checksum:      hex.EncodeToString(sum[:]),
		Content:       string(contentBytes),
		ProfileCount:  len(input.Profiles),
		ListenerCount: len(listenerViews),
		Listeners:     listenerViews,
	}, nil
}

func reserveUniqueName(base string, used map[string]int) string {
	candidate := sanitizeRuntimeName(base)
	if candidate == "" {
		candidate = "poolx"
	}
	if _, exists := used[candidate]; !exists {
		used[candidate] = 1
		return candidate
	}
	for index := used[candidate]; ; index++ {
		next := fmt.Sprintf("%s-%d", candidate, index+1)
		if _, exists := used[next]; !exists {
			used[candidate] = index + 1
			used[next] = 1
			return next
		}
	}
}

func sanitizeRuntimeName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "#", "-", ",", "-", " ", "-")
	return replacer.Replace(value)
}
