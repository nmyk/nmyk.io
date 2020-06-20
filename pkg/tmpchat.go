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

func (c *Channel) Start() {
	for m := range c.Messages {
		switch t := m.Type; t {
		case ENTRANCE:
			text := fmt.Sprintf("<span class=\"%s\">%s</span> joined", m.FromUser.ID, m.FromUser.Name)
			announcement := &Message{
				wsMsgType:   m.wsMsgType,
				ChannelName: m.ChannelName,
				FromUser:    nil,
				Type:        ENTRANCE,
				Text:        text,
			}
			c.Broadcast(announcement)
		case EXIT:
			text := fmt.Sprintf("<span class=\"%s\">%s</span> left", m.FromUser.ID, m.FromUser.Name)
			announcement := &Message{
				wsMsgType:   m.wsMsgType,
				ChannelName: m.ChannelName,
				FromUser:    nil,
				Type:        EXIT,
				Text:        text,
			}
			c.Broadcast(announcement)
		case NAME_CHANGE, CLEAR:
			c.Broadcast(m)
		}
	}
}

func (c *Channel) Broadcast(m *Message) {
	out, _ := json.Marshal(m)
	for _, conn := range c.Connections {
		err := conn.WriteMessage(m.wsMsgType, out)
		if err != nil {
			log.Println("write:", err)
		}
	}
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
	// Unmarshal all my messages and send them to the channel
	var channelName string
	for {
		mt, rawSignal, err := c.ReadMessage()
		if err != nil && channelName != "" {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", rawSignal)
		message := &Message{}
		if err := json.Unmarshal(rawSignal, message); err != nil {
			continue
		}
		channelName = message.ChannelName
		if message.Type == ENTRANCE {
			// First msg from this user. Need to associate the connection with the UserId.
			userID := message.FromUser.ID
			for i, conn := range Tmpchat[channelName].Connections {
				if conn.User.ID == userID {
					Tmpchat[channelName].Connections[i].Conn = c
					break
				}
			}
		}
		if message.Type == EXIT && len(Tmpchat[channelName].Connections) == 1 {
			// Last user out should turn off the lights to prevent a goroutine leak.
			close(Tmpchat[channelName].Messages)
			delete(Tmpchat, channelName)
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
		tmpl = getTemplate("tmpchat-channel")
		user = User{
			ID:   uuid.New().String(),
			Name: "anon_00", //TODO: some kind of name initialization. maybe just autoincrementing anon_XX
		}
		if _, ok := Tmpchat[channelName]; !ok {
			conn := &Connection{nil, &user}
			messages := make(chan *Message)
			channel := Channel{
				Connections: []*Connection{conn},
				Messages:    messages,
			}
			Tmpchat[channelName] = &channel
			go channel.Start()
		} else {
			conn := &Connection{nil, &user}
			Tmpchat[channelName].Connections = append(Tmpchat[channelName].Connections, conn)
		}
	}
	d := TmpchatData{getBgData(), channelName, user, r.Host}
	tmpl.Execute(w, d)
}
