package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var secret string
var store *sessions.CookieStore

func init() {
	flag.StringVar(&secret, "s", "", "secret used to set up cookie store")
	store = sessions.NewCookieStore([]byte(secret))
}

func main() {
	flag.Parse()
	if secret == "" {
		panic("cannot run without a session secret")
	}
	e := echo.New()
	lobbyBroker := NewLobbyBroker()
	lobby := NewLobby(lobbyBroker)
	lobbyBroker.setLobby(lobby)

	/**************
	* middlewares *
	**************/

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		/*
			! This should change in production. Do not deploy unless this is properly set.
		*/
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))
	e.Use(session.Middleware(store))

	/*********
	* routes *
	*********/

	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/lobby", echo.WrapHandler(lobbyBroker), CreateSessionToken)
	e.GET("/join", func(c echo.Context) error {
		log.Println("here i am", c.Get(sessionToken))
		/*
			... TODO
		*/
		return nil
	}, AuthenticateToken)
	e.POST("/create", func(c echo.Context) error {
		type parameters struct {
			Roomname string `json:"roomname" form:"roomname" query:"roomname"`
		}
		p := new(parameters)
		if err := c.Bind(p); err != nil {
			return err
		}
		if p.Roomname == "" {
			return c.String(http.StatusBadRequest, "roomname is required")
		}
		broker := NewRoomBroker()
		room := NewRoom(p.Roomname, broker)
		lobby.Add(room)
		return c.NoContent(http.StatusCreated)
	}, AuthenticateToken)

	e.Logger.Fatal(e.Start(":8080"))
}
