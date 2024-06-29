package summaly

import (
	"fmt"
	"net/url"

	"github.com/yulog/go-summaly/fetch"
	"golang.org/x/net/html"
)

type Summaly struct {
	URL  *url.URL
	Lang string
	Body []byte
	Node *html.Node

	Client *fetch.Client
}

type Summarizer interface {
	test() bool
	summarize(*Summaly) (Summary, error)
}

func New(u *url.URL, c *fetch.Client, options ...Option) *Summaly {
	s := &Summaly{
		URL:    u,
		Client: c,
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

type Option func(*Summaly)

func WithLang(lang string) func(*Summaly) {
	return func(s *Summaly) {
		s.Lang = lang
	}
}

// TODO: これ問題ないの？
var ss = []Summarizer{new(General)}

func (s *Summaly) Do() (Summary, error) {
	var err error
	s.Node, err = s.Client.NewRequest(s.URL,
		fetch.WithAcceptLanguage(s.Lang),
	).GetHtmlNode()
	if err != nil {
		return Summary{}, err
	}

	// ss := []Summarizer{new(General)}
	for _, v := range ss {
		if v.test() {
			return v.summarize(s)
		}
	}
	return Summary{}, fmt.Errorf("failed summarize")
}

// TODO: 不要な部分はomitemptyでも良い？nullにしないとダメ？
type Summary struct {
	Title       string  `json:"title"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	Thumbnail   string  `json:"thumbnail"`
	Player      *Player `json:"player,omitempty"`
	Sitename    string  `json:"sitename"`
	Sensitive   bool    `json:"sensitive"`
	URL         string  `json:"url"`
}

// TODO: 不要な部分はomitemptyでも良い？nullにしないとダメ？
type Player struct {
	URL    string   `json:"url,omitempty"`
	Width  *any     `json:"width,omitempty"`
	Height *any     `json:"height,omitempty"`
	Allow  []string `json:"allow,omitempty"`
}
