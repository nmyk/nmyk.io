package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

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
	Map map[string]*Member
}

func (c *Members) Get(id string) (*Member, bool) {
	c.RLock()
	member, ok := c.Map[id]
	c.RUnlock()
	return member, ok
}

func (c *Members) Set(id string, member *Member) {
	c.Lock()
	c.Map[id] = member
	c.Unlock()
}

func (c *Members) Delete(id string) {
	c.Lock()
	delete(c.Map, id)
	c.Unlock()
}

func (c *Members) Count() int {
	var n int
	c.Range(func(_ string, member *Member) bool {
		if member.Conn != nil {
			n++
		}
		return true
	})
	return n
}

func (c *Members) Range(f func(string, *Member) bool) {
	c.RLock()
	defer c.RUnlock()
	for id, member := range c.Map {
		if next := f(id, member); !next {
			return
		}
	}
}

type Member struct {
	Conn *websocket.Conn
}

func CreateIfNecessary(channelName string) *Channel {
	var i uint64
	c := &Channel{
		Name:      channelName,
		Members:   &Members{Map: make(map[string]*Member)},
		Messages:  make(chan Message),
		AnonIndex: &i,
	}
	if existing, ok := tmpchat.Get(channelName); !ok {
		go c.Run()
		tmpchat.Set(channelName, c)
		return c
	} else {
		return existing
	}
}

func (c *Channel) AddMember() User {
	user := User{uuid.New().String(), fmt.Sprintf("anon_%d", atomic.AddUint64(c.AnonIndex, 1))}
	c.Members.Set(user.ID, &Member{})
	return user
}

type TURNCreds struct {
	TURNUserName string `json:"username"`
	Pass         string `json:"credential"`
}

func GetTURNCreds(userID string) TURNCreds {
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	turnUserName := fmt.Sprintf("%d:%s", expiresAt, userID)
	k := []byte(os.Getenv("TURN_AUTH_SECRET"))
	h := hmac.New(sha1.New, k)
	h.Write([]byte(turnUserName))
	password := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return TURNCreds{turnUserName, password}
}

func (c *Channel) Run() {
	for msg := range c.Messages {
		switch msg.Type {
		case RTCOffer, RTCAnswer, RTCICECandidate:
			if member, ok := c.Members.Get(msg.ToUserID); ok {
				msg.SendTo(member)
			}
		case TURNCredRequest:
			if member, ok := c.Members.Get(msg.FromUser.ID); ok && member.Conn == nil {
				member.Conn = msg.fromConn
			}
			msg.Reply(
				Message{
					Type:    TURNCredResponse,
					Content: GetTURNCreds(msg.FromUser.ID),
				})
			c.Broadcast(
				Message{
					Type:    Entrance,
					Content: msg.FromUser,
				})
		}
	}
}

func (c *Channel) Close() {
	close(c.Messages)
	tmpchat.Delete(c.Name)
	log.Println(fmt.Sprintf("cleaned up empty channel %s", c.Name))
}

func (m Message) Reply(msg Message) {
	response, _ := json.Marshal(msg)
	if err := m.fromConn.WriteMessage(1, response); err != nil {
		log.Println("write:", err)
	}
}

func (m Message) SendTo(member *Member) {
	message, _ := json.Marshal(m)
	if err := member.Conn.WriteMessage(1, message); err != nil {
		log.Println("write:", err)
	}
}

func (c *Channel) Broadcast(msg Message) {
	message, _ := json.Marshal(msg)
	c.Members.Range(func(_ string, member *Member) bool {
		if err := member.Conn.WriteMessage(1, message); err != nil {
			log.Println("write:", err)
		}
		return true
	})
}

type Message struct {
	fromConn    *websocket.Conn
	ChannelName string      `json:"channel_name,omitempty"`
	FromUser    User        `json:"from_user,omitempty"`
	ToUserID    string      `json:"to_user_id,omitempty"`
	Type        EventType   `json:"type"`
	Content     interface{} `json:"content"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EventType int

const (
	Entrance EventType = iota
	RTCOffer
	RTCAnswer
	RTCICECandidate
	TURNCredRequest
	TURNCredResponse
)

func signalingHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool {
			originURL, err := url.Parse(r.Header["Origin"][0])
			if err != nil {
				return false
			}
			return originURL.String() == os.Getenv("TMPCHAT_URL")
		},
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	var userID, channelName string
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
		if channelName == "" {
			channelName = message.ChannelName
			userID = message.FromUser.ID
		}
		message.fromConn = c
		if ch, ok := tmpchat.Get(message.ChannelName); ok {
			ch.Messages <- message
		}
	}
	if ch, ok := tmpchat.Get(channelName); ok {
		ch.Members.Delete(userID)
		if ch.Members.Count() == 0 {
			ch.Close()
		}
	}
}

type tmpchatPageData struct {
	bgData
	ChannelName  string
	User         User
	AppURL       string
	SignalingURL string
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
		channel := CreateIfNecessary(channelName)
		newUser = channel.AddMember()
	}
	d := tmpchatPageData{getBgData(),
		channelName,
		newUser,
		os.Getenv("TMPCHAT_URL"),
		os.Getenv("TMPCHAT_SIGNALING_URL"),
	}
	_ = tmpl.Execute(w, d)
}
