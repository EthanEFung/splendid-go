package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoomSubscriber struct {
	room    string
	channel msgChan
}

type RoomMessage struct {
	room string
	msg  []byte
}

type RoomBroker struct {
	lobby *Lobby

	subscribing chan RoomSubscriber

	subscribers map[string]map[msgChan]bool

	unsubscribing chan RoomSubscriber

	messages chan RoomMessage

	validator *RoomValidator
}

func NewRoomBroker(v *RoomValidator) *RoomBroker {
	broker := &RoomBroker{
		subscribing:   make(chan RoomSubscriber),
		subscribers:   make(map[string]map[msgChan]bool),
		unsubscribing: make(chan RoomSubscriber),
		messages:      make(chan RoomMessage),
		validator:     v,
	}

	go broker.listen()

	return broker
}

/*
	HandlerFunc works similarly to the lobby brokers ServeHTTP function with a few
	acceptions. First it implements the echo HandlerFunc interface and requires
	returning an error. Second, it has the added responsibility of sorting which
	room the user is in.
*/
func (b *RoomBroker) HanderFunc(c echo.Context) error {
	// make sure that the response writer can support event streams
	_, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return c.String(http.StatusUnsupportedMediaType, "streaming is unsupported")
	}

	// check to see whether or not the suggested room is a valid room id
	id := c.Param("id")
	if uuid, err := uuid.Parse(id); err != nil || !b.validator.Validate(uuid) {
		return c.NoContent(http.StatusNotFound)
	}

	// all is well

	// set up a message channel
	messageChan := make(msgChan)

	// notify the room that a new occupant is joining
	b.subscribing <- RoomSubscriber{id, messageChan}

	// set up routine to notify broker of a disconnect
	go func() {
		<-c.Request().Context().Done()
		b.unsubscribing <- RoomSubscriber{id, messageChan}
		log.Println("Room disconnect")
	}()

	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	// echo.HeaderContentEncoding <- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Transfer-Encoding
	c.Response().Header().Set("Transfer-Encoding", "chunked")

	for {
		msg, open := <-messageChan
		if !open {
			// if our messageChan was closed this means the client has disconnected
			break
		}
		b.post(c, msg)

	}
	return nil
}

func (b *RoomBroker) Add(r * Room) {
	b.subscribers[r.ID.String()] = make(map[msgChan]bool)
}

func (b *RoomBroker) Remove(r *Room) {
	delete(b.subscribers, r.ID.String())
}

func (b *RoomBroker) listen() {
	for {
		select {
		case rs := <-b.subscribing:
			log.Printf("new subscriber for room %v\n", rs.room)
			b.subscribers[rs.room][rs.channel] = true
		case rs := <-b.unsubscribing:
			log.Printf("unsubcribing from room %v\n", rs.room)
			delete(b.subscribers[rs.room], rs.channel)
		case rm := <-b.messages:
			for channel := range b.subscribers[rm.room] {
				channel <- rm.msg
			}
		}
	}

}

func (b *RoomBroker) post(c echo.Context, msg []byte) {
	msg = append([]byte("event: message\ndata:"), msg...)
	msg = append(msg, '\n', '\n')
	c.Response().Write(msg)
	c.Response().Flush()
}

func (b *RoomBroker) setLobby(l *Lobby) {
	b.lobby = l
}