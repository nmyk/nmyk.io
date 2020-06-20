package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	ChannelName string    `json:"channel_name"`
	FromUser    *User     `json:"from_user"`
	Type        EventType `json:"type"`
	Text        string    `json:"text"`
}

type EventType int

const (
	// Event type 0 denotes a user chat message, which we never see.
	ENTRANCE EventType = iota + 1
	EXIT
	NAME_CHANGE
	CLEAR
)

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true // TODO: make this actually check the origin in a good way
	},
}

func getEntranceMessage(r *Message) []byte {
	text := fmt.Sprintf("<span class=\"%s\">%s</span> joined", r.FromUser.Id, r.FromUser.Name)
	s := Message{
		ChannelName: r.ChannelName,
		FromUser:    nil,
		Type:        ENTRANCE,
		Text:        text,
	}
	resp, _ := json.Marshal(s)
	log.Print(string(resp))
	return resp
}

func signalingHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, rawSignal, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", rawSignal)
		message := &Message{}
		_ = json.Unmarshal(rawSignal, message)
		if message.Type == ENTRANCE {
			err = c.WriteMessage(mt, getEntranceMessage(message))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}
