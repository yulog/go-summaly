package fetch

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"time"

	"golang.org/x/net/html/charset"

	"code.dny.dev/ssrf"
	"github.com/goccy/go-json"
	"github.com/mattn/go-encoding"
	"golang.org/x/net/html"
)

var defaultAllowType = []string{"text/html", "application/xhtml+xml"}

// const defaultLimit = 10 << 20 // 10MiB

type Request struct {
	url *url.URL

	allowType      []string
	limit          int64
	userAgent      string
	accept         string
	acceptLanguage string

	client *Client
}

type Option func(*Request)

type Client struct {
	HTTPClient *http.Client
}

type ClientOpts struct {
	AllowPrivateIP bool
	Timeout        time.Duration
}

// NewClient は Client を作成する
//
// プライベートIPを許可する場合は http.DefaultClient を返し、
// 許可しない場合は独自の Client を返す
func NewClient(c ClientOpts) *Client {
	if c.AllowPrivateIP {
		return &Client{HTTPClient: http.DefaultClient}
	}
	// https://budougumi0617.github.io/2021/09/13/how_to_copy_default_transport/
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil
	}
	t = t.Clone()
	t.DialContext = (&net.Dialer{
		// DefaultTransport
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// Custom
		// iana-ipv4/6-special-registry に記載のあるものを一律拒否
		// TODO: 不要なものがあるかも
		// Default:
		// https://github.com/daenney/ssrf/blob/main/ssrf_gen.go
		Control: ssrf.New(
			ssrf.WithDeniedV4Prefixes(
				// https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
				[]netip.Prefix{
					netip.MustParsePrefix("0.0.0.0/32"),         // "This host on this network" (RFC 1122, Section 3.2.1.3)
					netip.MustParsePrefix("192.0.0.0/29"),       // IPv4 Service Continuity Prefix (RFC 7335)
					netip.MustParsePrefix("192.0.0.8/32"),       // IPv4 dummy address (RFC 7600)
					netip.MustParsePrefix("192.0.0.9/32"),       // Port Control Protocol Anycast (RFC 7723)
					netip.MustParsePrefix("192.0.0.10/32"),      // Traversal Using Relays around NAT Anycast (RFC 8155)
					netip.MustParsePrefix("192.0.0.170/32"),     // NAT64/DNS64 Discovery (RFC 8880, RFC 7050, Section 2.2)
					netip.MustParsePrefix("192.0.0.171/32"),     // NAT64/DNS64 Discovery (RFC 8880, RFC 7050, Section 2.2)
					netip.MustParsePrefix("192.0.2.0/24"),       // Documentation (TEST-NET-1) (RFC 5737)
					netip.MustParsePrefix("255.255.255.255/32"), // Limited Broadcast (RFC 8190, RFC 919, Section 7)
				}...,
			),
			ssrf.WithDeniedV6Prefixes(
				//
				[]netip.Prefix{
					// https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
					netip.MustParsePrefix("::1/128"),         // Loopback Address (RFC 4291)
					netip.MustParsePrefix("::/128"),          // Unspecified Address (RFC 4291)
					netip.MustParsePrefix("::ffff:0:0/96"),   // IPv4-mapped Address (RFC 4291)
					ssrf.IPv6NAT64Prefix,                     // IPv4-IPv6 Translat. (RFC 6052)
					netip.MustParsePrefix("64:ff9b:1::/48"),  // IPv4-IPv6 Translat. (RFC 8215)
					netip.MustParsePrefix("100::/64"),        // Discard-Only Address Block (RFC 6666)
					netip.MustParsePrefix("2001::/32"),       // TEREDO (RFC4380, RFC8190)
					netip.MustParsePrefix("2001:1::1/128"),   // Port Control Protocol Anycast (RFC 7723)
					netip.MustParsePrefix("2001:1::2/128"),   // Traversal Using Relays around NAT Anycast (RFC 8155)
					netip.MustParsePrefix("2001:1::3/128"),   // DNS-SD Service Registration Protocol Anycast Address (RFC-ietf-dnssd-srp-25)
					netip.MustParsePrefix("2001:2::/48"),     // Benchmarking (RFC 5180, RFC Errata 1752)
					netip.MustParsePrefix("2001:3::/32"),     // AMT (RFC 7450)
					netip.MustParsePrefix("2001:4:112::/48"), // AS112-v6 (RFC 7535)
					netip.MustParsePrefix("2001:10::/28"),    // Deprecated (previously ORCHID) (RFC 4843)
					netip.MustParsePrefix("2001:20::/28"),    // ORCHIDv2 (RFC 7343)
					netip.MustParsePrefix("2001:30::/28"),    // Drone Remote ID Protocol Entity Tags (DETs) Prefix (RFC 9374)
					netip.MustParsePrefix("5f00::/16"),       // Segment Routing (SRv6) SIDs (RFC-ietf-6man-sids-06)
					netip.MustParsePrefix("fc00::/7"),        // Unique-Local (RFC 4193, RFC 8190)
					netip.MustParsePrefix("fe80::/10"),       // Link-Local Unicast (RFC 4291)
					// https://www.rfc-editor.org/rfc/rfc4291.html
					netip.MustParsePrefix("ff00::/8"), // Multicast
				}...,
			),
		).Safe,
	}).DialContext

	// TODO: MaxIdleConnsPerHost とか設定必要？
	return &Client{
		HTTPClient: &http.Client{
			Timeout:   c.Timeout,
			Transport: t,
		}}
}

