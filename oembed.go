package summaly

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"
	"github.com/yulog/go-summaly/fetch"
	"github.com/yulog/go-summaly/oembed"
)

var safeList = []string{
	"autoplay",
	"clipboard-write",
	"fullscreen",
	"encrypted-media",
	"picture-in-picture",
	"web-share",
}

var ignoredList = []string{
	"gyroscope",
	"accelerometer",
	"", // 空の値も除去する
}

func GetOembedPlayer(client *fetch.Client, doc *goquery.Document, ua string) (*Player, error) {
	oe := &oembed.Oembed{Client: client, UserAgent: ua}
	u, err := oe.Find(doc)
	if err != nil {
		return nil, err
	}
	var o oembed.Response
	err = oe.Fetch(u, &o)
	if err != nil {
		return nil, err
	}

	if o.Version != "1.0" || !slices.Contains([]string{oembed.TypeRich, oembed.TypeVideo}, o.Type) {
		return nil, fmt.Errorf("invalid version or type")
	}

	// adventar.org でhtmlの終端に\nが入っている
	// <iframe が含まれないことだけチェックする
	// if !strings.HasPrefix(ohtml, "<iframe") || !strings.HasSuffix(ohtml, "</iframe>") {
	// 	return OembedInfo{OK: false}, fmt.Errorf("iframe not contain")
	// }
	if !strings.Contains(o.HTML, "<iframe") {
		return nil, fmt.Errorf("iframe not contain")
	}
	odoc, err := goquery.NewDocumentFromReader(strings.NewReader(o.HTML))
	if err != nil {
		return nil, err
	}

	iframe := odoc.Find("iframe")
	if iframe.Length() != 1 {
		return nil, fmt.Errorf("iframe length not equals 1")
	}
	if iframe.Parents().Length() != 2 {
		return nil, fmt.Errorf("iframe parents length not equals 2")
	}

	src, exists := iframe.Attr("src")
	if !exists {
		return nil, fmt.Errorf("iframe src is not exists")
	}

	surl, err := url.Parse(src)
	if err != nil {
		return nil, err
	}
	if surl.Scheme != "https" {
		return nil, fmt.Errorf("scheme is not https")
	}

	var width any
	if v, exists := iframe.Attr("width"); exists {
		width, err = strconv.Atoi(v)
		if err != nil {
			width = nil
		}
	} else if v, ok := o.Width.(int); ok {
		width = v
	} else if v, ok := o.Width.(float64); ok {
		width = v
	} else {
		width = nil
	}

	var height any
	if v, exists := iframe.Attr("height"); exists {
		height, err = strconv.Atoi(v)
		if err != nil {
			height = nil
		}
	} else if v, ok := o.Height.(int); ok {
		height = v
	} else if v, ok := o.Height.(float64); ok {
		height = v
	} else {
		return nil, fmt.Errorf("height is incorrect")
	}
	if height != nil {
		if i, ok := height.(int); ok && i > 1024 {
			height = 1024
		} else if i, ok := height.(float64); ok && i > 1024 {
			height = 1024
		}
	}

	allow := strings.Split(iframe.AttrOr("allow", ""), ";")

	for i, v := range allow {
		allow[i] = strings.TrimSpace(v)
	}

	allow = lo.Filter(allow, func(x string, index int) bool {
		return !slices.Contains(ignoredList, x)
	})

	if v, exists := iframe.Attr("allowfullscreen"); exists && v == "" {
		allow = append(allow, "fullscreen")
	}

	if lo.SomeBy(allow, func(x string) bool {
		return !slices.Contains(safeList, x)
	}) {
		return nil, fmt.Errorf("iframe allow contains unsafe permission: %s", strings.Join(allow, ","))
	}

	return &Player{
		URL:    src,
		Width:  &width,
		Height: &height,
		Allow:  allow,
	}, nil
}
