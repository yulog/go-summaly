package main

import (
	"net/url"
)

type Summaly struct {
	URL  *url.URL
	Lang string
	Body []byte
}

type Summarizer interface {
	test() bool
	summarize(*Summaly) (Summary, error)
}

// var ss = []Summarizer{new(General)}

func (s *Summaly) Do() (Summary, error) {
	body, err := fetch(s.URL)
	if err != nil {
		// fmt.Println(err)
		return Summary{}, err
	}
	// fmt.Println(string(body))
	s.Body = body
	// ss = []Summarizer{&OGP{URL: s.URL, Lang: s.Lang, Body: body}}
	// o,_:=interface{}(s).(OGP)
	// ss = []Summarizer{&o}
	ss := []Summarizer{new(General)}
	for _, v := range ss {
		if v.test() {
			// fmt.Println(v.summarize(s))
			return v.summarize(s)
		}
	}
	return Summary{}, nil
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
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Allow  []string `json:"allow"`
}
