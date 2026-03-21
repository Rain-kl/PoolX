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
