package fetch

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"

	"golang.org/x/exp/slices"
	"golang.org/x/net/html/charset"

	"github.com/doyensec/safeurl"
	"github.com/mattn/go-encoding"
)

var config = safeurl.GetConfigBuilder().Build()

var client = safeurl.Client(config)

var allowType = []string{"text/html", "application/xhtml+xml"}

const limit = 10 << 20

type Options struct {
	AllowType []string
	Limit     int
	UserAgent string
	Accept    string
}

// New は Options を返す
func New() Options {
	return Options{
		AllowType: allowType,
		Limit:     limit,
		UserAgent: "SummalyBot/0.0.1",
		Accept:    "text/html, application/xhtml+xml",
	}
}

// Do は指定の url からBodyを取得する
func (o Options) Do(url *url.URL) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", o.UserAgent)
	req.Header.Set("Accept", o.Accept)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(o.AllowType, mediatype) {
		return nil, fmt.Errorf("rejected by type: %s", mediatype)
	}

	// Bodyサイズ制限
	// https://golang.hateblo.jp/entry/2019/10/08/215202
	// Apache-2.0 Copyright 2018 Adam Tauber
	// https://github.com/gocolly/colly/blob/master/http_backend.go#L198
	var bodyReader io.Reader = resp.Body
	bodyReader = io.LimitReader(bodyReader, int64(o.Limit))

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

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return body, nil
}
