package sourcefetch

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type staticResolver struct {
	lookup func(ctx context.Context, host string) ([]net.IPAddr, error)
}

const publicResolvedIP = "93.184.216.34"

func (s staticResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if s.lookup != nil {
		return s.lookup(ctx, host)
	}
	return nil, nil
}

func newMappedHTTPClient(hostMap map[string]string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err == nil {
					if target, ok := hostMap[host]; ok {
						addr = target
					}
				}
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
		},
		Timeout: 10 * time.Second,
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchYAMLRejectsInvalidURLOrMissingHost(t *testing.T) {
	fetcher := NewFetcherWithDependencies(nil, nil, defaultMaxBodyBytes, time.Now)

	for _, rawURL := range []string{"", "://bad", "ftp://example.com/config.yaml", "http://", "https:///config.yaml"} {
		t.Run(rawURL, func(t *testing.T) {
			result, err := fetcher.FetchYAML(context.Background(), rawURL)
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
			if err == nil || err.Error() != "请输入有效的 http/https 订阅地址" {
				t.Fatalf("expected invalid URL error, got %v", err)
			}
		})
	}
}

func TestFetchYAMLReturnsDNSFailureError(t *testing.T) {
	fetcher := NewFetcherWithDependencies(
		nil,
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return nil, errors.New("lookup failed")
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "解析订阅地址失败" {
		t.Fatalf("expected DNS failure error, got %v", err)
	}
}

func TestFetchYAMLRejectsDirectPrivateOrLoopbackIP(t *testing.T) {
	fetcher := NewFetcherWithDependencies(nil, nil, defaultMaxBodyBytes, time.Now)

	for _, rawURL := range []string{"http://127.0.0.1/config.yaml", "http://10.0.0.8/config.yaml"} {
		t.Run(rawURL, func(t *testing.T) {
			result, err := fetcher.FetchYAML(context.Background(), rawURL)
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
			if err == nil || err.Error() != "该订阅地址不允许访问" {
				t.Fatalf("expected blocked-address error, got %v", err)
			}
		})
	}
}

func TestFetchYAMLRejectsDirectUnspecifiedIP(t *testing.T) {
	fetcher := NewFetcherWithDependencies(nil, nil, defaultMaxBodyBytes, time.Now)

	for _, rawURL := range []string{"http://0.0.0.0/config.yaml", "http://[::]/config.yaml"} {
		t.Run(rawURL, func(t *testing.T) {
			result, err := fetcher.FetchYAML(context.Background(), rawURL)
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
			if err == nil || err.Error() != "该订阅地址不允许访问" {
				t.Fatalf("expected blocked-address error, got %v", err)
			}
		})
	}
}

func TestFetchYAMLRejectsDNSResolvedLoopbackIP(t *testing.T) {
	fetcher := NewFetcherWithDependencies(
		nil,
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "该订阅地址不允许访问" {
		t.Fatalf("expected blocked-address error, got %v", err)
	}
}

func TestFetchYAMLRejectsDNSResolvedPrivateIP(t *testing.T) {
	fetcher := NewFetcherWithDependencies(
		nil,
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("192.168.1.8")}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "该订阅地址不允许访问" {
		t.Fatalf("expected blocked-address error, got %v", err)
	}
}

func TestFetchYAMLReturnsFetchFailureForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	fetcher := NewFetcherWithDependencies(
		newMappedHTTPClient(map[string]string{
			"sub.example.com": server.Listener.Addr().String(),
			publicResolvedIP:  server.Listener.Addr().String(),
		}),
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "拉取订阅内容失败" {
		t.Fatalf("expected fetch failure error, got %v", err)
	}
}

func TestFetchYAMLReturnsOversizeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "123456")
	}))
	defer server.Close()

	fetcher := NewFetcherWithDependencies(
		newMappedHTTPClient(map[string]string{
			"sub.example.com": server.Listener.Addr().String(),
			publicResolvedIP:  server.Listener.Addr().String(),
		}),
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		5,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "订阅内容过大" {
		t.Fatalf("expected oversize error, got %v", err)
	}
}

