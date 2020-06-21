package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var Tmpchat = map[string]*Channel{}

type Channel struct {
	Name        string
	Connections Connections
	Messages    chan *Message
	AnonIndex   *uint64
}

type Connections map[string]*Conn

type Conn struct {
	WS       *websocket.Conn
	UserName string
}

func NewChannel(channelName string) *Channel {
	var i uint64
	c := &Channel{
		Name:        channelName,
		Connections: Connections{},
		Messages:    make(chan *Message),
		AnonIndex:   &i,
	}
	go c.run()
	return c
}

func (c *Channel) GetUsers() []User {
	users := make([]User, len(c.Connections))
	i := 0
	for userId := range c.Connections {
		users[i] = User{userId, (c.Connections)[userId].UserName}
		i++
	}
	return users
}

func (c *Channel) AddUser() User {
	user := User{uuid.New().String(), fmt.Sprintf("anon_%d", atomic.AddUint64(c.AnonIndex, 1))}
	c.Connections[user.ID] = &Conn{nil, user.Name}
	return user
}

func (c *Channel) NameIsAvailable(userName string) bool {
	if userName == "" {
		return false
	}
	for userId := range c.Connections {
		if userName == c.Connections[userId].UserName {
			return false
		}
	}
	return true
}

func (c *Channel) run() {
	defer c.close()
MessageLoop:
	for m := range c.Messages {
		switch t := m.Type; t {
		case Entrance:
			c.Connections[m.FromUser.ID].WS = m.fromConn
			Reply(m,
				&Message{
					wsMsgType:   1, // text
					ChannelName: m.ChannelName,
					Type:        Welcome,
					Content:     c.GetUsers(),
				})
		case Exit:
			_ = c.Connections[m.FromUser.ID].WS.Close()
			delete(c.Connections, m.FromUser.ID)
			if len(c.Connections) == 0 {
				return
			}
		case NameChange:
			if !c.NameIsAvailable(m.Content.(string)) {
				// a NameChange rejection is just a normal NameChange
				// message telling you to change back to your old name.
				Reply(m,
					&Message{
						wsMsgType: 1, // text
						FromUser:  m.FromUser,
						Type:      NameChange,
						Content:   m.FromUser.Name,
					})
				continue MessageLoop
			}
			c.Connections[m.FromUser.ID].UserName = m.Content.(string)
			log.Print(c.GetUsers())
		}
		c.Broadcast(m)
	}
}

func (c *Channel) close() {
	close(c.Messages)
	delete(Tmpchat, c.Name)
	log.Println(fmt.Sprintf("cleaned up empty channel %s", c.Name))
}

func Reply(to *Message, reply *Message) {
	m, _ := json.Marshal(reply)
	err := to.fromConn.WriteMessage(reply.wsMsgType, m)
	if err != nil {
		log.Println("write:", err)
	}
}

func (c *Channel) Broadcast(m *Message) {
	out, _ := json.Marshal(m)
	for id := range c.Connections {
		err := c.Connections[id].WS.WriteMessage(m.wsMsgType, out)
		if err != nil {
			log.Println("write:", err)
		}
	}
}

type Message struct {
	fromConn    *websocket.Conn
	wsMsgType   int         // Websocket message type as defined in RFC 6455, section 11.8
	ChannelName string      `json:"channel_name"`
	FromUser    User        `json:"from_user,omitempty"`
	Type        EventType   `json:"type"` // Tmpchat-specific signaling event type
	Content     interface{} `json:"content"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EventType int

const (
	// Event type 0 denotes a user chat message, which we never see here.
	Entrance EventType = iota + 1
	Exit
	NameChange
	Clear
	Welcome
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
		message.fromConn = c
		message.wsMsgType = mt
		Tmpchat[message.ChannelName].Messages <- message
	}
}

type TmpchatPageData struct {
	BgData
	ChannelName string
	User        User
	AppHost     string // So this works seamlessly in dev (localhost) and prod (tmpch.at)
}

func tmpchatHandler(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Path[1:] // Omit leading slash in path
	var tmpl *template.Template
	var newUser User
	if channelName == "" {
		tmpl = getTemplate("tmpchat-index")
	} else {
		tmpl = getTemplate("tmpchat-channel")
		if _, ok := Tmpchat[channelName]; !ok {
			Tmpchat[channelName] = NewChannel(channelName)
		}
		newUser = Tmpchat[channelName].AddUser()
	}
	d := TmpchatPageData{getBgData(), channelName, newUser, r.Host}
	_ = tmpl.Execute(w, d)
}
