package runtimeconfig

import (
	"strings"
	"testing"

	"poolx/internal/model"
)

func TestRenderFinalMihomoConfigAggregatesProfiles(t *testing.T) {
	result, err := RenderFinalMihomoConfig(AggregatedMihomoInput{
		AllowLAN:          false,
		Mode:              "rule",
		ControllerAddress: "127.0.0.1:19090",
		ControllerSecret:  "secret",
		Profiles: []*model.PortProfileWithNodes{
			{
				Profile: model.PortProfile{
					ID:         1,
					Name:       "GROUP-A",
					ListenHost: "127.0.0.1",
					MixedPort:  7890,
					ProxySettings: model.PortProfileProxySettings{
						StrategyType: model.PortProfileStrategySelect,
						UDPEnabled:   true,
					},
					IncludeInRuntime: true,
				},
				Nodes: []*model.ProxyNode{
					{
						ID:           11,
						Name:         "hk-1",
						MetadataJSON: `{"type":"ss","server":"1.1.1.1","port":443,"cipher":"aes-128-gcm","password":"secret"}`,
					},
				},
			},
			{
				Profile: model.PortProfile{
					ID:         2,
					Name:       "GROUP-A",
					ListenHost: "127.0.0.1",
					SocksPort:  7891,
					ProxySettings: model.PortProfileProxySettings{
						StrategyType:        model.PortProfileStrategyFallback,
						TestURL:             "https://cp.cloudflare.com/generate_204",
						TestIntervalSeconds: 300,
						UDPEnabled:          true,
					},
					IncludeInRuntime: true,
				},
				Nodes: []*model.ProxyNode{
					{
						ID:           12,
						Name:         "hk-1",
						MetadataJSON: `{"type":"ss","server":"2.2.2.2","port":443,"cipher":"aes-128-gcm","password":"secret-2"}`,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderFinalMihomoConfig returned error: %v", err)
	}
	if result.ListenerCount != 2 {
		t.Fatalf("expected 2 listeners, got %d", result.ListenerCount)
	}
	if !strings.Contains(result.Content, "external-controller: 127.0.0.1:19090") {
		t.Fatalf("expected controller in final config, got %s", result.Content)
	}
	if !strings.Contains(result.Content, "listeners:") {
		t.Fatalf("expected listeners section, got %s", result.Content)
	}
	if !strings.Contains(result.Content, "proxy: GROUP-A-2") {
		t.Fatalf("expected second group name to be uniqued, got %s", result.Content)
	}
}

func TestRenderFinalMihomoConfigSkipsProfilesExcludedFromRuntime(t *testing.T) {
	result, err := RenderFinalMihomoConfig(AggregatedMihomoInput{
		AllowLAN:          false,
		Mode:              "rule",
		ControllerAddress: "127.0.0.1:19090",
		ControllerSecret:  "secret",
		Profiles: []*model.PortProfileWithNodes{
			{
				Profile: model.PortProfile{
					ID:               1,
					Name:             "excluded",
					ListenHost:       "127.0.0.1",
					MixedPort:        7890,
					IncludeInRuntime: false,
				},
				Nodes: []*model.ProxyNode{
					{
						ID:           11,
						Name:         "hk-1",
						MetadataJSON: `{"type":"ss","server":"1.1.1.1","port":443,"cipher":"aes-128-gcm","password":"secret"}`,
					},
				},
			},
			{
				Profile: model.PortProfile{
					ID:               2,
					Name:             "included",
					ListenHost:       "127.0.0.1",
					MixedPort:        7891,
					IncludeInRuntime: true,
				},
				Nodes: []*model.ProxyNode{
					{
						ID:           12,
						Name:         "jp-1",
						MetadataJSON: `{"type":"ss","server":"2.2.2.2","port":443,"cipher":"aes-128-gcm","password":"secret-2"}`,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderFinalMihomoConfig returned error: %v", err)
	}
	if result.ListenerCount != 1 {
		t.Fatalf("expected 1 listener, got %d", result.ListenerCount)
	}
	if strings.Contains(result.Content, "7890") {
		t.Fatalf("expected excluded listener to be omitted, got %s", result.Content)
	}
	if !strings.Contains(result.Content, "7891") {
		t.Fatalf("expected included listener to remain, got %s", result.Content)
	}
}

func TestRenderFinalMihomoConfigUsesConfiguredClashSettings(t *testing.T) {
	result, err := RenderFinalMihomoConfig(AggregatedMihomoInput{
		AllowLAN:          true,
		Mode:              "global",
		ControllerAddress: "0.0.0.0:29090",
		ControllerSecret:  "fixed-secret",
		Profiles: []*model.PortProfileWithNodes{
			{
				Profile: model.PortProfile{
					ID:               2,
					Name:             "included",
					ListenHost:       "127.0.0.1",
					MixedPort:        7891,
					IncludeInRuntime: true,
				},
				Nodes: []*model.ProxyNode{
					{
						ID:           12,
						Name:         "jp-1",
						MetadataJSON: `{"type":"ss","server":"2.2.2.2","port":443,"cipher":"aes-128-gcm","password":"secret-2"}`,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderFinalMihomoConfig returned error: %v", err)
	}
	if !strings.Contains(result.Content, "allow-lan: true") {
		t.Fatalf("expected allow-lan in final config, got %s", result.Content)
	}
	if !strings.Contains(result.Content, "mode: global") {
		t.Fatalf("expected mode in final config, got %s", result.Content)
	}
	if !strings.Contains(result.Content, "external-controller: 0.0.0.0:29090") {
		t.Fatalf("expected controller address in final config, got %s", result.Content)
	}
	if !strings.Contains(result.Content, "secret: fixed-secret") {
		t.Fatalf("expected secret in final config, got %s", result.Content)
	}
}