func TestFetchYAMLBlocksRedirectFromPublicURLToPrivateAddress(t *testing.T) {
	var privateHits atomic.Int32
	privateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		privateHits.Add(1)
		_, _ = io.WriteString(w, "proxies:\n  - name: secret\n")
	}))
	defer privateServer.Close()

	publicServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, privateServer.URL+"/config.yaml", http.StatusFound)
	}))
	defer publicServer.Close()

	fetcher := NewFetcherWithDependencies(
		newMappedHTTPClient(map[string]string{
			"sub.example.com": publicServer.Listener.Addr().String(),
			publicResolvedIP:  publicServer.Listener.Addr().String(),
		}),
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "拉取订阅内容失败" {
		t.Fatalf("expected fetch failure error, got %v", err)
	}
	if got := privateHits.Load(); got != 0 {
		t.Fatalf("expected redirect target not to be requested, got %d requests", got)
	}
}

func TestFetchYAMLDisablesProxyOnGuardedTransport(t *testing.T) {
	const directBody = "proxies:\n  - name: direct\n"
	var proxyHits atomic.Int32

	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, directBody)
	}))
	defer directServer.Close()

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHits.Add(1)
		_, _ = io.WriteString(w, "proxies:\n  - name: proxied\n")
	}))
	defer proxyServer.Close()

	proxyURL, err := url.Parse(proxyServer.URL)
	if err != nil {
		t.Fatalf("expected proxy URL to parse: %v", err)
	}

	baseClient := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return proxyURL, nil
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, splitErr := net.SplitHostPort(addr)
				if splitErr == nil {
					switch host {
					case "sub.example.com", publicResolvedIP:
						addr = directServer.Listener.Addr().String()
					}
				}
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
		},
		Timeout: 10 * time.Second,
	}

	fetcher := NewFetcherWithDependencies(
		baseClient,
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	guardedTransport, ok := fetcher.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected guarded transport to stay as *http.Transport, got %T", fetcher.client.Transport)
	}
	if guardedTransport.Proxy != nil {
		t.Fatal("expected guarded transport proxy to be disabled")
	}

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if string(result.Content) != directBody {
		t.Fatalf("expected direct body %q, got %q", directBody, string(result.Content))
	}
	if got := proxyHits.Load(); got != 0 {
		t.Fatalf("expected proxy to remain unused, got %d hits", got)
	}
}

func TestFetchYAMLPreservesDefaultRedirectCapWhenBaseClientHasNilCheckRedirect(t *testing.T) {
	var finalReached atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.URL.Query().Get("remaining")) == "0" {
			finalReached.Add(1)
			_, _ = io.WriteString(w, "proxies:\n  - name: final\n")
			return
		}

		remaining := 11
		if raw := strings.TrimSpace(r.URL.Query().Get("remaining")); raw != "" {
			var parseErr error
			remaining, parseErr = strconv.Atoi(raw)
			if parseErr != nil {
				t.Fatalf("expected remaining to be numeric, got %q: %v", raw, parseErr)
			}
		}

		http.Redirect(w, r, "http://sub.example.com/config.yaml?remaining="+strconv.Itoa(remaining-1), http.StatusFound)
	}))
	defer server.Close()

	fetcher := NewFetcherWithDependencies(
		newMappedHTTPClient(map[string]string{
			"sub.example.com": server.Listener.Addr().String(),
			publicResolvedIP:  server.Listener.Addr().String(),
		}),
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	if fetcher.client.CheckRedirect == nil {
		t.Fatal("expected guarded client to always install redirect checks")
	}

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml?remaining=11")
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if err == nil || err.Error() != "拉取订阅内容失败" {
		t.Fatalf("expected fetch failure error, got %v", err)
	}
	if got := finalReached.Load(); got != 0 {
		t.Fatalf("expected redirect cap to stop before final response, got %d final hits", got)
	}
}

