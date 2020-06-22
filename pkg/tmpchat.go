package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Chat struct {
	sync.RWMutex
	Channels map[string]*Channel
}

var tmpchat = &Chat{
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
	Name      string
	Members   *Members
	Messages  chan Message
	AnonIndex *uint64
}

type Members struct {
	sync.RWMutex
	Map map[string]*Conn
}

func (c *Members) Get(userID string) (*Conn, bool) {
	c.RLock()
	conn, ok := c.Map[userID]
	c.RUnlock()
	return conn, ok
}

func (c *Members) Set(userID string, conn *Conn) {
	c.Lock()
	c.Map[userID] = conn
	c.Unlock()
}

func (c *Members) Delete(userID string) {
	c.Lock()
	delete(c.Map, userID)
	c.Unlock()
}

func (c *Members) Count() int {
	var n int
	c.Range(func(_ string, conn *Conn) bool {
		if conn.WS != nil {
			n++
		}
		return true
	})
	return n
}

func (c *Members) Range(f func(string, *Conn) bool) {
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
		Name:      channelName,
		Members:   &Members{Map: make(map[string]*Conn)},
		Messages:  make(chan Message),
		AnonIndex: &i,
	}
	if existing, ok := ch.Get(channelName); !ok {
		go c.Run()
		ch.Set(channelName, c)
		return c
	} else {
		return existing
	}
}

func (c *Channel) GetUsers() []User {
	users := make([]User, c.Members.Count())
	i := 0
	c.Members.Range(
		func(userID string, conn *Conn) bool {
			users[i] = User{userID, conn.UserName}
			i++
			return true
		})
	return users
}

func (c *Channel) AddUser() User {
	user := User{uuid.New().String(), fmt.Sprintf("anon_%d", atomic.AddUint64(c.AnonIndex, 1))}
	for !c.NameIsAvailable(user.Name) {
		user.Name = fmt.Sprintf("anon_%d", atomic.AddUint64(c.AnonIndex, 1))
	}
	c.Members.Set(user.ID, &Conn{nil, user.Name})
	return user
}

func (c *Channel) NameIsAvailable(userName string) bool {
	if userName == "" {
		return false
	}
	isAvailable := true
	c.Members.Range(func(userID string, conn *Conn) bool {
		if userName == conn.UserName {
			isAvailable = false
			return false
		}
		return true
	})
	return isAvailable
}

func (c *Channel) Run() {
	defer c.Close()
	for msg := range c.Messages {
		switch msg.Type {
		case Entrance:
			// Associate this websocket conn with the new user so we know
			// which one to close when they send us an Exit.
			if conn, ok := c.Members.Get(msg.FromUser.ID); ok && conn.WS == nil {
				conn.WS = msg.fromConn
				// A Welcome message lets new members know who else is here.
				msg.Reply(
					Message{
						Type:    Welcome,
						Content: c.GetUsers(),
					})
			} else {
				continue
			}
		case Exit:
			if conn, ok := c.Members.Get(msg.FromUser.ID); ok && msg.fromConn == conn.WS {
				_ = conn.WS.Close()
				c.Members.Delete(msg.FromUser.ID)
				if c.Members.Count() == 0 {
					return
				}
			} else {
				continue
			}
		case NameChange:
			if !c.NameIsAvailable(msg.Content.(string)) {
				// a NameChange rejection is just another NameChange
				// message telling you to change back to your old name.
				msg.Reply(
					Message{
						FromUser: msg.FromUser,
						Type:     NameChange,
						Content:  msg.FromUser.Name,
					})
				continue
			}
			if conn, ok := c.Members.Get(msg.FromUser.ID); ok && msg.fromConn == conn.WS {
				conn.UserName = msg.Content.(string)
			} else {
				continue
			}
		}
		c.Broadcast(msg)
	}
}

func (c *Channel) Close() {
	close(c.Messages)
	tmpchat.Delete(c.Name)
	log.Println(fmt.Sprintf("cleaned up empty channel %s", c.Name))
}

func (call Message) Reply(response Message) {
	msg, _ := json.Marshal(response)
	if err := call.fromConn.WriteMessage(1, msg); err != nil {
		log.Println("write:", err)
	}
}

func (c *Channel) Broadcast(m Message) {
	out, _ := json.Marshal(m)
	c.Members.Range(func(_ string, conn *Conn) bool {
		if err := conn.WS.WriteMessage(1, out); err != nil {
			log.Println("write:", err)
		}
		return true
	})
}

type Message struct {
	fromConn    *websocket.Conn
	ChannelName string      `json:"channel_name"`
	FromUser    User        `json:"from_user,omitempty"`
	Type        EventType   `json:"type"`
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
		_, rawSignal, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", rawSignal)
		message := Message{}
		if err := json.Unmarshal(rawSignal, &message); err != nil {
			continue
		}
		message.fromConn = c
		if ch, ok := tmpchat.Get(message.ChannelName); ok {
			ch.Messages <- message
		}
	}
}

type tmpchatPageData struct {
	bgData
	ChannelName   string
	User          User
	AppHost       string
	SignalingHost string
}

func tmpchatHandler(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Path[1:] // Omit leading slash in path
	var tmpl *template.Template
	var newUser User
	fm := template.FuncMap{
		"safeURL": func(u string) template.URL { return template.URL(u) },
	}
	if channelName == "" {
		tmpl = getTemplate("tmpchat-index", fm)
	} else {
		tmpl = getTemplate("tmpchat-channel", fm)
		log.Println(fmt.Sprintf("%v", tmpl))
		channel := tmpchat.CreateIfNecessary(channelName)
		newUser = channel.AddUser()
	}
	d := tmpchatPageData{getBgData(),
		channelName,
		newUser,
		os.Getenv("TMPCHAT_HOST"),
		os.Getenv("TMPCHAT_SIGNALING_HOST"),
	}
	_ = tmpl.Execute(w, d)
}
