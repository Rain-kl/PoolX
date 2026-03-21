package runtimeconfig

import (
	"strings"
	"testing"

	"poolx/internal/model"
)

func TestRenderMihomoConfigBuildsMergeableFragment(t *testing.T) {
	result, err := RenderMihomoConfig(MihomoRenderInput{
		Profile: model.PortProfile{
			ID:                  12,
			Name:                "default-workspace",
			ListenHost:          "127.0.0.1",
			MixedPort:           7890,
			SocksPort:           7891,
			StrategyType:        model.PortProfileStrategyFallback,
			StrategyGroupName:   "POOLX-FALLBACK",
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 180,
			IncludeInRuntime:    true,
			KernelType:          "mihomo",
		},
		Nodes: []*model.ProxyNode{
			{
				Name:         "hk-01",
				MetadataJSON: `{"type":"ss","server":"1.1.1.1","port":443,"cipher":"aes-128-gcm","password":"secret"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderMihomoConfig returned error: %v", err)
	}
	if !strings.Contains(result.Content, "fragment_kind: port-profile") {
		t.Fatalf("expected mergeable fragment marker, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "listeners:") {
		t.Fatalf("expected listeners section, got: %s", result.Content)
	}
	if strings.Contains(result.Content, "mixed-port:") {
		t.Fatalf("expected fragment output instead of final config, got: %s", result.Content)
	}
	if strings.Contains(result.Content, "allow-lan:") {
		t.Fatalf("expected fragment output instead of final config, got: %s", result.Content)
	}
}

func TestRenderMihomoConfigIncludesLoadBalanceOptions(t *testing.T) {
	result, err := RenderMihomoConfig(MihomoRenderInput{
		Profile: model.PortProfile{
			Name:                  "lb-workspace",
			ListenHost:            "127.0.0.1",
			MixedPort:             7890,
			StrategyType:          model.PortProfileStrategyLoadBalance,
			StrategyGroupName:     "POOLX-LB",
			TestURL:               "https://www.gstatic.com/generate_204",
			TestIntervalSeconds:   300,
			LoadBalanceStrategy:   "round-robin",
			LoadBalanceLazy:       true,
			LoadBalanceDisableUDP: true,
			IncludeInRuntime:      true,
			KernelType:            "mihomo",
		},
		Nodes: []*model.ProxyNode{
			{
				Name:         "jp-01",
				MetadataJSON: `{"type":"ss","server":"2.2.2.2","port":443,"cipher":"aes-128-gcm","password":"secret"}`,
			},
			{
				Name:         "sg-01",
				MetadataJSON: `{"type":"ss","server":"3.3.3.3","port":443,"cipher":"aes-128-gcm","password":"secret"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderMihomoConfig returned error: %v", err)
	}
	if !strings.Contains(result.Content, "strategy: round-robin") {
		t.Fatalf("expected round-robin strategy, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "lazy: true") {
		t.Fatalf("expected lazy flag, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "disable-udp: true") {
		t.Fatalf("expected disable-udp flag, got: %s", result.Content)
	}
}
