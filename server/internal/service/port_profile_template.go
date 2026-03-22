package service

import (
	"encoding/json"
	"fmt"
	"poolx/internal/model"
	"strings"
)

type PortProfileTemplateView struct {
	Template model.PortProfileTemplate `json:"template"`
	NodeIDs  []int                     `json:"node_ids"`
}

func ListPortProfileTemplates() ([]*PortProfileTemplateView, error) {
	items, err := model.ListPortProfileTemplates()
	if err != nil {
		return nil, err
	}
	result := make([]*PortProfileTemplateView, 0, len(items))
	for _, item := range items {
		view, err := buildPortProfileTemplateView(item)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

func SavePortProfileTemplate(name string, payload PortProfilePayload) (*PortProfileTemplateView, error) {
	normalized, err := normalizePortProfilePayload(payload)
	if err != nil {
		return nil, err
	}
	templateName := strings.TrimSpace(name)
	if templateName == "" {
		templateName = normalized.Name
	}
	if templateName == "" {
		return nil, fmt.Errorf("模板名称不能为空")
	}
	nodeIDsJSON, err := json.Marshal(normalized.NodeIDs)
	if err != nil {
		return nil, fmt.Errorf("序列化模板节点失败: %v", err)
	}
	template := &model.PortProfileTemplate{
		Name:              templateName,
		ListenHost:        normalized.ListenHost,
		MixedPort:         normalized.MixedPort,
		SocksPort:         normalized.SocksPort,
		HTTPPort:          normalized.HTTPPort,
		ProxySettingsJSON: mustEncodePortProfileProxySettings(normalized.ProxySettings),
		ProxySettings:     normalized.ProxySettings,
		IncludeInRuntime:  normalized.IncludeInRuntime,
		NodeIDsJSON:       string(nodeIDsJSON),
	}
	if err := model.DB.Create(template).Error; err != nil {
		return nil, err
	}
	return buildPortProfileTemplateView(template)
}

func DeletePortProfileTemplate(id int) error {
	if id <= 0 {
		return fmt.Errorf("无效的模板 ID")
	}
	return model.DB.Delete(&model.PortProfileTemplate{}, "id = ?", id).Error
}

func buildPortProfileTemplateView(template *model.PortProfileTemplate) (*PortProfileTemplateView, error) {
	if err := template.HydrateProxySettings(); err != nil {
		return nil, fmt.Errorf("解析模板代理设置失败: %v", err)
	}
	var nodeIDs []int
	if err := json.Unmarshal([]byte(template.NodeIDsJSON), &nodeIDs); err != nil {
		return nil, fmt.Errorf("解析模板节点失败: %v", err)
	}
	return &PortProfileTemplateView{
		Template: *template,
		NodeIDs:  nodeIDs,
	}, nil
}
