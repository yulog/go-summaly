package main

import (
	"bytes"
	"cmp"
	"fmt"
	"html"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/otiai10/opengraph/v2"
	"github.com/yulog/go-favicon"
	xhtml "golang.org/x/net/html"
)

type General struct{}

func (*General) test() bool {
	return true
}

func (*General) summarize(s *Summaly) (Summary, error) {
	node, err := xhtml.Parse(bytes.NewReader(s.Body))
	if err != nil {
		return Summary{}, err
	}

	ogp := &opengraph.OpenGraph{}
	err = ogp.Walk(node)
	if err != nil {
		return Summary{}, err
	}

	doc := goquery.NewDocumentFromNode(node)

	title := ""
	if ogp.Title != "" {
		title = ogp.Title
	} else if v := doc.Find("meta[property='twitter:title']").AttrOr("content", ""); v != "" {
		title = v
	} else if v := doc.Find("meta[name='twitter:title']").AttrOr("content", ""); v != "" {
		title = v
	} else if v := doc.Find("title").Text(); v != "" {
		title = v
	}

	title = Clip(html.UnescapeString(title), 100)

	icons, err := favicon.New(favicon.NopSort).FindGoQueryDocument(doc, s.URL.String())
	if err != nil {
		return Summary{}, err
	}
	// for _, i := range icons {
	// 	fmt.Printf("%dx%d\t%s\t%s\n", i.Width, i.Height, i.FileExt, i.URL)
	// }
	icon := ""
	if len(icons) > 0 {
		// sort.Slice(icons, func(i, j int) bool {
		// 	a, b := icons[i], icons[j]
		// 	switch {
		// 	case formatRank[a.MimeType] > formatRank[b.MimeType]:
		// 		return true
		// 	case formatRank[a.MimeType] < formatRank[b.MimeType]:
		// 		return false
		// 	default:
		// 		return a.Width > b.Width
		// 	}
		// })

		// cmp.Compare(a, b) -> asc
		// cmp.Compare(b, a) -> desc
		slices.SortFunc(icons, func(a, b *favicon.Icon) int {
			if n := cmp.Compare(formatRank[b.MimeType], formatRank[a.MimeType]); n != 0 {
				return n
			}
			return cmp.Compare(b.Width, a.Width)
		})
		for _, i := range icons {
			fmt.Printf("%dx%d\t%s,%s\t%s\n", i.Width, i.Height, i.FileExt, i.MimeType, i.URL)
		}
		icon = icons[0].URL
	}

	if icon != "" {
		u, err := url.Parse(icon)
		if err != nil {
			return Summary{}, err
		}
		icon = s.URL.ResolveReference(u).String()
	}

	description := ""
	if ogp.Description != "" {
		description = ogp.Description
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
	if len(ogp.Image) > 0 {
		image = ogp.Image[0].URL
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
		u, err := url.Parse(image)
		if err != nil {
			return Summary{}, err
		}
		image = s.URL.ResolveReference(u).String()
	}

	// Twitter/X Player
	tc := doc.Find("meta[property='twitter:card']").AttrOr("content", "")

	playerUrl := ""
	if v := doc.Find("meta[property='twitter:player']").AttrOr("content", ""); tc != "summary_large_image" && v != "" {
		playerUrl = v
	} else if v := doc.Find("meta[name='twitter:player']").AttrOr("content", ""); tc != "summary_large_image" && v != "" {
		playerUrl = v
	}

	playerWidth := 0
	if v := doc.Find("meta[property='twitter:player:width']").AttrOr("content", ""); v != "" {
		playerWidth, _ = strconv.Atoi(v)
	} else if v := doc.Find("meta[name='twitter:player:width']").AttrOr("content", ""); v != "" {
		playerWidth, _ = strconv.Atoi(v)
	}

	playerHeight := 0
	if v := doc.Find("meta[property='twitter:player:height']").AttrOr("content", ""); v != "" {
		playerHeight, _ = strconv.Atoi(v)
	} else if v := doc.Find("meta[name='twitter:player:height']").AttrOr("content", ""); v != "" {
		playerHeight, _ = strconv.Atoi(v)
	}

	// OGP Player
	if playerUrl == "" {
		for _, v := range ogp.Video {
			if v.URL != "" {
				playerUrl = v.URL
			} else if v.SecureURL != "" {
				playerUrl = v.SecureURL
			}
			if playerUrl != "" {
				playerWidth = v.Width
				playerHeight = v.Height
				break
			}
		}
	}

	sitename := ""
	if ogp.SiteName != "" {
		sitename = ogp.SiteName
	} else if v := doc.Find("meta[name='application-name']").AttrOr("content", ""); v != "" {
		sitename = v
	} else {
		sitename = s.URL.Host
	}

	sitename = html.UnescapeString(strings.TrimSpace(sitename))

	title = CleanupTitle(title, sitename)

	if title == "" {
		title = sitename
	}

	sensitive := doc.Find(".tweet").AttrOr("data-possibly-sensitive", "") == "true"

	info, err := GetOembedPlayer(doc)
	if err != nil {
		fmt.Println(err)
	}
	player := Player{}
	if info.OK {
		player = info.Player
	} else {
		player = Player{
			URL:    playerUrl,
			Width:  playerWidth,
			Height: playerHeight,
			Allow:  []string{"autoplay", "encrypted-media", "fullscreen"},
		}
	}

	return Summary{
		Title:       title,
		Icon:        icon,
		Description: description,
		Thumbnail:   image,
		Player:      player,
		Sitename:    sitename,
		Sensitive:   sensitive,
		URL:         s.URL.String(),
	}, nil
}
