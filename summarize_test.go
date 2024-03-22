package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSummaly_Do(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmp := template.Must(template.ParseFiles("testdata/htmls/oembed.html"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmp.Execute(w, ts.URL)
		// http.ServeFile(w, r, "testdata/htmls/oembed.html")
	})
	mux.HandleFunc("/oembed.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/oembed/oembed.json")
	})
	u, _ := url.Parse(ts.URL)
	tests := []struct {
		name    string
		s       *Summaly
		want    Summary
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			s: &Summaly{
				URL: u,
			},
			want: Summary{
				Player: Player{
					URL: "https://example.com/",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Do()
			if (err != nil) != tt.wantErr {
				t.Errorf("Summaly.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Player.URL != tt.want.Player.URL {
				t.Errorf("Summaly.Do() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSummaly_Do(b *testing.B) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmp := template.Must(template.ParseFiles("testdata/htmls/oembed.html"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmp.Execute(w, ts.URL)
		// http.ServeFile(w, r, "testdata/htmls/oembed.html")
	})
	mux.HandleFunc("/oembed.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/oembed/oembed.json")
	})
	u, _ := url.Parse(ts.URL)
	s := Summaly{
		URL: u,
	}
	for i := 0; i < b.N; i++ {
		s.Do()
	}
}
