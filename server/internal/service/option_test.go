package service

import (
	"net"
	"path/filepath"
	"testing"

	"ginnexttemplate/internal/model"
	"ginnexttemplate/internal/pkg/common"
	"ginnexttemplate/internal/pkg/utils/geoip"
)

type fakeGeoIPProvider struct {
	name string
}

func (f *fakeGeoIPProvider) Name() string { return f.name }

func (f *fakeGeoIPProvider) GetGeoInfo(ip net.IP) (*geoip.GeoInfo, error) { return nil, nil }

func (f *fakeGeoIPProvider) UpdateDatabase() error { return nil }

func (f *fakeGeoIPProvider) Close() error { return nil }

func setupServiceTestDB(t *testing.T) {
	t.Helper()

	originalSQLitePath := common.SQLitePath
	common.SQLitePath = filepath.Join(t.TempDir(), "service-test.db")

	if err := model.InitDB(); err != nil {
		t.Fatalf("init test db: %v", err)
	}
	model.InitOptionMap()

	t.Cleanup(func() {
		if err := model.CloseDB(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
		model.DB = nil
		common.SQLitePath = originalSQLitePath
	})
}

func TestUpdateEditableOptionAppliesServerUpdateRepo(t *testing.T) {
	setupServiceTestDB(t)
	originalRepo := common.ServerUpdateRepo
	t.Cleanup(func() {
		common.ServerUpdateRepo = originalRepo
	})

	err := UpdateEditableOption(model.Option{
		Key:   "ServerUpdateRepo",
		Value: "example/custom-template",
	})
	if err != nil {
		t.Fatalf("update editable option: %v", err)
	}

	if common.ServerUpdateRepo != "example/custom-template" {
		t.Fatalf("unexpected ServerUpdateRepo: %s", common.ServerUpdateRepo)
	}
}

func TestUpdateEditableOptionAppliesGeoIPProvider(t *testing.T) {
	setupServiceTestDB(t)
	originalProvider := common.GeoIPProvider
	originalFactory := geoip.ProviderFactoryForTest()
	t.Cleanup(func() {
		common.GeoIPProvider = originalProvider
		geoip.SetProviderFactoryForTest(originalFactory)
		geoip.InitGeoIP()
	})

	geoip.SetProviderFactoryForTest(func(provider string) (geoip.GeoIPService, error) {
		return &fakeGeoIPProvider{name: provider}, nil
	})

	err := UpdateEditableOption(model.Option{
		Key:   "GeoIPProvider",
		Value: "geojs",
	})
	if err != nil {
		t.Fatalf("update editable option: %v", err)
	}

	if common.GeoIPProvider != "geojs" {
		t.Fatalf("unexpected GeoIPProvider: %s", common.GeoIPProvider)
	}
	if geoip.CurrentProvider == nil || geoip.CurrentProvider.Name() != "geojs" {
		t.Fatalf("expected active provider to be geojs")
	}
}

func TestUpdateEditableOptionRejectsInvalidServerUpdateRepo(t *testing.T) {
	setupServiceTestDB(t)

	err := UpdateEditableOption(model.Option{
		Key:   "ServerUpdateRepo",
		Value: "invalid repo",
	})
	if err == nil {
		t.Fatal("expected invalid repo format to fail")
	}
}
