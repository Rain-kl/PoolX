package service

import (
	"os"
	"poolx/internal/pkg/common"
	"strings"
)

type KernelCapability struct {
	KernelType            string   `json:"kernel_type"`
	BinaryConfigured      bool     `json:"binary_configured"`
	BinaryExists          bool     `json:"binary_exists"`
	SupportsStart         bool     `json:"supports_start"`
	SupportsStop          bool     `json:"supports_stop"`
	SupportsReload        bool     `json:"supports_reload"`
	SupportsTemplates     bool     `json:"supports_templates"`
	SupportsNodeTags      bool     `json:"supports_node_tags"`
	SupportsAutoRefresh   bool     `json:"supports_auto_refresh"`
	SupportsNodeTestCache bool     `json:"supports_node_test_cache"`
	SupportsNodeTestBatch bool     `json:"supports_node_test_batch"`
	SupportedStrategies   []string `json:"supported_strategies"`
	RuntimeControllerType string   `json:"runtime_controller_type"`
	Message               string   `json:"message"`
}

func GetKernelCapability() *KernelCapability {
	binaryPath := strings.TrimSpace(common.MihomoBinaryPath)
	binaryConfigured := binaryPath != ""
	binaryExists := false
	if binaryConfigured {
		if info, err := os.Stat(binaryPath); err == nil && !info.IsDir() {
			binaryExists = true
		}
	}

	message := "当前内核能力正常。"
	if !binaryConfigured {
		message = "尚未配置 Mihomo 二进制路径，运行控制会被禁用。"
	} else if !binaryExists {
		message = "已配置 Mihomo 二进制路径，但当前文件不存在或不可访问。"
	}

	return &KernelCapability{
		KernelType:            common.KernelType,
		BinaryConfigured:      binaryConfigured,
		BinaryExists:          binaryExists,
		SupportsStart:         binaryExists,
		SupportsStop:          true,
		SupportsReload:        binaryExists,
		SupportsTemplates:     true,
		SupportsNodeTags:      true,
		SupportsAutoRefresh:   true,
		SupportsNodeTestCache: false,
		SupportsNodeTestBatch: true,
		SupportedStrategies: []string{
			"select",
			"url-test",
			"fallback",
			"load-balance",
		},
		RuntimeControllerType: "external-controller",
		Message:               message,
	}
}
