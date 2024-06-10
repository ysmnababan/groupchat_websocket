package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// create upgrade for upgrading to websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// create maps of client, so it can handle multiple client simultaneusly
var clients = make(map[*websocket.Conn]bool)

// create channel for broadcasting message between client
var broadcast = make(chan Message)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

// create a handler for establish a client to a websocket server
// also responsible to broadcast message to a broadcast channel
// this is implemented by continously reading data from client
func handleConnections(w http.ResponseWriter, r *http.Request) {
	//upgrade http connection to a websocket protocol
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true
	log.Println("Client: ", ws)

	// read message from client and send to channel
	for {
		var msg Message
		err := ws.ReadJSON(&msg)

		// if error while reading message from client
		if err != nil {
			log.Printf("error: %v", err)

			//delete particular client
			delete(clients, ws)
			break
		}

		// send the message to broadcast channel
		broadcast <- msg
	}
}

// create a handler for reading message from broadcast channel
// after that send that message for each client to be printed on browser
func handleMessage() {
	for {
		msg := <-broadcast

		// send for each clients
		for client := range clients {
			// write message to client baack
			err := client.WriteJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				//delte particular client
				delete(clients, client)
			}
		}
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// add endpoint for websocket
	e.GET("/ws", func(c echo.Context) error {
		handleConnections(c.Response().Writer, c.Request())
		return nil
	})

	// run go routines for handling message
	go handleMessage()

	log.Println("http server started on :1323")
	e.Logger.Fatal(e.Start(":1323"))
}
