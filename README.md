summaly
================================================================

[![Go Reference](https://pkg.go.dev/badge/github.com/yulog/go-summaly.svg)](https://pkg.go.dev/github.com/yulog/go-summaly)
[![][mit-badge]][mit]
![GitHub go.mod Go version][go-version-badge]
![GitHub Tag][tag-badge]
![GitHub Release][release-badge]

fork of [misskey-dev/summaly](https://github.com/misskey-dev/summaly)

- Go版

Installation
----------------------------------------------------------------

```
go install github.com/yulog/go-summaly/cmd/summaly@latest
```

🚧 工事中 🚧
----------------------------------------------------------------

Usage
----------------------------------------------------------------

サーバーとして:

```
summaly
```

```
http://localhost:1323/?url=https://example.com
```

コンテナとして:

```yml
services:

  summaly:
    image: ghcr.io/yulog/go-summaly:latest
    ports:
      - "8080:1323"
    environment:
      - PORT=$PORT
      - TIMEOUT=$TIMEOUT
      - REQUIRE_NON_BOT_UA_FILE=$REQUIRE_NON_BOT_UA_FILE
    volumes:
      - type: bind
        source: "./nonbot.txt"
        target: "/nonbot.txt"
```

### Options

See [environments.md](https://github.com/yulog/go-summaly/blob/go/environments.md)

#### Plugins

未対応

urls are WHATWG URL since v4.

### Returns

A Promise of an Object that contains properties below:

※ Almost all values are nullable. player should not be null.

#### Root

| Property        | Type               | Description                                 |
| :-------------- | :-------           | :------------------------------------------ |
| **title**       | *string*           | The title of the web page                   |
| **icon**        | *string*           | The url of the icon of the web page         |
| **description** | *string*           | The description of the web page             |
| **thumbnail**   | *string*           | The url of the thumbnail of the web page    |
| **player**      | *Player*           | The player of the web page                  |
| **sitename**    | *string*           | The name of the web site                    |
| **sensitive**   | *boolean*          | Whether the url is sensitive                |
| **url**         | *string*           | The url of the web page                     |

#### Player

| Property        | Type       | Description                                     |
| :-------------- | :--------- | :---------------------------------------------- |
| **url**         | *string*   | The url of the player                           |
| **width**       | *number* \| *null*   | The width of the player                         |
| **height**      | *number* \| *null*   | The height of the player                        |
| **allow**       | *string[]* | The names of the allowed permissions for iframe |

Currently the possible items in `allow` are:

* `autoplay`
* `clipboard-write`
* `fullscreen`
* `encrypted-media`
* `picture-in-picture`
* `web-share`

See [Permissions Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Permissions_Policy) in MDN for details of them.

### Example

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/yulog/go-summaly"
	"github.com/yulog/go-summaly/fetch"
)

var c = fetch.NewClient(fetch.ClientOpts{})

func main() {
	u, _ := url.Parse("https://www.youtube.com/watch?v=NMIEAhH_fTU")
	summary, _ := summaly.New(u, c).Do()

	v, _ := json.Marshal(summary)

	fmt.Println(string(v))
}
```

will be ... ↓

```json
{
	"title": "【アイドルマスター】「Stage Bye Stage」(歌：島村卯月、渋谷凛、本田未央)",
	"icon": "https://www.gstatic.com/youtube/img/web/monochrome/logo_512x512.png",
	"description": "Website▶https://columbia.jp/idolmaster/Playlist▶https://www.youtube.com/playlist?list=PL83A2998CF3BBC86D2018年7月18日発売予定THE IDOLM@STER CINDERELLA GIRLS CG STAR...",
	"thumbnail": "https://i.ytimg.com/vi/NMIEAhH_fTU/maxresdefault.jpg",
	"player": {
		"url": "https://www.youtube.com/embed/NMIEAhH_fTU?feature=oembed",
		"width": 200,
		"height": 113,
		"allow": [
			"autoplay",
			"clipboard-write",
			"encrypted-media",
			"picture-in-picture",
			"web-share",
			"fullscreen"
		]
	},
	"sitename": "YouTube",
	"sensitive": false,
	"url": "https://www.youtube.com/watch?v=NMIEAhH_fTU"
}
```

Testing
----------------------------------------------------------------

`go test`

License
----------------------------------------------------------------

[MIT](LICENSE)

[mit]:            http://opensource.org/licenses/MIT
[mit-badge]:      https://img.shields.io/badge/License-MIT-yellow.svg
[go-version-badge]:https://img.shields.io/github/go-mod/go-version/yulog/go-summaly
[tag-badge]:https://img.shields.io/github/v/tag/yulog/go-summaly
[release-badge]:https://img.shields.io/github/v/release/yulog/go-summaly