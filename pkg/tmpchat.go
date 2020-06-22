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

func (ch *Chat) Get(channelName string) (*Channel, bool) {
	ch.RLock()
	channel, ok := ch.Channels[channelName]
	ch.RUnlock()
	return channel, ok
}

func (ch *Chat) Set(channelName string, channel *Channel) {
	ch.Lock()
	ch.Channels[channelName] = channel
	ch.Unlock()
}

func (ch *Chat) Delete(channelName string) {
	ch.Lock()
	delete(ch.Channels, channelName)
	ch.Unlock()
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

func (c *Conns) Get(userID string) *Conn {
	c.RLock()
	val := c.Map[userID]
	c.RUnlock()
	return val
}

func (c *Conns) Set(userID string, conn *Conn) {
	c.Lock()
	c.Map[userID] = conn
	c.Unlock()
}

func (c *Conns) Delete(key string) {
	c.Lock()
	delete(c.Map, key)
	c.Unlock()
}

func (c *Conns) Count() int {
	c.RLock()
	n := len(c.Map)
	c.RUnlock()
	return n
}

func (c *Conns) Range(f func(string, *Conn) bool) {
	c.RLock()
	defer c.RUnlock()
	for userID, conn := range c.Map {
		if next := f(userID, conn); !next {
			return
		}
	}
}

type Conn struct {
	WS       *websocket.Conn
	UserName string
}

func (ch *Chat) CreateIfNecessary(channelName string) *Channel {
	var i uint64
	c := &Channel{
		Name:        channelName,
		Connections: &Conns{Map: make(map[string]*Conn)},
		Messages:    make(chan *Message),
		AnonIndex:   &i,
	}
	if existing, ok := Tmpchat.Get(channelName); !ok {
		go c.Run()
		Tmpchat.Set(channelName, c)
		return c
	} else {
		return existing
	}
}

func (c *Channel) GetUsers() []User {
	users := make([]User, c.Connections.Count())
	i := 0
	c.Connections.Range(
		func(userID string, conn *Conn) bool {
			users[i] = User{userID, conn.UserName}
			i++
			return true
		})
	return users
}

func (c *Channel) AddUser() User {
	user := User{uuid.New().String(), fmt.Sprintf("anon_%d", atomic.AddUint64(c.AnonIndex, 1))}
	c.Connections.Set(user.ID, &Conn{nil, user.Name})
	return user
}

func (c *Channel) NameIsAvailable(userName string) bool {
	if userName == "" {
		return false
	}
	c.Connections.Range(func(userID string, conn *Conn) bool {
		if userName == conn.UserName {
			return false
		}
		return true
	})
	return true
}

func (c *Channel) Run() {
	defer c.Close()
MessageLoop:
	for msg := range c.Messages {
		switch msg.Type {
		case Entrance:
			// Associate this websocket conn with the new user so we know
			// which one to close when they send us an Exit.
			c.Connections.Get(msg.FromUser.ID).WS = msg.fromConn
			Reply(msg,
				&Message{
					wsMsgType:   1, // text
					ChannelName: msg.ChannelName,
					Type:        Welcome,
					Content:     c.GetUsers(),
				})
		case Exit:
			_ = c.Connections.Get(msg.FromUser.ID).WS.Close()
			c.Connections.Delete(msg.FromUser.ID)
			if c.Connections.Count() == 0 {
				return
			}
		case NameChange:
			if !c.NameIsAvailable(msg.Content.(string)) {
				// a NameChange rejection is just another NameChange
				// message telling you to change back to your old name.
				Reply(msg,
					&Message{
						wsMsgType: 1, // text
						FromUser:  msg.FromUser,
						Type:      NameChange,
						Content:   msg.FromUser.Name,
					})
				continue MessageLoop
			}
			c.Connections.Get(msg.FromUser.ID).UserName = msg.Content.(string)
		}
		if _, ok := Tmpchat.Get(c.Name); ok {
			c.Broadcast(msg)
		}
	}
}

func (c *Channel) Close() {
	close(c.Messages)
	Tmpchat.Delete(c.Name)
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
	c.Connections.Range(func(_ string, conn *Conn) bool {
		err := conn.WS.WriteMessage(m.wsMsgType, out)
		if err != nil {
			log.Println("write:", err)
			return false
		}
		return true
	})
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
		if ch, ok := Tmpchat.Get(message.ChannelName); ok {
			ch.Messages <- message
		}
	}
}

type tmpchatPageData struct {
	bgData
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
		channel := Tmpchat.CreateIfNecessary(channelName)
		newUser = channel.AddUser()
	}
	d := tmpchatPageData{getBgData(), channelName, newUser, r.Host}
	_ = tmpl.Execute(w, d)
}
