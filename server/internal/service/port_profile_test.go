package service

import (
	"strings"
	"testing"

	"poolx/internal/model"
)

func TestCreatePortProfileUsesStrategyGroupNameAsName(t *testing.T) {
	setupServiceTestDB(t)

	node := seedRuntimeTestNode(t, "profile-name-node")
	profile, err := CreatePortProfile(PortProfilePayload{
		Name:       "POOLX-PRIMARY",
		ListenHost: "127.0.0.1",
		MixedPort:  17890,
		ProxySettings: model.PortProfileProxySettings{
			StrategyType:        model.PortProfileStrategyFallback,
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 300,
			UDPEnabled:          true,
		},
		IncludeInRuntime: true,
		NodeIDs:          []int{node.ID},
	})
	if err != nil {
		t.Fatalf("create port profile: %v", err)
	}

	if profile.Profile.Name != "POOLX-PRIMARY" {
		t.Fatalf("expected profile name to use the submitted name, got %q", profile.Profile.Name)
	}
}

func TestCreatePortProfileRejectsDuplicateStrategyGroupName(t *testing.T) {
	setupServiceTestDB(t)

	firstNode := seedRuntimeTestNode(t, "duplicate-group-node-1")
	secondNode := seedRuntimeTestNode(t, "duplicate-group-node-2")

	if _, err := CreatePortProfile(PortProfilePayload{
		Name:       "POOLX-DUPLICATE",
		ListenHost: "127.0.0.1",
		MixedPort:  17891,
		ProxySettings: model.PortProfileProxySettings{
			StrategyType:        model.PortProfileStrategyFallback,
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 300,
			UDPEnabled:          true,
		},
		IncludeInRuntime: true,
		NodeIDs:          []int{firstNode.ID},
	}); err != nil {
		t.Fatalf("create first port profile: %v", err)
	}

	_, err := CreatePortProfile(PortProfilePayload{
		Name:       "poolx-duplicate",
		ListenHost: "127.0.0.1",
		MixedPort:  17892,
		ProxySettings: model.PortProfileProxySettings{
			StrategyType:        model.PortProfileStrategyFallback,
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 300,
			UDPEnabled:          true,
		},
		IncludeInRuntime: true,
		NodeIDs:          []int{secondNode.ID},
	})
	if err == nil {
		t.Fatal("expected duplicate strategy group name to fail")
	}
	if !strings.Contains(err.Error(), "策略组名称已存在") {
		t.Fatalf("expected duplicate strategy group name error, got %v", err)
	}
}

func TestCreatePortProfileRejectsDuplicateMixedPort(t *testing.T) {
	setupServiceTestDB(t)

	firstNode := seedRuntimeTestNode(t, "duplicate-port-node-1")
	secondNode := seedRuntimeTestNode(t, "duplicate-port-node-2")

	if _, err := CreatePortProfile(PortProfilePayload{
		Name:       "POOLX-ONE",
		ListenHost: "127.0.0.1",
		MixedPort:  17893,
		ProxySettings: model.PortProfileProxySettings{
			StrategyType:        model.PortProfileStrategyFallback,
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 300,
			UDPEnabled:          true,
		},
		IncludeInRuntime: true,
		NodeIDs:          []int{firstNode.ID},
	}); err != nil {
		t.Fatalf("create first port profile: %v", err)
	}

	_, err := CreatePortProfile(PortProfilePayload{
		Name:       "POOLX-TWO",
		ListenHost: "127.0.0.1",
		MixedPort:  17893,
		ProxySettings: model.PortProfileProxySettings{
			StrategyType:        model.PortProfileStrategyFallback,
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 300,
			UDPEnabled:          true,
		},
		IncludeInRuntime: true,
		NodeIDs:          []int{secondNode.ID},
	})
	if err == nil {
		t.Fatal("expected duplicate mixed port to fail")
	}
	if !strings.Contains(err.Error(), "Mixed 端口已被其他端口配置占用") {
		t.Fatalf("expected duplicate mixed port error, got %v", err)
	}
}

func TestCreatePortProfileIncludesProxySettingsInJSON(t *testing.T) {
	setupServiceTestDB(t)

	node := seedRuntimeTestNode(t, "json-settings-node")
	profile, err := CreatePortProfile(PortProfilePayload{
		Name:       "POOLX-JSON",
		ListenHost: "127.0.0.1",
		MixedPort:  17894,
		ProxySettings: model.PortProfileProxySettings{
			StrategyType:        model.PortProfileStrategyFallback,
			TestURL:             "https://cp.cloudflare.com/generate_204",
			TestIntervalSeconds: 300,
			UDPEnabled:          true,
			AuthEnabled:         true,
			AuthUsername:        "username1",
			AuthPassword:        "password1",
		},
		IncludeInRuntime: true,
		NodeIDs:          []int{node.ID},
	})
	if err != nil {
		t.Fatalf("create port profile: %v", err)
	}
	if !profile.Profile.ProxySettings.AuthEnabled || profile.Profile.ProxySettings.AuthUsername != "username1" {
		t.Fatalf("expected proxy settings to round-trip from json, got %+v", profile.Profile.ProxySettings)
	}
}
