package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const sessionToken = "session-token"

func CreateSessionToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get(sessionToken, c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
		}
		sess.Values[sessionToken] = uuid.NewString()
		err := sess.Save(c.Request(), c.Response())
		if err != nil {
			fmt.Println(err)
			return err
		}
		return next(c)
	}
}

func AuthenticateToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(sessionToken, c)
		if err != nil {
			fmt.Println("this is far as you go", err)
			return c.NoContent(http.StatusForbidden)
		}
		if sess.IsNew {
			return c.NoContent(http.StatusUnauthorized)
		}
		id := sess.Values[sessionToken]
	  c.Set(sessionToken, id)
		return next(c)
	}
}