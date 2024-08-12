package server

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

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yulog/go-summaly"
	"github.com/yulog/go-summaly/fetch"
)

type Server struct {
	client *fetch.Client
	once   sync.Once

	config Config

	version string
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

func New() *Server {
	var config Config
	if err := env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
	return &Server{
		config: config,
	}
}

func (srv *Server) SetVersion(version string) *Server {
	srv.version = version
	return srv
}

func (srv *Server) getClient() *fetch.Client {
	srv.once.Do(func() {
		srv.client = fetch.NewClient(fetch.ClientOpts{
			AllowPrivateIP: srv.config.AllowPrivateIP,
			Timeout:        srv.config.Timeout,
		})
	})
	return srv.client
}

func (srv *Server) getSummaly(c echo.Context) error {
	q := new(Query)
	if err := c.Bind(q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if err := c.Validate(q); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	u, err := url.Parse(q.URL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if !strings.Contains(u.Hostname(), ".") {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if pass, _ := u.User.Password(); u.User.Username() != "" || pass != "" {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	summary, err := summaly.New(
		u,
		srv.getClient(),
		summaly.WithLang(q.Lang),
		summaly.WithBotUA(srv.config.BotUA),
		summaly.WithNonBotUA(srv.config.NonBotUA),
		summaly.WithRequireNonBot(srv.config.RequireNonBotUA),
	).ResolveUserAgent().Do()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, summary)
}

func (srv *Server) Start() {
	if !srv.config.HideBanner {
		PrintBanner(srv.version)
	}
	e := echo.New()
	e.HideBanner = srv.config.HideBanner
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
		if err := e.Start(fmt.Sprintf(":%d", srv.config.Port)); err != nil && err != http.ErrServerClosed {
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
