package oembed

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/yulog/go-summaly/fetch"
)

type Client struct {
	Client    *fetch.Client
	UserAgent string
}

type Oembed struct {
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

func (c *Client) Find(doc *goquery.Document) (*url.URL, error) {
	if v, ok := doc.Find("link[type='application/json+oembed']").Attr("href"); ok {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		return doc.Url.ResolveReference(u), nil
	}
	return nil, fmt.Errorf("oembed not found")
}

func (c *Client) Fetch(u *url.URL, out any) error {
	options := c.Client.NewRequest(u,
		fetch.WithAccept("application/json"),
		fetch.WithAllowType(oembedAllowType),
		fetch.WithLimit(500<<10), // 500KiB
		fetch.WithUserAgent(c.UserAgent),
	)

	if err := options.GetJSON(&out); err != nil {
		return err
	}

	return nil
}

func (o *Oembed) Validate() error {
	if !(o.Type == TypePhoto || o.Type == TypeVideo || o.Type == TypeLink || o.Type == TypeRich) {
		return fmt.Errorf("invalid type: %s", o.Type)
	}
	if o.Version != "1.0" {
		return fmt.Errorf("invalid version: %s", o.Version)
	}

	return nil
}
