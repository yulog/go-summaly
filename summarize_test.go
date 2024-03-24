package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func convptr[T any](i T) *any {
	var o any = i
	return &o
}

var nilany any = nil

var tmp = template.Must(template.ParseGlob("testdata/htmls/*"))

func setupServer(template, file string) (mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	ts := httptest.NewServer(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmp.ExecuteTemplate(w, template, ts.URL)
	})
	mux.HandleFunc("/oembed.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/oembed/"+file)
	})

	return mux, ts.URL, ts.Close
}

func TestSummaly_Do_oEmbed(t *testing.T) {
	tests := []struct {
		name     string
		s        *Summaly
		want     Summary
		wantErr  bool
		file     string
		template string
	}{
		// TODO: Add test cases.
		{
			name: "basic properties",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed.json",
			template: "oembed.html",
		},
		{
			name: "type: video",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed-video.json",
			template: "oembed.html",
		},
		{
			name: "max height",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  &nilany,
					Height: convptr(1024),
					Allow:  []string{},
				},
			},
			file:     "oembed-too-tall.json",
			template: "oembed.html",
		},
		{
			name: "children are ignored",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed-iframe-child.json",
			template: "oembed.html",
		},
		{
			name: "allows fullscreen",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{"fullscreen"},
				},
			},
			file:     "oembed-allow-fullscreen.json",
			template: "oembed.html",
		},
		{
			name: "allows legacy allowfullscreen",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{"fullscreen"},
				},
			},
			file:     "oembed-allow-fullscreen-legacy.json",
			template: "oembed.html",
		},
		{
			name: "allows safelisted permissions",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow: []string{
						"autoplay",
						"clipboard-write",
						"fullscreen",
						"encrypted-media",
						"picture-in-picture",
						"web-share",
					},
				},
			},
			file:     "oembed-allow-safelisted-permissions.json",
			template: "oembed.html",
		},
		{
			name: "ignores rare permissions",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{"autoplay"},
				},
			},
			file:     "oembed-ignore-rare-permissions.json",
			template: "oembed.html",
		},
		{
			name: "oEmbed with relative path",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow: []string{
						"autoplay",
						"encrypted-media",
						"fullscreen",
					},
				},
			},
			file:     "oembed.json",
			template: "oembed-relative.html",
		},
		{
			name: "oEmbed with nonexistent path",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon:        "/apple-touch-icon.png",
				Description: "nonexistent",
				Player: Player{
					URL:    "",
					Width:  &nilany,
					Height: &nilany,
					Allow: []string{
						"autoplay",
						"encrypted-media",
						"fullscreen",
					},
				},
			},
			file:     "oembed.json",
			template: "oembed-nonexistent-path.html",
		},
		{
			name: "oEmbed with wrong path",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon:        "/apple-touch-icon.png",
				Description: "wrong url",
				Player: Player{
					URL:    "",
					Width:  &nilany,
					Height: &nilany,
					Allow: []string{
						"autoplay",
						"encrypted-media",
						"fullscreen",
					},
				},
			},
			file:     "oembed.json",
			template: "oembed-wrong-path.html",
		},
		{
			name: "oEmbed with OpenGraph",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon:        "/apple-touch-icon.png",
				Description: "blobcats rule the world",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed.json",
			template: "oembed-and-og.html",
		},
		{
			name: "Invalid oEmbed with valid OpenGraph",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon:        "/apple-touch-icon.png",
				Description: "blobcats rule the world",
				Player: Player{
					URL:    "",
					Width:  &nilany,
					Height: &nilany,
					Allow: []string{
						"autoplay",
						"encrypted-media",
						"fullscreen",
					},
				},
			},
			file:     "invalid/oembed-insecure.json",
			template: "oembed-and-og.html",
		},
		{
			name: "oEmbed with og:video",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed.json",
			template: "oembed-and-og-video.html",
		},
		{
			name: "width: 100%",
			s: &Summaly{
				URL: nil,
			},
			want: Summary{
				Icon: "/apple-touch-icon.png",
				Player: Player{
					URL:    "https://example.com/",
					Width:  &nilany,
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed-percentage-width.json",
			template: "oembed.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, serverURL, teardown := setupServer(tt.template, tt.file)
			defer teardown()

			u, _ := url.Parse(serverURL)
			// テスト用サーバのURLをセット。この方法は良くないかも？
			tt.s.URL = u
			tt.want.Title = u.Host
			tt.want.Icon = u.String() + tt.want.Icon
			tt.want.Sitename = u.Host
			tt.want.URL = u.String()

			got, err := tt.s.Do()
			if (err != nil) != tt.wantErr {
				t.Errorf("Summaly.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkSummaly_Do(b *testing.B) {
	_, serverURL, teardown := setupServer("oembed.html", "oembed.json")
	defer teardown()

	u, _ := url.Parse(serverURL)
	s := Summaly{
		URL: u,
	}
	for i := 0; i < b.N; i++ {
		s.Do()
	}
}