func TestNewFetcherWithDependenciesForcesNonZeroTimeoutForInjectedClient(t *testing.T) {
	fetcher := NewFetcherWithDependencies(&http.Client{Timeout: 0}, nil, defaultMaxBodyBytes, time.Now)

	if fetcher.client == nil {
		t.Fatal("expected guarded client to be initialized")
	}
	if fetcher.client.Timeout == 0 {
		t.Fatal("expected guarded client to enforce a non-zero timeout")
	}
}

func TestFetchYAMLReusesGuardedTransportAcrossRequests(t *testing.T) {
	var newConnections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "proxies:\n  - name: keepalive\n")
	}))
	server.Config.ConnState = func(conn net.Conn, state http.ConnState) {
		if state == http.StateNew {
			newConnections.Add(1)
		}
	}
	server.Start()
	defer server.Close()

	fetcher := NewFetcherWithDependencies(
		newMappedHTTPClient(map[string]string{
			"sub.example.com": server.Listener.Addr().String(),
			publicResolvedIP:  server.Listener.Addr().String(),
		}),
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	for i := 0; i < 2; i++ {
		result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/config.yaml")
		if err != nil {
			t.Fatalf("expected nil error on fetch %d, got %v", i+1, err)
		}
		if result == nil {
			t.Fatalf("expected non-nil result on fetch %d", i+1)
		}
	}

	if got := newConnections.Load(); got != 1 {
		t.Fatalf("expected guarded transport to reuse one connection, got %d new connections", got)
	}
}

func TestFetchYAMLUsesInjectedNonTransportRoundTripper(t *testing.T) {
	var calls atomic.Int32
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/yaml"}},
				Body:       io.NopCloser(strings.NewReader("proxies:\n  - name: custom\n")),
				Request:    req,
			}, nil
		}),
		Timeout: 200 * time.Millisecond,
	}

	fetcher := NewFetcherWithDependencies(
		client,
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		time.Now,
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com:1/config.yaml")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected injected round tripper to be called once, got %d", got)
	}
}

func TestFetchYAMLReturnsExpectedMetadataAndHeadersOnSuccess(t *testing.T) {
	const body = "proxies:\n  - name: demo\n"
	fetchedAt := time.Unix(1700000000, 0)
	var acceptHeader string
	var userAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptHeader = r.Header.Get("Accept")
		userAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		_, _ = io.WriteString(w, body)
	}))
	defer server.Close()

	fetcher := NewFetcherWithDependencies(
		newMappedHTTPClient(map[string]string{
			"sub.example.com": server.Listener.Addr().String(),
			publicResolvedIP:  server.Listener.Addr().String(),
		}),
		staticResolver{lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP(publicResolvedIP)}}, nil
		}},
		defaultMaxBodyBytes,
		func() time.Time { return fetchedAt },
	)

	result, err := fetcher.FetchYAML(context.Background(), "http://sub.example.com/path/clash.yaml")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.DisplayName != "clash.yaml" {
		t.Fatalf("expected display name clash.yaml, got %q", result.DisplayName)
	}
	if !bytes.Equal(result.Content, []byte(body)) {
		t.Fatalf("expected body %q, got %q", body, string(result.Content))
	}
	if result.ContentType != "text/yaml; charset=utf-8" {
		t.Fatalf("expected content type text/yaml; charset=utf-8, got %q", result.ContentType)
	}
	if !result.FetchedAt.Equal(fetchedAt) {
		t.Fatalf("expected fetched at %v, got %v", fetchedAt, result.FetchedAt)
	}
	if acceptHeader != "application/x-yaml,text/yaml,text/plain,*/*" {
		t.Fatalf("expected Accept header to be set, got %q", acceptHeader)
	}
	if userAgent != "PoolX-Server" {
		t.Fatalf("expected User-Agent header PoolX-Server, got %q", userAgent)
	}
}
