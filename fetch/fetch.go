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

// var config *safeurl.Config

// var client *safeurl.WrappedClient

var allowType = []string{"text/html", "application/xhtml+xml"}

const limit = 10 << 20 // 10MiB

type Options struct {
	AllowType      []string
	Limit          int64
	UserAgent      string
	Accept         string
	AcceptLanguage string

	// allowPrivateIP bool
	client *Client
}

type Client struct {
	SafeClient *safeurl.WrappedClient
	TestClient *http.Client

	AllowPrivateIP bool
}

type ClientOpts struct {
	AllowPrivateIP bool
	Timeout        time.Duration
}

// TODO: optionとかNewを整理する
func NewClient(c ClientOpts) *Client {
	if c.AllowPrivateIP {
		return &Client{TestClient: http.DefaultClient, AllowPrivateIP: c.AllowPrivateIP}
	}
	config := safeurl.GetConfigBuilder().SetTimeout(c.Timeout).Build()
	return &Client{SafeClient: safeurl.Client(config), AllowPrivateIP: c.AllowPrivateIP}
}

// New は Options を返す
func New(c *Client) *Options {
	return &Options{
		AllowType: allowType,
		Limit:     limit,
		UserAgent: "SummalyBot/0.0.1",
		Accept:    "text/html, application/xhtml+xml",
		client:    c,
	}
}

// func (o *Options) AllowPrivateIP(allow bool) *Options {
// 	o.allowPrivateIP = allow
// 	return o
// }

func (o *Options) clientdo(req *http.Request) (*http.Response, error) {
	if o.client.AllowPrivateIP {
		return o.client.TestClient.Do(req)
	} else {
		return o.client.SafeClient.Do(req)
	}
}

func (o *Options) limitEncode(resp *http.Response) io.Reader {
	// Bodyサイズ制限
	// https://golang.hateblo.jp/entry/2019/10/08/215202
	// Apache-2.0 Copyright 2018 Adam Tauber
	// https://github.com/gocolly/colly/blob/master/http_backend.go#L198
	var bodyReader io.Reader = resp.Body
	bodyReader = io.LimitReader(bodyReader, o.Limit)

	// Encoding
	// https://mattn.kaoriya.net/software/lang/go/20171205164150.htm
	br := bufio.NewReader(bodyReader)
	var r io.Reader = br
	if data, err := br.Peek(4096); err == nil {
		enc, name, _ := charset.DetermineEncoding(data, resp.Header.Get("content-type"))
		if enc != nil {
			r = enc.NewDecoder().Reader(br)
		} else if name != "" {
			if enc := encoding.GetEncoding(name); enc != nil {
				r = enc.NewDecoder().Reader(br)
			}
		}
	}

	return r
}

// Do は指定の url から response を取得する
func (o *Options) do(url *url.URL) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", o.UserAgent)
	req.Header.Set("Accept", o.Accept)
	if o.AcceptLanguage != "" {
		req.Header.Set("Accept-Language", o.AcceptLanguage)
	}

	resp, err := o.clientdo(req)
	if err != nil {
		return nil, err
	}

	ct := resp.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(o.AllowType, mediatype) {
		return nil, fmt.Errorf("rejected by type: %s", mediatype)
	}

	return resp, nil
}

// Do は指定の url からBodyを取得する
func (o *Options) Do(url *url.URL) ([]byte, error) {
	resp, err := o.do(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(o.limitEncode(resp))
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetHtmlNode は指定の url から Body を取得し、 html.Node を返す
func (o *Options) GetHtmlNode(url *url.URL) (*html.Node, error) {
	resp, err := o.do(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	node, err := html.Parse(o.limitEncode(resp))
	if err != nil {
		return nil, err
	}
	return node, nil
}

// GetJSON は指定の url から Body を取得し、 out に decode する
func (o *Options) GetJSON(url *url.URL, out any) error {
	resp, err := o.do(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(o.limitEncode(resp)).Decode(out)
}
