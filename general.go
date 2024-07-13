package summaly

import (
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
	ogp := &opengraph.OpenGraph{Intent: opengraph.Intent{Strict: true}}
	err := ogp.Walk(s.Node)
	if err != nil {
		return Summary{}, err
	}

	doc := goquery.NewDocumentFromNode(s.Node)
	doc.Url = s.URL // URLをセット(oEmbedで使う)

	var m = &info{}
	m.walk(s.Node)

	title := cmp.Or(ogp.Title, m.Twitter.Title, m.Title)
	title = Clip(html.UnescapeString(title), 100)

	icons, err := favicon.New(favicon.NopSort, favicon.IgnoreWellKnown).FindGoQueryDocument(doc, s.URL.String())
	if err != nil {
		// iconが取得できなくてもエラーにしない
		fmt.Println(err)
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
		// for _, i := range icons {
		// 	fmt.Printf("%dx%d\t%s,%s\t%s\n", i.Width, i.Height, i.FileExt, i.MimeType, i.URL)
		// }
		icon = icons[0].URL
	}

	description := cmp.Or(ogp.Description, m.Twitter.Description, m.MetaInfo.Description)
	description = Clip(html.UnescapeString(description), 300)

	if title == description {
		description = ""
	}

	image := ""
	if len(ogp.Image) > 0 {
		image = ogp.Image[0].URL
	} else {
		image = cmp.Or(m.Twitter.Image, m.LinkImage.ImageSrc, m.LinkImage.AppleTouchIcon, m.LinkImage.AppleTouchIconImageSrc)
	}

	if image != "" {
		u, err := url.Parse(image)
		if err != nil {
			// url.Parseできないなら空にする
			fmt.Println(err)
			image = ""
		} else {
			image = s.URL.ResolveReference(u).String()
		}
	}

	sitename := cmp.Or(ogp.SiteName, m.MetaInfo.ApplicationName, s.URL.Host)
	sitename = html.UnescapeString(strings.TrimSpace(sitename))

	title = CleanupTitle(title, sitename)

	// 使えないらしい
	// sensitive := doc.Find(".tweet").AttrOr("data-possibly-sensitive", "") == "true"
	sensitive := cmp.Or(m.Rating.MixiContentRating == "1", m.Rating.Rating == "adult", m.Rating.Rating == "RTA-5042-1996-1400-1577-RTA")

	player, err := GetOembedPlayer(s.Client, doc, s.UserAgent)
	if err != nil {
		fmt.Println(err)
		// oEmbedを優先、ないときにはほかを使う
		player = getPlayer(m, ogp)
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

func (m *info) walk(n *xhtml.Node) {
	if n.Type == xhtml.ElementNode {
		switch n.Data {
		case "title":
			title := opengraph.TitleTag(n)
			m.Title = title.Text
		case "link":
			link := opengraph.LinkTag(n)
			switch link.Rel {
			case "image_src":
				m.LinkImage.ImageSrc = link.Href
			case "apple-touch-icon":
				m.LinkImage.AppleTouchIcon = link.Href
			case "apple-touch-icon image_src":
				m.LinkImage.AppleTouchIconImageSrc = link.Href
			}
		case "meta":
			meta := opengraph.MetaTag(n)
			prop := cmp.Or(meta.Property, meta.Name)
			if prop != "" && meta.Content != "" {
				switch prop {
				case "twitter:title":
					if m.Twitter.Title == "" {
						m.Twitter.Title = meta.Content
					}
				case "twitter:description":
					if m.Twitter.Description == "" {
						m.Twitter.Description = meta.Content
					}
				case "description":
					if m.MetaInfo.Description == "" {
						m.MetaInfo.Description = meta.Content
					}
				case "twitter:image":
					if m.Twitter.Image == "" {
						m.Twitter.Image = meta.Content
					}
				case "twitter:card":
					if m.Twitter.Card == "" {
						m.Twitter.Card = meta.Content
					}
				case "twitter:player":
					if m.Twitter.Player == "" {
						m.Twitter.Player = meta.Content
					}
				case "twitter:player:width":
					if m.Twitter.PlayerWidth == "" {
						m.Twitter.PlayerWidth = meta.Content
					}
				case "twitter:player:height":
					if m.Twitter.PlayerHeight == "" {
						m.Twitter.PlayerHeight = meta.Content
					}
				case "application-name":
					if m.MetaInfo.ApplicationName == "" {
						m.MetaInfo.ApplicationName = meta.Content
					}
				case "mixi:content-rating":
					if m.Rating.MixiContentRating == "" {
						m.Rating.MixiContentRating = meta.Content
					}
				case "rating":
					if m.Rating.Rating == "" {
						m.Rating.Rating = meta.Content
					}
				}
			}
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		m.walk(child)
	}
}

type info struct {
	Title     string
	MetaInfo  metaInfo
	Twitter   twitter
	LinkImage linkImage
	Rating    rating
}

type linkImage struct {
	ImageSrc               string
	AppleTouchIcon         string
	AppleTouchIconImageSrc string
}

type metaInfo struct {
	Description     string
	ApplicationName string
}

type twitter struct {
	Title        string
	Description  string
	Image        string
	Card         string
	Player       string
	PlayerWidth  string
	PlayerHeight string
}

type rating struct {
	MixiContentRating string
	Rating            string
}

// getPlayer は Twitter/X, OGP の *Player を返す
func getPlayer(m *info, ogp *opengraph.OpenGraph) *Player {
	var playerUrl string
	var playerWidth any
	var playerHeight any

	// Twitter/X Player
	if m.Twitter.Card != "summary_large_image" && m.Twitter.Player != "" {
		playerUrl = m.Twitter.Player
	}

	if m.Twitter.PlayerWidth != "" {
		playerWidth, _ = strconv.Atoi(m.Twitter.PlayerWidth)
	}

	if m.Twitter.PlayerHeight != "" {
		playerHeight, _ = strconv.Atoi(m.Twitter.PlayerHeight)
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

	if playerUrl == "" {
		return nil
	}

	return &Player{
		URL:    playerUrl,
		Width:  &playerWidth,
		Height: &playerHeight,
		Allow:  []string{"autoplay", "encrypted-media", "fullscreen"},
	}
}