func WithAllowType(allowType []string) func(*Request) {
	return func(r *Request) {
		r.allowType = allowType
	}
}

func WithLimit(limit int64) func(*Request) {
	return func(r *Request) {
		r.limit = limit
	}
}

func WithUserAgent(userAgent string) func(*Request) {
	return func(r *Request) {
		r.userAgent = userAgent
	}
}

func WithAccept(accept string) func(*Request) {
	return func(r *Request) {
		r.accept = accept
	}
}

func WithAcceptLanguage(acceptLanguage string) func(*Request) {
	return func(r *Request) {
		r.acceptLanguage = acceptLanguage
	}
}

// NewRequest は *Request を返す
func (c *Client) NewRequest(url *url.URL, options ...Option) *Request {
	// p.110 Go言語プログラミングエッセンス
	// Functional Options Pattern
	req := &Request{
		url:       url,
		allowType: defaultAllowType,
		limit:     10 << 20, // 10MiB
		userAgent: "SummalyBot/0.0.1",
		accept:    "text/html, application/xhtml+xml",
		client:    c,
	}
	for _, opt := range options {
		opt(req)
	}
	return req
}

func (reqs *Request) limitEncode(resp *http.Response) io.Reader {
	// Bodyサイズ制限
	// https://golang.hateblo.jp/entry/2019/10/08/215202
	// Apache-2.0 Copyright 2018 Adam Tauber
	// https://github.com/gocolly/colly/blob/master/http_backend.go#L198
	r := io.LimitReader(resp.Body, reqs.limit)

	// Encoding
	// https://mattn.kaoriya.net/software/lang/go/20171205164150.htm
	br := bufio.NewReader(r)
	if data, err := br.Peek(4096); err == nil {
		enc, name, _ := charset.DetermineEncoding(data, resp.Header.Get("content-type"))
		if enc != nil {
			return enc.NewDecoder().Reader(br)
		} else if name != "" {
			if enc := encoding.GetEncoding(name); enc != nil {
				return enc.NewDecoder().Reader(br)
			}
		}
	}

	return br
}

// Do は指定の url から response を取得する
func (reqs *Request) do() (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, reqs.url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", reqs.userAgent)
	req.Header.Set("Accept", reqs.accept)
	if reqs.acceptLanguage != "" {
		req.Header.Set("Accept-Language", reqs.acceptLanguage)
	}

	resp, err := reqs.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	ct := resp.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(reqs.allowType, mediatype) {
		return nil, fmt.Errorf("rejected by type: %s", mediatype)
	}

	return resp, nil
}

// Do は指定の url からBodyを取得する
func (reqs *Request) Do() ([]byte, error) {
	resp, err := reqs.do()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(reqs.limitEncode(resp))
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetHtmlNode は指定の url から Body を取得し、 html.Node を返す
func (reqs *Request) GetHtmlNode() (*html.Node, error) {
	resp, err := reqs.do()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	node, err := html.Parse(reqs.limitEncode(resp))
	if err != nil {
		return nil, err
	}
	return node, nil
}

// GetJSON は指定の url から Body を取得し、 out に decode する
func (reqs *Request) GetJSON(out any) error {
	resp, err := reqs.do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(reqs.limitEncode(resp)).Decode(out)
}
