package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var Tmpchat = map[string]*Channel{}

type Channel struct {
	Connections []*Connection
	Messages    chan *Message
}

type Connection struct {
	*websocket.Conn
	User *User
}

type Message struct {
	wsMsgType   int       // Websocket message type as defined in RFC 6455, section 11.8
	ChannelName string    `json:"channel_name"`
	FromUser    *User     `json:"from_user"`
	Type        EventType `json:"type"` // Tmpchat-specific signaling event type
	Text        string    `json:"text"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EventType int

const (
	// Event type 0 denotes a user chat message, which we never see here.
	ENTRANCE EventType = iota + 1
	EXIT
	NAME_CHANGE
	CLEAR
)

func getEntranceMessage(i *Message) []byte {
	text := fmt.Sprintf("<span class=\"%s\">%s</span> joined", i.FromUser.ID, i.FromUser.Name)
	o := Message{
		ChannelName: i.ChannelName,
		FromUser:    nil,
		Type:        ENTRANCE,
		Text:        text,
	}
	resp, _ := json.Marshal(o)
	return resp
}

func signalingHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool {
			return true // TODO: make this actually check the origin in a good way
		},
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	channelName := r.URL.Path[1:]
	// Unmarshal all my messages and send them to the channel
	for {
		mt, rawSignal, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", rawSignal)
		message := &Message{}
		if err := json.Unmarshal(rawSignal, message); err != nil {
			continue
		}
		message.wsMsgType = mt
		Tmpchat[channelName].Messages <- message
	}
}

type TmpchatData struct {
	BgData
	ChannelName string
	User        User
	AppHost     string // So this works seamlessly in dev (localhost) and prod (tmpch.at)
}

func tmpchatHandler(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Host)
	channelName := r.URL.Path[1:] // Omit leading slash in path
	var tmpl *template.Template
	var user User
	if channelName == "" {
		tmpl = getTemplate("tmpchat-index")
	} else {
		user = User{
			uuid.New().String(),
			"anon-00",
		}
		if _, ok := Tmpchat[channelName]; !ok {
			conn := &Connection{nil, &user}
			messages := make(chan *Message)
			Tmpchat[channelName] = &Channel{
				Connections: []*Connection{conn},
				Messages:    messages,
			}
			go func() { // Receive messages from channel and propagate to all users
				for m := range Tmpchat[channelName].Messages {
					switch t := m.Type; t {
					case ENTRANCE:
						err := c.WriteMessage(m.wsMsgType, getEntranceMessage(m))
						if err != nil {
							log.Println("write:", err)
							break
						}
					}
				}
			}()
			tmpl = getTemplate("tmpchat-channel")
		}
	}
	d := TmpchatData{getBgData(), channelName, user, r.Host}
	tmpl.Execute(w, d)
}
