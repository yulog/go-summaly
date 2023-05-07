package main

import (
	"bytes"
	"fmt"
	"html"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/dyatlov/go-opengraph/opengraph"
	"go.deanishe.net/favicon"
)

type General struct{}

func (*General) test() bool {
	return true
}

func (*General) summarize(s *Summaly) Summary {
	og := opengraph.NewOpenGraph()
	// fmt.Println(string(s.Body))
	og.ProcessHTML(bytes.NewReader(s.Body))

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(s.Body))

	title := ""
	if og.Title != "" {
		title = og.Title
	} else if v := doc.Find("meta[property='twitter:title']").AttrOr("content", ""); v != "" {
		title = v
	} else if v := doc.Find("meta[name='twitter:title']").AttrOr("content", ""); v != "" {
		title = v
	} else if v := doc.Find("title").Text(); v != "" {
		title = v
	}

	title = Clip(html.UnescapeString(title), 100)

	icons, _ := favicon.FindReader(bytes.NewReader(s.Body))
	for _, i := range icons {
		fmt.Printf("%dx%d\t%s\t%s\n", i.Width, i.Height, i.FileExt, i.URL)
	}
	icon := ""
	if len(icons) > 0 {
		icon = icons[0].URL
	}

	if icon != "" {
		u, _ := url.Parse(icon)

		icon = s.URL.ResolveReference(u).String()
	}

	description := ""
	if og.Description != "" {
		description = og.Description
	} else if v := doc.Find("meta[property='twitter:description']").AttrOr("content", ""); v != "" {
		description = v
	} else if v := doc.Find("meta[name='twitter:description']").AttrOr("content", ""); v != "" {
		description = v
	} else if v := doc.Find("meta[name='description']").AttrOr("content", ""); v != "" {
		description = v
	}

	description = Clip(html.UnescapeString(description), 300)

	if title == description {
		description = ""
	}

	image := ""
	if len(og.Images) > 0 {
		image = og.Images[0].URL
	} else if v := doc.Find("meta[property='twitter:image']").AttrOr("content", ""); v != "" {
		image = v
	} else if v := doc.Find("link[rel='image_src']").AttrOr("href", ""); v != "" {
		image = v
	} else if v := doc.Find("link[rel='apple-touch-icon']").AttrOr("href", ""); v != "" {
		image = v
	} else if v := doc.Find("link[rel='apple-touch-icon image_src']").AttrOr("href", ""); v != "" {
		image = v
	}

	if image != "" {
		u, _ := url.Parse(image)

		image = s.URL.ResolveReference(u).String()
	}

	sitename := ""
	if og.SiteName != "" {
		sitename = og.SiteName
	} else if v := doc.Find("meta[name='application-name']").AttrOr("content", ""); v != "" {
		sitename = v
	} else {
		sitename = s.URL.Host
	}

	sensitive := doc.Find(".tweet").AttrOr("data-possibly-sensitive", "") == "true"

	return Summary{
		Title:       title,
		Icon:        icon,
		Description: description,
		Thumbnail:   image,
		Player:      Player{},
		Sitename:    sitename,
		Sensitive:   sensitive,
		URL:         s.URL.String(),
	}
}
