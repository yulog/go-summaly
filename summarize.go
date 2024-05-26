package main

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

var ss = []Summarizer{new(General)}

func (s *Summaly) Do() (Summary, error) {
	req := s.Client.NewRequest(s.URL,
		fetch.WithAcceptLanguage(s.Lang),
	)

	node, err := req.GetHtmlNode()
	if err != nil {
		return Summary{}, err
	}
	// fmt.Println(string(body))
	s.Node = node

	// ss := []Summarizer{new(General)}
	for _, v := range ss {
		if v.test() {
			return v.summarize(s)
		}
	}
	return Summary{}, fmt.Errorf("failed summarize")
}

type Summary struct {
	Title       string `json:"title"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Player      Player `json:"player"`
	Sitename    string `json:"sitename"`
	Sensitive   bool   `json:"sensitive"`
	URL         string `json:"url"`
}

type Player struct {
	URL    string   `json:"url"`
	Width  *any     `json:"width"`
	Height *any     `json:"height"`
	Allow  []string `json:"allow"`
}
