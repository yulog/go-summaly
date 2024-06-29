package summaly

import (
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/yulog/go-summaly/fetch"
)

func convptr[T any](i T) *any {
	var o any = i
	return &o
}

var nilany any = nil

var emptyPlayer = Player{
	URL:    "",
	Width:  nil,
	Height: nil,
	Allow:  nil,
}

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

func testClient(allowPrivateIP bool) *fetch.Client {
	return fetch.NewClient(fetch.ClientOpts{
		AllowPrivateIP: allowPrivateIP,
		Timeout:        60 * time.Second,
	})
}

func TestSummaly_Do_NoFavicon(t *testing.T) {
	client := testClient(true)

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
			name: "title cleanup",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:  "Strawberry Pasta",
				Icon:   "",
				Player: nil,
			},
			file:     "oembed.json",
			template: "no-favicon.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, serverURL, teardown := setupServer(tt.template, tt.file)
			defer teardown()

			u, _ := url.Parse(serverURL)
			// テスト用サーバのURLをセット。この方法は良くないかも？
			tt.s.URL = u
			tt.want.URL = u.String()
			tt.want.Sitename = u.Host

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

func TestSummaly_Do_TitleCleanup(t *testing.T) {
	client := testClient(true)

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
			name: "title cleanup",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:    "Strawberry Pasta",
				Player:   nil,
				Sitename: "Alice's Site",
			},
			file:     "oembed.json",
			template: "dirty-title.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, serverURL, teardown := setupServer(tt.template, tt.file)
			defer teardown()

			u, _ := url.Parse(serverURL)
			// テスト用サーバのURLをセット。この方法は良くないかも？
			tt.s.URL = u
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

func TestSummaly_Do_PrivateIPBlocking(t *testing.T) {
	client := testClient(false)

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
			name: "private ip blocking",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want:     Summary{},
			wantErr:  true,
			file:     "oembed.json",
			template: "og-title.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, serverURL, teardown := setupServer(tt.template, tt.file)
			defer teardown()

			u, _ := url.Parse(serverURL)
			// テスト用サーバのURLをセット。この方法は良くないかも？
			tt.s.URL = u
			// tt.want.URL = u.String()

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

func TestSummaly_Do_OGP(t *testing.T) {
	client := testClient(true)

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
			name: "title",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:    "Strawberry Pasta",
				Player:   nil,
				Sitename: "WANT_URL",
			},
			file:     "oembed.json",
			template: "og-title.html",
		},
		{
			name: "description",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:       "YEE HAW",
				Description: "Strawberry Pasta",
				Player:      nil,
				Sitename:    "WANT_URL",
			},
			file:     "oembed.json",
			template: "og-description.html",
		},
		{
			name: "site_name",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:    "YEE HAW",
				Player:   nil,
				Sitename: "Strawberry Pasta",
			},
			file:     "oembed.json",
			template: "og-site_name.html",
		},
		{
			name: "thumbnail",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:     "YEE HAW",
				Icon:      "https://himasaku.net/himasaku.png",
				Thumbnail: "https://himasaku.net/himasaku.png",
				Player:    nil,
				Sitename:  "WANT_URL",
			},
			file:     "oembed.json",
			template: "og-image.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, serverURL, teardown := setupServer(tt.template, tt.file)
			defer teardown()

			u, _ := url.Parse(serverURL)
			// テスト用サーバのURLをセット。この方法は良くないかも？
			tt.s.URL = u
			if tt.want.Sitename == "WANT_URL" {
				tt.want.Sitename = u.Host
			}
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

