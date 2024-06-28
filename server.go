package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

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
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	if err := c.Validate(q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	u, err := url.Parse(q.URL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	if !strings.Contains(u.Hostname(), ".") {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	if pass, _ := u.User.Password(); u.User.Username() != "" || pass != "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	summary, err := New(u, srv.getClient(), WithLang(q.Lang)).Do()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request "+err.Error())
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

	// https://echo.labstack.com/docs/cookbook/graceful-shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", config.Port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Graceful Shutdown
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
