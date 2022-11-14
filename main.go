package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type M map[string]interface{}

var ActionIndex = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("from action index"))
}

var ActionHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("from action home"))
	},
)

var ActionAbout = echo.WrapHandler(
	http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("from action about"))
		},
	),
)

type User struct {
	Name  string `json:"name" form:"name" query:"name"`
	Email string `json:"email" form:"email" query:"email"`
}

// validation
type User2 struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=80"`
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	r := echo.New()
	r.Validator = &CustomValidator{validator: validator.New()}

	// custom error handler
	r.HTTPErrorHandler = func(err error, c echo.Context) {
		report, ok := err.(*echo.HTTPError)

		if !ok {
			report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if castedObject, ok := err.(validator.ValidationErrors); ok {
			for _, err := range castedObject {
				switch err.Tag() {
				case "required":
					report.Message = fmt.Sprintf("%s is required", err.Field())
				case "email":
					report.Message = fmt.Sprintf("%s is not valid email", err.Field())
				case "gte":
					report.Message = fmt.Sprintf("%s value must be greater than %s", err.Field(), err.Param())
				case "lte":
					report.Message = fmt.Sprintf("%s value must be lower than %s", err.Field(), err.Param())
				}

				break
			}

		}
		c.Logger().Error(report)
		c.JSON(report.Code, report)

	}

	r.GET("/index", func(c echo.Context) error {
		data := "Hello from /index"
		return c.String(http.StatusOK, data)
	})

	r.GET("/html", func(c echo.Context) error {
		data := "Hello from /html"
		return c.HTML(http.StatusOK, data)
	})

	r.GET("/redirect", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/index")
	})

	r.GET("/json", func(c echo.Context) error {
		data := M{"Message": "Hello", "Counter": 2}
		return c.JSON(http.StatusOK, data)
	})

	r.GET("/page1", func(c echo.Context) error {
		name := c.QueryParam("name")
		data := fmt.Sprintf("hello %s", name)

		return c.String(http.StatusOK, data)
	})

	r.GET("/page2/:name", func(c echo.Context) error {
		name := c.Param("name")
		data := fmt.Sprintf("Hello %s", name)

		return c.String(http.StatusOK, data)
	})

	r.GET("/page3/:name/*", func(c echo.Context) error {
		name := c.Param("name")
		msg := c.Param("*")

		data := fmt.Sprintf("Hello %s, I have a message for you: %s", name, msg)

		return c.String(http.StatusOK, data)
	})

	r.POST("/page4", func(c echo.Context) error {
		name := c.FormValue("name")
		msg := c.FormValue("message")

		data := fmt.Sprintf("Hello %s, I have a message for you: %s", name, strings.Replace(msg, "/", "", 1))

		return c.String(http.StatusOK, data)
	})

	// penggunaan wrap handler
	r.GET("/index2", echo.WrapHandler(http.HandlerFunc(ActionIndex)))
	r.GET("/home", echo.WrapHandler(ActionHome))
	r.GET("/about", ActionAbout)

	// static
	r.Static("/static", "assets")

	// parsing request payload
	r.Any("/user", func(c echo.Context) (err error) {
		u := new(User)
		if err = c.Bind(u); err != nil {
			return
		}
		return c.JSON(http.StatusOK, u)
	})
	// validator
	r.POST("/users", func(c echo.Context) (err error) {
		u := new(User2)
		if err = c.Bind(u); err != nil {
			log.Println("err bind")
			return err
		}

		if err := c.Validate(u); err != nil {
			log.Println("err validate")
			return err
		}

		return c.JSON(http.StatusOK, true)
	})

	r.Start(":9000")
}
