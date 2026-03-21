package service

import (
	"fmt"
	"net"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"poolx/internal/pkg/utils"
	"poolx/internal/pkg/utils/geoip"
	"regexp"
	"strconv"
	"strings"
)

var removedTemplateOptionKeys = map[string]struct{}{
	"AgentDiscoveryToken":                   {},
	"AgentHeartbeatInterval":                {},
	"NodeOfflineThreshold":                  {},
	"AgentUpdateRepo":                       {},
	"DatabaseAutoCleanupEnabled":            {},
	"DatabaseAutoCleanupRetentionDays":      {},
	"OpenRestyWorkerProcesses":              {},
	"OpenRestyWorkerConnections":            {},
	"OpenRestyWorkerRlimitNofile":           {},
	"OpenRestyEventsUse":                    {},
	"OpenRestyEventsMultiAcceptEnabled":     {},
	"OpenRestyKeepaliveTimeout":             {},
	"OpenRestyKeepaliveRequests":            {},
	"OpenRestyClientHeaderTimeout":          {},
	"OpenRestyClientBodyTimeout":            {},
	"OpenRestyClientMaxBodySize":            {},
	"OpenRestyLargeClientHeaderBuffers":     {},
	"OpenRestySendTimeout":                  {},
	"OpenRestyProxyConnectTimeout":          {},
	"OpenRestyProxySendTimeout":             {},
	"OpenRestyProxyReadTimeout":             {},
	"OpenRestyWebsocketEnabled":             {},
	"OpenRestyProxyRequestBufferingEnabled": {},
	"OpenRestyProxyBufferingEnabled":        {},
	"OpenRestyProxyBuffers":                 {},
	"OpenRestyProxyBufferSize":              {},
	"OpenRestyProxyBusyBuffersSize":         {},
	"OpenRestyGzipEnabled":                  {},
	"OpenRestyGzipMinLength":                {},
	"OpenRestyGzipCompLevel":                {},
	"OpenRestyCacheEnabled":                 {},
	"OpenRestyCachePath":                    {},
	"OpenRestyCacheLevels":                  {},
	"OpenRestyCacheInactive":                {},
	"OpenRestyCacheMaxSize":                 {},
	"OpenRestyCacheKeyTemplate":             {},
	"OpenRestyCacheLockEnabled":             {},
	"OpenRestyCacheLockTimeout":             {},
	"OpenRestyCacheUseStale":                {},
	"OpenRestyMainConfigTemplate":           {},
	"OpenRestyResolvers":                    {},
}

var githubRepoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

type GeoIPLookupPreview struct {
	Provider  string   `json:"provider"`
	IP        string   `json:"ip"`
	ISOCode   string   `json:"iso_code"`
	Name      string   `json:"name"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

func ListEditableOptions() []*model.Option {
	options := make([]*model.Option, 0, len(common.OptionMap))
	common.OptionMapRWMutex.Lock()
	defer common.OptionMapRWMutex.Unlock()

	for key, value := range common.OptionMap {
		if strings.Contains(key, "Token") || strings.Contains(key, "Secret") {
			continue
		}
		options = append(options, &model.Option{
			Key:   key,
			Value: utils.Interface2String(value),
		})
	}
	return options
}

func UpdateEditableOption(option model.Option) error {
	switch option.Key {
	case "GitHubOAuthEnabled":
		if option.Value == "true" && common.GitHubClientId == "" {
			return fmt.Errorf("GitHub OAuth requires GitHub client configuration first")
		}
	case "WeChatAuthEnabled":
		if option.Value == "true" && common.WeChatServerAddress == "" {
			return fmt.Errorf("WeChat auth requires WeChat server configuration first")
		}
	case "TurnstileCheckEnabled":
		if option.Value == "true" && common.TurnstileSiteKey == "" {
			return fmt.Errorf("Turnstile requires site key and secret key first")
		}
	case "ServerUpdateRepo":
		if !isValidGitHubRepo(option.Value) {
			return fmt.Errorf("ServerUpdateRepo must be in owner/repo format")
		}
	case "KernelType":
		if err := validateKernelType(option.Value); err != nil {
			return err
		}
	case "MihomoBinaryPath":
		if strings.TrimSpace(option.Value) == "" {
			return fmt.Errorf("MihomoBinaryPath cannot be empty")
		}
	case "MihomoBinarySource":
		if err := validateMihomoBinarySource(option.Value); err != nil {
			return err
		}
	case "GeoIPProvider":
		if !geoip.IsValidProvider(option.Value) {
			return fmt.Errorf("GeoIPProvider is invalid")
		}
	}
	if _, removed := removedTemplateOptionKeys[option.Key]; removed {
		return fmt.Errorf("%s has been removed from PoolX options", option.Key)
	}
	if err := validateRateLimitOption(option.Key, option.Value); err != nil {
		return err
	}
	if err := model.UpdateOption(option.Key, option.Value); err != nil {
		return err
	}
	if option.Key == "GeoIPProvider" {
		geoip.InitGeoIP()
	}
	return nil
}

func validateRateLimitOption(key string, value string) error {
	maxDurationSeconds := int(common.RateLimitKeyExpirationDuration.Seconds())

	switch key {
	case "GlobalApiRateLimitNum", "GlobalWebRateLimitNum", "UploadRateLimitNum", "DownloadRateLimitNum", "CriticalRateLimitNum":
		intValue, err := strconv.Atoi(value)
		if err != nil || intValue <= 0 {
			return fmt.Errorf("%s must be a positive integer", key)
		}
		return nil
	case "GlobalApiRateLimitDuration", "GlobalWebRateLimitDuration", "UploadRateLimitDuration", "DownloadRateLimitDuration", "CriticalRateLimitDuration":
		intValue, err := strconv.Atoi(value)
		if err != nil || intValue <= 0 {
			return fmt.Errorf("%s must be a positive integer", key)
		}
		if intValue > maxDurationSeconds {
			return fmt.Errorf("%s cannot exceed %d seconds", key, maxDurationSeconds)
		}
		return nil
	default:
		return nil
	}
}

func isValidGitHubRepo(value string) bool {
	return githubRepoPattern.MatchString(strings.TrimSpace(value))
}

func validateKernelType(value string) error {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case KernelTypeMihomo:
		return nil
	case KernelTypeXray, KernelTypeSingbox:
		return fmt.Errorf("当前版本仅支持 mihomo，%s 暂未开放", strings.TrimSpace(value))
	default:
		return fmt.Errorf("KernelType is invalid")
	}
}

func validateMihomoBinarySource(value string) error {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case MihomoBinarySourceExisting, MihomoBinarySourceUpload, MihomoBinarySourceDownload:
		return nil
	default:
		return fmt.Errorf("MihomoBinarySource is invalid")
	}
}

func PreviewGeoIPLookup(provider string, ipValue string) (*GeoIPLookupPreview, error) {
	if !geoip.IsValidProvider(provider) {
		return nil, fmt.Errorf("GeoIPProvider is invalid")
	}

	ip := net.ParseIP(strings.TrimSpace(ipValue))
	if ip == nil {
		return nil, fmt.Errorf("IP 地址格式无效")
	}

	info, err := geoip.LookupGeoInfoWithProvider(provider, ip)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("未获取到归属信息")
	}

	return &GeoIPLookupPreview{
		Provider:  strings.TrimSpace(provider),
		IP:        ip.String(),
		ISOCode:   info.ISOCode,
		Name:      info.Name,
		Latitude:  info.Latitude,
		Longitude: info.Longitude,
	}, nil
}
