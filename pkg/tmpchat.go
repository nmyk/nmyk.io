package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Chat struct {
	sync.RWMutex
	Channels map[string]*Channel
}

var Tmpchat = &Chat{
	Channels: make(map[string]*Channel),
}

type Channel struct {
	Name        string
	Connections *Conns
	Messages    chan *Message
	AnonIndex   *uint64
}

type Conns struct {
	sync.RWMutex
	Map map[string]*Conn
}

type Conn struct {
	WS       *websocket.Conn
	UserName string
}

func (ch *Chat) AddChannelIfNotExists(channelName string) {
	if channelName == "" {
		return
	}
	var i uint64
	c := &Channel{
		Name:        channelName,
		Connections: &Conns{Map: make(map[string]*Conn)},
		Messages:    make(chan *Message),
		AnonIndex:   &i,
	}
	Tmpchat.Lock()
	defer Tmpchat.Unlock()
	if _, ok := Tmpchat.Channels[channelName]; !ok {
		go c.run()
		Tmpchat.Channels[channelName] = c
	}
}

func (c *Channel) GetUsers() []User {
	c.Connections.RLock()
	defer c.Connections.RUnlock()
	users := make([]User, len(c.Connections.Map))
	i := 0
	for userId := range c.Connections.Map {
		users[i] = User{userId, (c.Connections.Map)[userId].UserName}
		i++
	}
	return users
}

func (c *Channel) AddUser() User {
	user := User{uuid.New().String(), fmt.Sprintf("anon_%d", atomic.AddUint64(c.AnonIndex, 1))}
	c.Connections.Lock()
	c.Connections.Map[user.ID] = &Conn{nil, user.Name}
	c.Connections.Unlock()
	return user
}

func (c *Channel) NameIsAvailable(userName string) bool {
	if userName == "" {
		return false
	}
	c.Connections.RLock()
	defer c.Connections.RUnlock()
	for userId := range c.Connections.Map {
		if userName == c.Connections.Map[userId].UserName {
			return false
		}
	}
	return true
}

func (c *Channel) run() {
	defer c.close()
MessageLoop:
	for m := range c.Messages {
		switch m.Type {
		case Entrance:
			c.Connections.Lock()
			c.Connections.Map[m.FromUser.ID].WS = m.fromConn
			c.Connections.Unlock()
			Reply(m,
				&Message{
					wsMsgType:   1, // text
					ChannelName: m.ChannelName,
					Type:        Welcome,
					Content:     c.GetUsers(),
				})
		case Exit:
			c.Connections.Lock()
			_ = c.Connections.Map[m.FromUser.ID].WS.Close()
			delete(c.Connections.Map, m.FromUser.ID)
			c.Connections.Unlock()
			c.Connections.RLock()
			if len(c.Connections.Map) == 0 {
				return
			}
			c.Connections.RUnlock()
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
			c.Connections.Lock()
			c.Connections.Map[m.FromUser.ID].UserName = m.Content.(string)
			c.Connections.Unlock()
		}
		c.Broadcast(m)
	}
}

func (c *Channel) close() {
	close(c.Messages)
	Tmpchat.Lock()
	delete(Tmpchat.Channels, c.Name)
	Tmpchat.Unlock()
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
	c.Connections.RLock()
	defer c.Connections.RUnlock()
	for id := range c.Connections.Map {
		err := c.Connections.Map[id].WS.WriteMessage(m.wsMsgType, out)
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
		Tmpchat.Channels[message.ChannelName].Messages <- message
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
		Tmpchat.AddChannelIfNotExists(channelName)
		newUser = Tmpchat.Channels[channelName].AddUser()
	}
	d := TmpchatPageData{getBgData(), channelName, newUser, r.Host}
	_ = tmpl.Execute(w, d)
}
