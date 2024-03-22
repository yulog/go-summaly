package main

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"
	"github.com/yulog/go-summaly/fetch"
)

func getOembed(doc *goquery.Document) (*OembedJSON, error) {
	if v, ok := doc.Find("link[type='application/json+oembed']").Attr("href"); ok {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}

		options := fetch.New()
		options.Accept = "application/json"
		options.AllowType = []string{"application/json"}
		options.Limit = 500 << 10 // 500KiB

		// body, err := options.Do(u)
		var o OembedJSON
		if err = options.GetJSON(u, &o); err != nil {
			return nil, err
		}

		return &o, nil
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

type OembedJSON struct {
	Version string
	Type    string
	HTML    string
	Width   any
	Height  any
}

func GetOembedPlayer(doc *goquery.Document) (OembedInfo, error) {
	o, err := getOembed(doc)
	if err != nil {
		return OembedInfo{OK: false}, err
	}

	// var o OembedJSON
	// err = json.NewDecoder(bytes.NewReader(body)).Decode(&o)
	// if err != nil {
	// 	return OembedInfo{OK: false}, err
	// }
	// fmt.Printf("%#v\n", o)

	// oversion, err := jsonparser.GetString(body, "version")
	// if err != nil {
	// 	return OembedInfo{OK: false}, err
	// }

	// otype, err := jsonparser.GetString(body, "type")
	// if err != nil {
	// 	return OembedInfo{OK: false}, err
	// }

	if o.Version != "1.0" || !slices.Contains([]string{"rich", "video"}, o.Type) {
		return OembedInfo{OK: false}, fmt.Errorf("invalid version or type")
	}

	// ohtml, err := jsonparser.GetString(body, "html")
	// if err != nil {
	// 	return OembedInfo{OK: false}, err
	// }
	// adventar.org でhtmlの終端に\nが入っている
	// if !strings.HasPrefix(ohtml, "<iframe") || !strings.HasSuffix(ohtml, "</iframe>") {
	// 	return OembedInfo{OK: false}, fmt.Errorf("iframe not contain")
	// }
	if !strings.Contains(o.HTML, "<iframe") {
		return OembedInfo{OK: false}, fmt.Errorf("iframe not contain")
	}
	odoc, err := goquery.NewDocumentFromReader(strings.NewReader(o.HTML))
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

	// strwidth := ""
	var width any
	if v, exists := iframe.Attr("width"); exists {
		width, err = strconv.Atoi(v)
		if err != nil {
			width = nil
		}
		// } else if v, err := jsonparser.GetString(body, "width"); err == nil {
		// 	strwidth = v
		// }
	} else if v, ok := o.Width.(int); ok {
		// strwidth = v
		width = v
	} else {
		width = nil
	}
	// width, err = strconv.Atoi(strwidth)
	// if err != nil {
	// 	width = nil
	// }

	// strheight := ""
	var height any
	if v, exists := iframe.Attr("height"); exists {
		// strheight = v
		height, err = strconv.Atoi(v)
		if err != nil {
			height = nil
		}
		// } else if v, err := jsonparser.GetString(body, "height"); err == nil {
		// 	strheight = v
		// }
	} else if v, ok := o.Height.(int); ok {
		// strheight = v
		height = v
	} else {
		height = nil
	}
	// height, err = strconv.Atoi(strheight)
	// if err != nil {
	// 	height = nil
	// }
	if height != nil && height.(int) > 1024 {
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
			Width:  &width,
			Height: &height,
			Allow:  allow,
		},
	}, nil
}
