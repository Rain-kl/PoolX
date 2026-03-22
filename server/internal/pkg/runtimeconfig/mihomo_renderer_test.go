package runtimeconfig

import (
	"strings"
	"testing"

	"poolx/internal/model"
)

func TestRenderMihomoConfigBuildsMergeableFragment(t *testing.T) {
	result, err := RenderMihomoConfig(MihomoRenderInput{
		Profile: model.PortProfile{
			ID:         12,
			Name:       "POOLX-FALLBACK",
			ListenHost: "127.0.0.1",
			MixedPort:  7890,
			SocksPort:  7891,
			ProxySettings: model.PortProfileProxySettings{
				StrategyType:        model.PortProfileStrategyFallback,
				TestURL:             "https://cp.cloudflare.com/generate_204",
				TestIntervalSeconds: 180,
				UDPEnabled:          true,
			},
			IncludeInRuntime: true,
			KernelType:       "mihomo",
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
			Name:       "POOLX-LB",
			ListenHost: "127.0.0.1",
			MixedPort:  7890,
			ProxySettings: model.PortProfileProxySettings{
				StrategyType:          model.PortProfileStrategyLoadBalance,
				TestURL:               "https://www.gstatic.com/generate_204",
				TestIntervalSeconds:   300,
				LoadBalanceStrategy:   "round-robin",
				LoadBalanceLazy:       true,
				LoadBalanceDisableUDP: true,
				UDPEnabled:            true,
			},
			IncludeInRuntime: true,
			KernelType:       "mihomo",
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

func TestRenderMihomoConfigIncludesListenerAuthAndUDP(t *testing.T) {
	result, err := RenderMihomoConfig(MihomoRenderInput{
		Profile: model.PortProfile{
			Name:       "auth-workspace",
			ListenHost: "127.0.0.1",
			MixedPort:  7890,
			ProxySettings: model.PortProfileProxySettings{
				StrategyType: model.PortProfileStrategyFallback,
				UDPEnabled:   true,
				AuthEnabled:  true,
				AuthUsername: "username1",
				AuthPassword: "password1",
			},
			IncludeInRuntime: true,
		},
		Nodes: []*model.ProxyNode{
			{
				Name:         "auth-node",
				MetadataJSON: `{"type":"ss","server":"4.4.4.4","port":443,"cipher":"aes-128-gcm","password":"secret"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderMihomoConfig returned error: %v", err)
	}
	if !strings.Contains(result.Content, "udp: true") {
		t.Fatalf("expected udp flag, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "users:") || !strings.Contains(result.Content, "username: username1") {
		t.Fatalf("expected listener auth users, got: %s", result.Content)
	}
}
