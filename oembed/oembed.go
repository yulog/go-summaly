package oembed

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/yulog/go-summaly/fetch"
)

type Oembed struct {
	Client    *fetch.Client
	UserAgent string
}

type Response struct {
	Type    string
	Version string
	HTML    string
	Width   any
	Height  any
}

const (
	TypePhoto = "photo"
	TypeVideo = "video"
	TypeLink  = "link"
	TypeRich  = "rich"
)

var oembedAllowType = []string{"application/json"}

func (o *Oembed) Find(doc *goquery.Document) (*url.URL, error) {
	if v, ok := doc.Find("link[type='application/json+oembed']").Attr("href"); ok {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		return doc.Url.ResolveReference(u), nil
	}
	return nil, fmt.Errorf("oembed not found")
}

func (o *Oembed) Fetch(u *url.URL, out any) error {
	options := o.Client.NewRequest(u,
		fetch.WithAccept("application/json"),
		fetch.WithAllowType(oembedAllowType),
		fetch.WithLimit(500<<10), // 500KiB
		fetch.WithUserAgent(o.UserAgent),
	)

	if err := options.GetJSON(&out); err != nil {
		return err
	}

	return nil
}
