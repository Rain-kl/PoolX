package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"poolx/internal/model"
	"poolx/internal/pkg/sourcefetch"
)

func TestParseAndStoreSourceConfigFromURLStoresSourceMetadata(t *testing.T) {
	setupServiceTestDB(t)

	originalFetcher := SourceConfigURLFetcherForTest()
	t.Cleanup(func() {
		SetSourceConfigURLFetcherForTest(originalFetcher)
	})

	fetchedAt := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	SetSourceConfigURLFetcherForTest(func(context.Context, string) (*sourcefetch.FetchResult, error) {
		return &sourcefetch.FetchResult{
			DisplayName: "clash.yaml",
			Content: []byte(`
proxies:
  - name: hk-1
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: secret
`),
			ContentType: "text/yaml",
			FetchedAt:   fetchedAt,
		}, nil
	})

	result, err := ParseAndStoreSourceConfigFromURL(context.Background(), SourceSubscriptionInput{
		URL:          "https://sub.example.com/clash.yaml",
		UploadedBy:   "root",
		UploadedByID: 1,
	})
	if err != nil {
		t.Fatalf("ParseAndStoreSourceConfigFromURL returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected parse response")
	}

	if result.SourceConfig.SourceType != model.SourceConfigSourceTypeSubscriptionURL {
		t.Fatalf("expected source type %q, got %q", model.SourceConfigSourceTypeSubscriptionURL, result.SourceConfig.SourceType)
	}
	if result.SourceConfig.SourceURL != "https://sub.example.com/clash.yaml" {
		t.Fatalf("expected source url to be stored, got %q", result.SourceConfig.SourceURL)
	}
	if result.SourceConfig.Filename != "clash.yaml" {
		t.Fatalf("expected filename clash.yaml, got %q", result.SourceConfig.Filename)
	}
	if result.SourceConfig.ContentType != "text/yaml" {
		t.Fatalf("expected content type text/yaml, got %q", result.SourceConfig.ContentType)
	}
	if result.SourceConfig.FetchedAt == nil {
		t.Fatal("expected fetched at to be set")
	}
	if result.Summary.ValidNodes != 1 {
		t.Fatalf("expected 1 valid node, got %d", result.Summary.ValidNodes)
	}

	saved, err := model.GetSourceConfigByID(result.SourceConfig.ID)
	if err != nil {
		t.Fatalf("expected saved source config row: %v", err)
	}
	if saved.SourceType != model.SourceConfigSourceTypeSubscriptionURL {
		t.Fatalf("expected persisted source type %q, got %q", model.SourceConfigSourceTypeSubscriptionURL, saved.SourceType)
	}
	if saved.SourceURL != "https://sub.example.com/clash.yaml" {
		t.Fatalf("expected persisted source url, got %q", saved.SourceURL)
	}
	if saved.ContentType != "text/yaml" {
		t.Fatalf("expected persisted content type text/yaml, got %q", saved.ContentType)
	}
	if saved.FetchedAt == nil {
		t.Fatal("expected persisted fetched at to be set")
	}
	if !saved.FetchedAt.Equal(fetchedAt) {
		t.Fatalf("expected persisted fetched at %s, got %s", fetchedAt, saved.FetchedAt)
	}
}

func TestParseAndStoreSourceConfigFromURLPropagatesFetcherError(t *testing.T) {
	setupServiceTestDB(t)

	originalFetcher := SourceConfigURLFetcherForTest()
	t.Cleanup(func() {
		SetSourceConfigURLFetcherForTest(originalFetcher)
	})

	SetSourceConfigURLFetcherForTest(func(context.Context, string) (*sourcefetch.FetchResult, error) {
		return nil, errors.New("boom")
	})

	result, err := ParseAndStoreSourceConfigFromURL(context.Background(), SourceSubscriptionInput{
		URL:          "https://sub.example.com/clash.yaml",
		UploadedBy:   "root",
		UploadedByID: 1,
	})
	if err == nil {
		t.Fatal("expected fetch error")
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestParseAndStoreSourceConfigFromURLRejectsNilFetchResult(t *testing.T) {
	setupServiceTestDB(t)

	originalFetcher := SourceConfigURLFetcherForTest()
	t.Cleanup(func() {
		SetSourceConfigURLFetcherForTest(originalFetcher)
	})

	SetSourceConfigURLFetcherForTest(func(context.Context, string) (*sourcefetch.FetchResult, error) {
		return nil, nil
	})

	result, err := ParseAndStoreSourceConfigFromURL(context.Background(), SourceSubscriptionInput{
		URL:          "https://sub.example.com/clash.yaml",
		UploadedBy:   "root",
		UploadedByID: 1,
	})
	if err == nil {
		t.Fatal("expected nil fetch result to be rejected")
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestParseAndStoreSourceConfig(t *testing.T) {
	setupServiceTestDB(t)

	result, err := ParseAndStoreSourceConfig(SourceUploadInput{
		Filename:     "upload.yaml",
		UploadedBy:   "root",
		UploadedByID: 1,
		Content: []byte(`
proxies:
  - name: sg-1
    type: ss
    server: 2.2.2.2
    port: 443
    cipher: aes-128-gcm
    password: secret
`),
	})
	if err != nil {
		t.Fatalf("ParseAndStoreSourceConfig returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected parse response")
	}

	if result.SourceConfig.SourceType != model.SourceConfigSourceTypeUpload {
		t.Fatalf("expected source type %q, got %q", model.SourceConfigSourceTypeUpload, result.SourceConfig.SourceType)
	}
	if result.Summary.TotalNodes != 1 || result.Summary.ValidNodes != 1 || result.Summary.ImportableNodes != 1 {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected one preview node, got %d", len(result.Nodes))
	}
	if result.Nodes[0].Name != "sg-1" || result.Nodes[0].Type != "ss" || result.Nodes[0].Server != "2.2.2.2" || result.Nodes[0].Port != 443 {
		t.Fatalf("unexpected preview node: %#v", result.Nodes[0])
	}
	if result.Nodes[0].DuplicateScope != "none" {
		t.Fatalf("expected duplicate scope none, got %q", result.Nodes[0].DuplicateScope)
	}

	saved, err := model.GetSourceConfigByID(result.SourceConfig.ID)
	if err != nil {
		t.Fatalf("expected saved source config row: %v", err)
	}
	if saved.SourceType != model.SourceConfigSourceTypeUpload {
		t.Fatalf("expected persisted source type %q, got %q", model.SourceConfigSourceTypeUpload, saved.SourceType)
	}
	if saved.SourceURL != "" {
		t.Fatalf("expected persisted source url to be empty for upload, got %q", saved.SourceURL)
	}
	if saved.ContentType != "" {
		t.Fatalf("expected persisted content type to be empty for upload, got %q", saved.ContentType)
	}
	if saved.FetchedAt != nil {
		t.Fatalf("expected persisted fetched at to be nil for upload, got %v", saved.FetchedAt)
	}
}
