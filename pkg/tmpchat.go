package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Message struct {
	ChannelName string `json:"channel_name"`
	FromUser    User   `json:"from_user"`
	Type        int    `json:"type"`
	Data        string `json:"data"`
}

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true // TODO: make this actually check the origin in a good way
	},
}

func signalingHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		rawMessage, err := url.QueryUnescape(string(message))
		if err != nil {
			log.Println("unescape:", err)
			break
		}
		log.Printf("recv: %s", rawMessage)
		err = c.WriteMessage(mt, []byte(rawMessage))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
