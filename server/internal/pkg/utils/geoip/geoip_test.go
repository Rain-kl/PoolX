package geoip

import (
	"net"
	"testing"
)

type fakeProvider struct {
	calls int
}

func (f *fakeProvider) Name() string {
	return "fake"
}

func (f *fakeProvider) GetGeoInfo(ip net.IP) (*GeoInfo, error) {
	f.calls++
	return &GeoInfo{
		ISOCode: "CN",
		Name:    "China",
	}, nil
}

func (f *fakeProvider) UpdateDatabase() error {
	return nil
}

func (f *fakeProvider) Close() error {
	return nil
}

func TestGetGeoInfoCachesByProviderAndIP(t *testing.T) {
	originalProvider := CurrentProvider
	geoCache.Flush()
	fake := &fakeProvider{}
	CurrentProvider = fake
	defer func() {
		CurrentProvider = originalProvider
	}()

	ip := net.ParseIP("8.8.8.8")
	record, err := GetGeoInfo(ip)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if record == nil || record.ISOCode != "CN" {
		t.Fatalf("expected cached record, got %#v", record)
	}

	_, err = GetGeoInfo(ip)
	if err != nil {
		t.Fatalf("expected nil error on second call, got %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", fake.calls)
	}
}

func TestUnicodeEmoji(t *testing.T) {
	emoji := GetRegionUnicodeEmoji("CN")
	if emoji != "🇨🇳" {
		t.Errorf("expected emoji for CN, got %s", emoji)
	}
}

func TestIsValidProvider(t *testing.T) {
	cases := map[string]bool{
		"disabled": true,
		"mmdb":     true,
		"ip-api":   true,
		"geojs":    true,
		"ipinfo":   true,
		"unknown":  false,
	}

	for provider, want := range cases {
		if got := IsValidProvider(provider); got != want {
			t.Fatalf("provider %s validity mismatch: want %v, got %v", provider, want, got)
		}
	}
}

func TestLookupGeoInfoWithProviderUsesTemporaryProvider(t *testing.T) {
	previousFactory := providerFactory
	providerFactory = func(provider string) (GeoIPService, error) {
		return &fakeProvider{}, nil
	}
	defer func() {
		providerFactory = previousFactory
	}()

	info, err := LookupGeoInfoWithProvider("ipinfo", net.ParseIP("8.8.8.8"))
	if err != nil {
		t.Fatalf("expected lookup to succeed, got %v", err)
	}
	if info == nil || info.ISOCode != "CN" || info.Name != "China" {
		t.Fatalf("unexpected geo info: %#v", info)
	}
}
