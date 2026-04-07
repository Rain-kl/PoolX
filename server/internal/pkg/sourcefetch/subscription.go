package sourcefetch

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	defaultMaxBodyBytes  int64 = 2 << 20
	defaultClientTimeout       = 10 * time.Second
	defaultRedirectLimit       = 10
)

var (
	errInvalidSubscriptionURL = errors.New("请输入有效的 http/https 订阅地址")
	errResolveSubscriptionURL = errors.New("解析订阅地址失败")
	errSubscriptionBlocked    = errors.New("该订阅地址不允许访问")
	errFetchSubscription      = errors.New("拉取订阅内容失败")
	errSubscriptionTooLarge   = errors.New("订阅内容过大")
)

type ipResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type FetchResult struct {
	DisplayName string
	Content     []byte
	ContentType string
	FetchedAt   time.Time
}

type Fetcher struct {
	client       *http.Client
	resolver     ipResolver
	maxBodyBytes int64
	now          func() time.Time
}

func NewFetcher() *Fetcher {
	return NewFetcherWithDependencies(
		&http.Client{Timeout: defaultClientTimeout},
		net.DefaultResolver,
		defaultMaxBodyBytes,
		time.Now,
	)
}

func NewFetcherWithDependencies(client *http.Client, resolver ipResolver, maxBodyBytes int64, now func() time.Time) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: defaultClientTimeout}
	}
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxBodyBytes
	}
	if now == nil {
		now = time.Now
	}

	fetcher := &Fetcher{
		resolver:     resolver,
		maxBodyBytes: maxBodyBytes,
		now:          now,
	}
	fetcher.client = fetcher.newGuardedClient(client)

	return fetcher
}

func (f *Fetcher) FetchYAML(ctx context.Context, rawURL string) (*FetchResult, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return nil, errInvalidSubscriptionURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errInvalidSubscriptionURL
	}
	if parsed.Hostname() == "" {
		return nil, errInvalidSubscriptionURL
	}

	if err := f.ensurePublicHost(ctx, parsed.Hostname()); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, errFetchSubscription
	}
	req.Header.Set("Accept", "application/x-yaml,text/yaml,text/plain,*/*")
	req.Header.Set("User-Agent", "PoolX-Server")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, errFetchSubscription
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, errFetchSubscription
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.maxBodyBytes+1))
	if err != nil {
		return nil, errFetchSubscription
	}
	if int64(len(body)) > f.maxBodyBytes {
		return nil, errSubscriptionTooLarge
	}

	return &FetchResult{
		DisplayName: deriveDisplayName(parsed),
		Content:     body,
		ContentType: resp.Header.Get("Content-Type"),
		FetchedAt:   f.now(),
	}, nil
}

func (f *Fetcher) ensurePublicHost(ctx context.Context, host string) error {
	host = strings.TrimSpace(host)
	if host == "" {
		return errInvalidSubscriptionURL
	}
	if strings.EqualFold(host, "localhost") {
		return errSubscriptionBlocked
	}

	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return errSubscriptionBlocked
		}
		return nil
	}

	addrs, err := f.resolver.LookupIPAddr(ctx, host)
	if err != nil || len(addrs) == 0 {
		return errResolveSubscriptionURL
	}
	for _, addr := range addrs {
		if isPrivateIP(addr.IP) {
			return errSubscriptionBlocked
		}
	}

	return nil
}

func (f *Fetcher) newGuardedClient(base *http.Client) *http.Client {
	if base == nil {
		base = &http.Client{Timeout: defaultClientTimeout}
	}

	guarded := *base
	if guarded.Timeout == 0 {
		guarded.Timeout = defaultClientTimeout
	}

	originalCheckRedirect := base.CheckRedirect
	guarded.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req == nil || req.URL == nil || req.URL.Hostname() == "" {
			return errFetchSubscription
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return errFetchSubscription
		}
		if err := f.ensurePublicHost(req.Context(), req.URL.Hostname()); err != nil {
			return err
		}
		if originalCheckRedirect != nil {
			return originalCheckRedirect(req, via)
		}
		return defaultCheckRedirect(req, via)
	}
	guarded.Transport = f.newGuardedTransport(base.Transport)

	return &guarded
}

func (f *Fetcher) newGuardedTransport(base http.RoundTripper) http.RoundTripper {
	transport, ok := cloneTransport(base)
	if !ok {
		return &guardedRoundTripper{
			base:    base,
			fetcher: f,
		}
	}

	transport.Proxy = nil
	baseDialContext := transport.DialContext
	if baseDialContext == nil {
		baseDialContext = (&net.Dialer{}).DialContext
	}

	transport.DialTLSContext = nil
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		resolvedIPs, err := f.resolvePublicIPs(ctx, host)
		if err != nil {
			return nil, err
		}

		var lastErr error
		for _, ip := range resolvedIPs {
			conn, dialErr := baseDialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		if lastErr != nil {
			return nil, lastErr
		}

		return nil, errFetchSubscription
	}

	return transport
}

func (f *Fetcher) resolvePublicIPs(ctx context.Context, host string) ([]net.IP, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, errInvalidSubscriptionURL
	}
	if strings.EqualFold(host, "localhost") {
		return nil, errSubscriptionBlocked
	}

	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return nil, errSubscriptionBlocked
		}
		return []net.IP{ip}, nil
	}

	addrs, err := f.resolver.LookupIPAddr(ctx, host)
	if err != nil || len(addrs) == 0 {
		return nil, errResolveSubscriptionURL
	}

	resolved := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if isPrivateIP(addr.IP) {
			return nil, errSubscriptionBlocked
		}
		resolved = append(resolved, addr.IP)
	}
	if len(resolved) == 0 {
		return nil, errResolveSubscriptionURL
	}

	return resolved, nil
}

type guardedRoundTripper struct {
	base    http.RoundTripper
	fetcher *Fetcher
}

func (g *guardedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil || req.URL.Hostname() == "" {
		return nil, errFetchSubscription
	}
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return nil, errInvalidSubscriptionURL
	}
	if err := g.fetcher.ensurePublicHost(req.Context(), req.URL.Hostname()); err != nil {
		return nil, err
	}

	return g.base.RoundTrip(req)
}

func cloneTransport(base http.RoundTripper) (*http.Transport, bool) {
	if base == nil {
		return http.DefaultTransport.(*http.Transport).Clone(), true
	}
	if transport, ok := base.(*http.Transport); ok && transport != nil {
		return transport.Clone(), true
	}
	return nil, false
}

func defaultCheckRedirect(_ *http.Request, via []*http.Request) error {
	if len(via) >= defaultRedirectLimit {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func deriveDisplayName(u *url.URL) string {
	if u == nil {
		return "subscription.yaml"
	}

	name := path.Base(strings.Trim(u.Path, "/"))
	if name == "." || name == "/" || name == "" {
		if host := u.Hostname(); host != "" {
			return host
		}
		return "subscription.yaml"
	}

	return name
}
