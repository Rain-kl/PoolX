package service

import (
	"net"
	"path/filepath"
	"testing"

	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"poolx/internal/pkg/utils/geoip"
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

func TestUpdateEditableOptionAppliesNodeTestDefaults(t *testing.T) {
	setupServiceTestDB(t)

	originalURL := common.NodeTestDefaultURL
	originalTimeout := common.NodeTestDefaultTimeoutMS
	t.Cleanup(func() {
		common.NodeTestDefaultURL = originalURL
		common.NodeTestDefaultTimeoutMS = originalTimeout
	})

	if err := UpdateEditableOption(model.Option{
		Key:   "NodeTestDefaultURL",
		Value: "https://example.com/generate_204",
	}); err != nil {
		t.Fatalf("update node test default url: %v", err)
	}
	if err := UpdateEditableOption(model.Option{
		Key:   "NodeTestDefaultTimeoutMS",
		Value: "12000",
	}); err != nil {
		t.Fatalf("update node test default timeout: %v", err)
	}

	if common.NodeTestDefaultURL != "https://example.com/generate_204" {
		t.Fatalf("unexpected NodeTestDefaultURL: %s", common.NodeTestDefaultURL)
	}
	if common.NodeTestDefaultTimeoutMS != 12000 {
		t.Fatalf("unexpected NodeTestDefaultTimeoutMS: %d", common.NodeTestDefaultTimeoutMS)
	}
}

func TestUpdateEditableOptionRejectsInvalidNodeTestDefaultTimeout(t *testing.T) {
	setupServiceTestDB(t)

	err := UpdateEditableOption(model.Option{
		Key:   "NodeTestDefaultTimeoutMS",
		Value: "0",
	})
	if err == nil {
		t.Fatal("expected invalid timeout to fail")
	}
}

func TestUpdateEditableOptionAppliesClashRuntimeSettings(t *testing.T) {
	setupServiceTestDB(t)

	originalAllowLAN := common.ClashAllowLAN
	originalController := common.ClashExternalController
	originalMode := common.ClashMode
	originalSecret := common.ClashSecret
	t.Cleanup(func() {
		common.ClashAllowLAN = originalAllowLAN
		common.ClashExternalController = originalController
		common.ClashMode = originalMode
		common.ClashSecret = originalSecret
	})

	for _, option := range []model.Option{
		{Key: "ClashAllowLAN", Value: "true"},
		{Key: "ClashExternalController", Value: "127.0.0.1:29090"},
		{Key: "ClashMode", Value: "global"},
		{Key: "ClashSecret", Value: "fixed-secret"},
	} {
		if err := UpdateEditableOption(option); err != nil {
			t.Fatalf("update %s: %v", option.Key, err)
		}
	}

	if !common.ClashAllowLAN {
		t.Fatal("expected ClashAllowLAN to be true")
	}
	if common.ClashExternalController != "127.0.0.1:29090" {
		t.Fatalf("unexpected ClashExternalController: %s", common.ClashExternalController)
	}
	if common.ClashMode != "global" {
		t.Fatalf("unexpected ClashMode: %s", common.ClashMode)
	}
	if common.ClashSecret != "fixed-secret" {
		t.Fatalf("unexpected ClashSecret: %s", common.ClashSecret)
	}
}

func TestUpdateEditableOptionRejectsInvalidClashExternalController(t *testing.T) {
	setupServiceTestDB(t)

	err := UpdateEditableOption(model.Option{
		Key:   "ClashExternalController",
		Value: "127.0.0.1",
	})
	if err == nil {
		t.Fatal("expected invalid ClashExternalController to fail")
	}
}