func TestSummaly_Do_TwitterCard(t *testing.T) {
	client := testClient(true)

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
			name: "title",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:  "Strawberry Pasta",
				Player: nil,
			},
			file:     "oembed.json",
			template: "twitter-title.html",
		},
		{
			name: "description",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:       "YEE HAW",
				Description: "Strawberry Pasta",
				Player:      nil,
			},
			file:     "oembed.json",
			template: "twitter-description.html",
		},
		{
			name: "thumbnail",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:     "YEE HAW",
				Icon:      "https://himasaku.net/himasaku.png",
				Thumbnail: "https://himasaku.net/himasaku.png",
				Player:    nil,
			},
			file:     "oembed.json",
			template: "twitter-image.html",
		},
		{
			name: "Player detection - PeerTube:video => video",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:       "Title",
				Description: "Desc",
				Thumbnail:   "https://example.com/imageurl",
				Player: &Player{
					URL:    "https://example.com/embedurl",
					Width:  convptr(int(640)),
					Height: convptr(int(480)),
					Allow:  []string{"autoplay", "encrypted-media", "fullscreen"},
				},
				Sitename: "Site",
			},
			file:     "oembed.json",
			template: "player-peertube-video.html",
		},
		{
			name: "Player detection - Pleroma:video => video",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:       "Title",
				Description: "Desc",
				Thumbnail:   "https://example.com/imageurl",
				Player: &Player{
					URL:    "https://example.com/embedurl",
					Width:  convptr(int(480)),
					Height: convptr(int(480)),
					Allow:  []string{"autoplay", "encrypted-media", "fullscreen"},
				},
			},
			file:     "oembed.json",
			template: "player-pleroma-video.html",
		},
		{
			name: "Player detection - Pleroma:image => image",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Title:       "Title",
				Description: "Desc",
				Thumbnail:   "https://example.com/imageurl",
				Player:      nil,
			},
			file:     "oembed.json",
			template: "player-pleroma-image.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, serverURL, teardown := setupServer(tt.template, tt.file)
			defer teardown()

			u, _ := url.Parse(serverURL)
			// テスト用サーバのURLをセット。この方法は良くないかも？
			tt.s.URL = u
			if tt.want.Sitename == "" {
				tt.want.Sitename = u.Host
			}
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

func TestSummaly_Do_oEmbed(t *testing.T) {
	client := testClient(true)

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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
					URL:    "https://example.com/",
					Width:  convptr(float64(500)),
					Height: convptr(float64(300)),
					Allow:  []string{},
				},
			},
			file:     "oembed.json",
			template: "oembed-relative.html",
		},
		{
			name: "oEmbed with nonexistent path",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Description: "nonexistent",
				Player:      nil,
			},
			file:     "oembed.json",
			template: "oembed-nonexistent-path.html",
		},
		{
			name: "oEmbed with wrong path",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Description: "wrong url",
				Player:      nil,
			},
			file:     "oembed.json",
			template: "oembed-wrong-path.html",
		},
		{
			name: "oEmbed with OpenGraph",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Description: "blobcats rule the world",
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Description: "blobcats rule the world",
				Player:      nil,
			},
			file:     "invalid/oembed-insecure.json",
			template: "oembed-and-og.html",
		},
		{
			name: "oEmbed with og:video",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
				URL:    nil,
				Client: client,
			},
			want: Summary{
				Player: &Player{
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
			if tt.want.Icon != "" {
				tt.want.Icon = u.String() + tt.want.Icon
			}
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

func TestSummaly_Do_oEmbedInvalid(t *testing.T) {
	client := testClient(true)

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
			name: "oEmbed invalidity test:",
			s: &Summaly{
				URL:    nil,
				Client: client,
			},
			want: Summary{
				// Icon: "/apple-touch-icon.png",
				Player: nil,
			},
			file:     "dummy",
			template: "oembed.html",
		},
	}
	paths, err := fs.Glob(os.DirFS("testdata/oembed/invalid"), "*.json")
	if err != nil {
		return
	}
	for _, path := range paths {
		for _, tt := range tests {
			t.Run(tt.name+path, func(t *testing.T) {
				_, serverURL, teardown := setupServer(tt.template, "invalid/"+path)
				defer teardown()

				u, _ := url.Parse(serverURL)
				// テスト用サーバのURLをセット。この方法は良くないかも？
				tt.s.URL = u
				tt.want.Title = u.Host
				if tt.want.Icon != "" {
					tt.want.Icon = u.String() + tt.want.Icon
				}
				tt.want.Sitename = u.Host
				tt.want.URL = u.String()

				got, err := tt.s.Do()
				if (err != nil) != tt.wantErr {
					t.Errorf("Summaly.Do() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := cmp.Diff(tt.want.Player, got.Player); diff != "" {
					t.Errorf("(-want +got):\n%s", diff)
				}
			})
		}
	}
}

func BenchmarkSummaly_Do(b *testing.B) {
	client := testClient(true)

	_, serverURL, teardown := setupServer("oembed.html", "oembed.json")
	defer teardown()

	u, _ := url.Parse(serverURL)
	s := Summaly{
		URL:    u,
		Client: client,
	}
	for i := 0; i < b.N; i++ {
		s.Do()
	}
}
