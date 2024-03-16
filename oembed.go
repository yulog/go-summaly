package main

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
	"github.com/samber/lo"
	"github.com/yulog/go-summaly/fetch"
)

func getOembed(doc *goquery.Document) ([]byte, error) {
	if v, ok := doc.Find("link[type='application/json+oembed']").Attr("href"); ok {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}

		options := fetch.New()
		options.Accept = "application/json"
		options.AllowType = []string{"application/json"}
		options.Limit = 500 << 10 // 500KiB

		body, err := options.Do(u)
		if err != nil {
			return nil, err
		}

		return body, nil
	}
	return nil, fmt.Errorf("oembed not found")
}

type OembedInfo struct {
	OK     bool
	Player Player
}

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
}

func GetOembedPlayer(doc *goquery.Document) (OembedInfo, error) {
	body, err := getOembed(doc)
	if err != nil {
		return OembedInfo{OK: false}, err
	}

	oversion, err := jsonparser.GetString(body, "version")
	if err != nil {
		return OembedInfo{OK: false}, err
	}

	otype, err := jsonparser.GetString(body, "type")
	if err != nil {
		return OembedInfo{OK: false}, err
	}

	if oversion != "1.0" || !slices.Contains([]string{"rich", "video"}, otype) {
		return OembedInfo{OK: false}, fmt.Errorf("invalid version or type")
	}

	ohtml, err := jsonparser.GetString(body, "html")
	if err != nil {
		return OembedInfo{OK: false}, err
	}
	if !strings.HasPrefix(ohtml, "<iframe") || !strings.HasSuffix(ohtml, "</iframe>") {
		return OembedInfo{OK: false}, fmt.Errorf("iframe not contain")
	}
	odoc, err := goquery.NewDocumentFromReader(strings.NewReader(ohtml))
	if err != nil {
		return OembedInfo{OK: false}, err
	}

	iframe := odoc.Find("iframe")
	if iframe.Length() != 1 {
		return OembedInfo{OK: false}, fmt.Errorf("iframe length not equals 1")
	}
	if iframe.Parents().Length() != 2 {
		return OembedInfo{OK: false}, fmt.Errorf("iframe parents length not equals 2")
	}

	src, exists := iframe.Attr("src")
	if !exists {
		return OembedInfo{OK: false}, fmt.Errorf("iframe src is not exists")
	}

	surl, err := url.Parse(src)
	if err != nil {
		return OembedInfo{OK: false}, err
	}
	if surl.Scheme != "https" {
		return OembedInfo{OK: false}, fmt.Errorf("scheme is not https")
	}

	strwidth := ""
	if v, exists := iframe.Attr("width"); exists {
		strwidth = v
	} else if v, err := jsonparser.GetString(body, "width"); err == nil {
		strwidth = v
	}
	width, err := strconv.Atoi(strwidth)
	if err != nil {
		width = 0
	}

	strheight := ""
	if v, exists := iframe.Attr("height"); exists {
		strheight = v
	} else if v, err := jsonparser.GetString(body, "height"); err == nil {
		strheight = v
	}
	height, err := strconv.Atoi(strheight)
	if err != nil {
		height = 0
	}
	if height > 1024 {
		height = 1024
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
		return OembedInfo{OK: false}, fmt.Errorf("iframe allow contains unsafe permission: %s", strings.Join(allow, ","))
	}

	return OembedInfo{
		OK: true,
		Player: Player{
			URL:    src,
			Width:  width,
			Height: height,
			Allow:  allow,
		},
	}, nil
}
