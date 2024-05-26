package fetch

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"time"

	"golang.org/x/net/html/charset"

	"github.com/doyensec/safeurl"
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
	SafeClient *safeurl.WrappedClient
	TestClient *http.Client

	allowPrivateIP bool
}

type ClientOpts struct {
	AllowPrivateIP bool
	Timeout        time.Duration
}

// NewClient は Client を作成する
//
// プライベートIPを許可する場合は http.DefaultClient を返し、
// 許可しない場合は doyensec/safeurl の Client を返す
func NewClient(c ClientOpts) *Client {
	if c.AllowPrivateIP {
		return &Client{TestClient: http.DefaultClient, allowPrivateIP: c.AllowPrivateIP}
	}
	config := safeurl.GetConfigBuilder().SetTimeout(c.Timeout).Build()
	return &Client{SafeClient: safeurl.Client(config), allowPrivateIP: c.AllowPrivateIP}
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

func (o *Request) clientdo(req *http.Request) (*http.Response, error) {
	if o.client.allowPrivateIP {
		return o.client.TestClient.Do(req)
	} else {
		return o.client.SafeClient.Do(req)
	}
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

	resp, err := reqs.clientdo(req)
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
