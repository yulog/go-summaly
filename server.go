package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yulog/go-summaly/fetch"
)

type Server struct {
	client *fetch.Client
	once   sync.Once
}

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

func NewServer() *Server {
	return &Server{}
}

func (srv *Server) getClient() *fetch.Client {
	srv.once.Do(func() {
		srv.client = fetch.NewClient(fetch.ClientOpts{
			AllowPrivateIP: config.AllowPrivateIP,
			Timeout:        config.Timeout,
		})
	})
	return srv.client
}

func (srv *Server) getSummaly(c echo.Context) error {
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
	if pass, _ := u.User.Password(); u.User.Username() != "" || pass != "" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	s := Summaly{URL: u, Lang: q.Lang, Client: srv.getClient()}
	summary, err := s.Do()
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request "+err.Error())
	}
	return c.JSON(http.StatusOK, summary)
}

func main() {
	loadConfig()
	srv := NewServer()

	e := echo.New()
	e.JSONSerializer = &JSONSerializer{}
	e.Use(middleware.Logger())
	// e.Use(middleware.Gzip())
	e.Use(middleware.Recover())
	e.Validator = &Validator{validator: validator.New()}
	e.GET("/", srv.getSummaly)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Port)))
}
