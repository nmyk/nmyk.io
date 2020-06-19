package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Signal struct {
	ChannelName string      `json:"channel_name"`
	FromUser    User        `json:"from_user"`
	Type        MessageType `json:"type"`
	Data        string      `json:"data"`
}

type MessageType int

const (
	ENTRANCE MessageType = iota
	EXIT
	NAME_CHANGE
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

func getEntranceMessage(r *Signal) []byte {
	s := Signal{
		ChannelName: r.ChannelName,
		FromUser:    User{},
		Type:        ENTRANCE,
		Data:        fmt.Sprintf("%s joined", r.FromUser.Name),
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
		signal := &Signal{}
		_ = json.Unmarshal(rawSignal, signal)
		if signal.Type == ENTRANCE {
			err = c.WriteMessage(mt, getEntranceMessage(signal))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}
