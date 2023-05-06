package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Query struct {
	URL  string `query:"url" json:"url" validate:"http_url,required"`
	Lang string `query:"lang" json:"lang"`
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
	fmt.Println(u)
	s := Summaly{URL: u}
	summary, err := s.Do()
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	return c.JSON(http.StatusOK, summary)
}

func main() {
	e := echo.New()
	e.Validator = &Validator{validator: validator.New()}
	e.GET("/", getSummaly)
	e.Logger.Fatal(e.Start(":1323"))
}
