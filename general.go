package main

import (
	"bytes"
	"fmt"

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

	icons, _ := favicon.FindReader(bytes.NewReader(s.Body))
	for _, i := range icons {
		fmt.Printf("%dx%d\t%s\t%s\n", i.Width, i.Height, i.FileExt, i.URL)
	}
	icon := ""
	if len(icons) > 0 {
		icon = icons[0].URL
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

	image := ""
	if len(og.Images) > 0 {
		image = og.Images[0].URL
	}
	return Summary{
		Title:       title,
		Icon:        icon,
		Description: description,
		Thumbnail:   image,
		Player:      Player{},
		Sitename:    og.SiteName,
		Sensitive:   false,
		URL:         s.URL.String(),
	}
}
