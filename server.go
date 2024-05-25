package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yulog/go-summaly/fetch"
)

type Query struct {
	URL  string `query:"url" json:"url" validate:"required,http_url"`
	Lang string `query:"lang" json:"lang" validate:"omitempty,bcp47_language_tag"`
}

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func getSummaly(c echo.Context) error {
	q := new(Query)
	if err := c.Bind(q); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	if err := c.Validate(q); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	u, err := url.Parse(q.URL)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	if !strings.Contains(u.Hostname(), ".") {
		return c.String(http.StatusBadRequest, "bad request")
	}

	s := Summaly{URL: u, Lang: q.Lang}
	summary, err := s.Do()
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request "+err.Error())
	}
	return c.JSON(http.StatusOK, summary)
}

var (
	client *fetch.Client
	once   sync.Once
)

func getClient() *fetch.Client {
	once.Do(func() {
		client = fetch.NewClient(fetch.ClientOpts{
			AllowPrivateIP: config.AllowPrivateIP,
			Timeout:        60 * time.Second,
		})
	})
	return client
}

func main() {
	loadConfig()

	e := echo.New()
	e.JSONSerializer = &JSONSerializer{}
	e.Use(middleware.Logger())
	// e.Use(middleware.Gzip())
	e.Use(middleware.Recover())
	e.Validator = &Validator{validator: validator.New()}
	e.GET("/", getSummaly)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Port)))
}
