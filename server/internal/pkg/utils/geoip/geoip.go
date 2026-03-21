package geoip

import (
	"fmt"
	"ginnexttemplate/internal/pkg/common"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"
	"unicode"

	ristretto "github.com/dgraph-io/ristretto/v2"
)

var CurrentProvider GeoIPService
var geoCache *providerCache
var providerMutex sync.RWMutex
var providerFactory = newProvider

const (
	ProviderDisabled = "disabled"
	ProviderMaxMind  = "mmdb"
	ProviderIPAPI    = "ip-api"
	ProviderGeoJS    = "geojs"
	ProviderIPInfo   = "ipinfo"
)

type GeoInfo struct {
	ISOCode   string
	Name      string
	Latitude  *float64
	Longitude *float64
}

func init() {
	CurrentProvider = &EmptyProvider{}
	geoCache = newProviderCache(48 * time.Hour)
}

// GeoIPService 接口定义了获取地理位置信息的核心方法。
type GeoIPService interface {
	Name() string
	GetGeoInfo(ip net.IP) (*GeoInfo, error)
	UpdateDatabase() error
	Close() error
}

type cachedGeoInfo struct {
	info      *GeoInfo
	expiresAt time.Time
}

type providerCache struct {
	items    *ristretto.Cache[string, cachedGeoInfo]
	duration time.Duration
}

func newProviderCache(duration time.Duration) *providerCache {
	items, err := ristretto.NewCache(&ristretto.Config[string, cachedGeoInfo]{
		NumCounters: 1e5,
		MaxCost:     2e4,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	return &providerCache{
		items:    items,
		duration: duration,
	}
}

func (c *providerCache) Get(key string) (*GeoInfo, bool) {
	entry, ok := c.items.Get(key)
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		c.items.Del(key)
		return nil, false
	}
	return entry.info, true
}

func (c *providerCache) Set(key string, info *GeoInfo) {
	c.items.Set(key, cachedGeoInfo{
		info:      info,
		expiresAt: time.Now().Add(c.duration),
	}, 1)
	c.items.Wait()
}

func (c *providerCache) Flush() {
	c.items.Clear()
}

func GetRegionUnicodeEmoji(isoCode string) string {
	if len(isoCode) != 2 {
		return ""
	}
	isoCode = strings.ToUpper(isoCode)

	if !unicode.IsLetter(rune(isoCode[0])) || !unicode.IsLetter(rune(isoCode[1])) {
		return ""
	}

	rune1 := rune(0x1F1E6 + (rune(isoCode[0]) - 'A'))
	rune2 := rune(0x1F1E6 + (rune(isoCode[1]) - 'A'))
	return string(rune1) + string(rune2)
}

func InitGeoIP() {
	providerName := normalizeProvider(common.GeoIPProvider)
	nextProvider, err := providerFactory(providerName)
	if err != nil {
		slog.Error("initialize GeoIP provider failed", "provider", providerName, "error", err)
		nextProvider = &EmptyProvider{}
	}
	setProvider(nextProvider)
	if providerName == ProviderDisabled {
		slog.Info("GeoIP provider disabled")
		return
	}
	slog.Info("GeoIP provider configured", "provider", CurrentProvider.Name())
}

func GetGeoInfo(ip net.IP) (*GeoInfo, error) {
	if ip == nil {
		return nil, fmt.Errorf("IP address cannot be nil")
	}
	provider := getProvider()
	cacheKey := provider.Name() + ":" + ip.String()

	if cachedInfo, found := geoCache.Get(cacheKey); found {
		return cachedInfo, nil
	}

	info, err := provider.GetGeoInfo(ip)
	if err == nil && info != nil {
		geoCache.Set(cacheKey, info)
	}
	return info, err
}

func LookupGeoInfoWithProvider(providerName string, ip net.IP) (*GeoInfo, error) {
	if ip == nil {
		return nil, fmt.Errorf("IP address cannot be nil")
	}

	provider, err := providerFactory(normalizeProvider(providerName))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := provider.Close(); closeErr != nil {
			slog.Warn("close temporary GeoIP provider failed", "provider", provider.Name(), "error", closeErr)
		}
	}()

	return provider.GetGeoInfo(ip)
}

func UpdateDatabase() error {
	err := getProvider().UpdateDatabase()
	if err == nil {
		geoCache.Flush()
		slog.Info("GeoIP cache cleared due to database update.")
	}
	return err
}

func IsValidProvider(provider string) bool {
	switch normalizeProvider(provider) {
	case ProviderDisabled, ProviderMaxMind, ProviderIPAPI, ProviderGeoJS, ProviderIPInfo:
		return true
	default:
		return false
	}
}

func normalizeProvider(provider string) string {
	normalized := strings.TrimSpace(strings.ToLower(provider))
	if normalized == "" {
		return ProviderDisabled
	}
	return normalized
}

func newProvider(provider string) (GeoIPService, error) {
	switch provider {
	case ProviderDisabled:
		return &EmptyProvider{}, nil
	case ProviderMaxMind:
		return NewMaxMindGeoIPService()
	case ProviderIPAPI:
		return NewIPAPIService()
	case ProviderGeoJS:
		return NewGeoJSService()
	case ProviderIPInfo:
		return NewIPInfoService()
	default:
		return nil, fmt.Errorf("unsupported GeoIP provider %q", provider)
	}
}

func setProvider(provider GeoIPService) {
	providerMutex.Lock()
	previous := CurrentProvider
	CurrentProvider = provider
	providerMutex.Unlock()
	geoCache.Flush()
	if previous != nil && previous != provider {
		if err := previous.Close(); err != nil {
			slog.Warn("close previous GeoIP provider failed", "error", err)
		}
	}
}

func getProvider() GeoIPService {
	providerMutex.RLock()
	defer providerMutex.RUnlock()
	if CurrentProvider == nil {
		return &EmptyProvider{}
	}
	return CurrentProvider
}

func float64Pointer(value float64) *float64 {
	return &value
}

func ProviderFactoryForTest() func(string) (GeoIPService, error) {
	return providerFactory
}

func SetProviderFactoryForTest(factory func(string) (GeoIPService, error)) {
	if factory == nil {
		providerFactory = newProvider
		return
	}
	providerFactory = factory
}
