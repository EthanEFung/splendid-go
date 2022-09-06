package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

/*
	LobbyBroker is responsible for sending push notifications to subscribers regarding
	the current rooms that are open and players that join the current rooms.
*/
type LobbyBroker struct {
	/*
		subscribers is a map of the current user channels that will receive the messages.
	*/
	subscribers map[chan string]bool
	/*
		subscribing is the channel that will be used to add a new user channel
		to `subscribers`.
	*/
	subscribing chan chan string

	/*
		unsubscribing is the channel used to remove an existing user channel
		from `subscribers`.
	*/
	unsubscribing chan chan string
	/*
		messages is the channel that receives all the messages that should be broadcasted
		to subscribers.
	*/
	messages chan string
}

func NewLobbyBroker() *LobbyBroker {
	broker := &LobbyBroker{
		subscribers:   make(map[chan string]bool),
		subscribing:   make(chan chan string),
		unsubscribing: make(chan chan string),
		messages:      make(chan string),
	}

	go broker.listen()

	return broker
}

func (b *LobbyBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan string)

	b.subscribing <- messageChan

	ctx := r.Context()
	go func() {
		<-ctx.Done()
		b.unsubscribing <- messageChan
		log.Println("HTTP connection closed")
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	for {
		msg, open := <-messageChan
		if !open {
			// if our messageChan was closed this means the client has disconnected
			break
		}
		log.Println("flushing message", msg)
		msg = "event: message\ndata:" + msg + "\n\n"
		fmt.Fprint(w, msg)

		flusher.Flush()
	}
}

func (b *LobbyBroker) listen() {
	for {
		select {
		case s := <-b.subscribing:
			b.subscribers[s] = true
			log.Printf("Client added. %d clients registered\n", len(b.subscribers))
		case s := <-b.unsubscribing:
			delete(b.subscribers, s)
			log.Printf("Client removed. %d clients registered\n", len(b.subscribers))
		case msg := <-b.messages:
			for channel := range b.subscribers {
				channel <- msg
			}
		}
	}
}

func main() {
	e := echo.New()
	b := NewLobbyBroker()
	e.Use(middleware.CORS())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World")
	})
	e.GET("/lobby", echo.WrapHandler(b))
	go func() {
		for {
			time.Sleep(time.Second * 2)
			msg := fmt.Sprintf("the time is %v", time.Now())
			b.messages <- msg
		}
	}()
	e.Logger.Fatal(e.Start(":8080"))
}
