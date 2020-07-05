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
	"time"

	"github.com/gorilla/websocket"
)

type Chat struct {
	sync.RWMutex
	Turnstile *Turnstile
	Channels  map[string]*Channel
}

var tmpchat = &Chat{
	Turnstile: &Turnstile{m: make(map[string]struct{})},
	Channels:  make(map[string]*Channel),
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
	Name     string
	Members  *Members
	Messages chan Message
}

type Members struct {
	sync.RWMutex
	m map[string]*websocket.Conn
}

func (c *Members) Get(id string) (*websocket.Conn, bool) {
	c.RLock()
	member, ok := c.m[id]
	c.RUnlock()
	return member, ok
}

func (c *Members) Set(id string, member *websocket.Conn) {
	c.Lock()
	c.m[id] = member
	c.Unlock()
}

func (c *Members) Delete(id string) {
	c.Lock()
	delete(c.m, id)
	c.Unlock()
}

func (c *Members) Count() int {
	var n int
	c.Range(func(member *websocket.Conn) bool {
		if member != nil {
			n++
		}
		return true
	})
	return n
}

func (c *Members) Range(f func(*websocket.Conn) bool) {
	c.RLock()
	defer c.RUnlock()
	for _, member := range c.m {
		if next := f(member); !next {
			return
		}
	}
}

func Materialize(channelName string) *Channel {
	c := &Channel{
		Name:     channelName,
		Members:  &Members{m: make(map[string]*websocket.Conn)},
		Messages: make(chan Message),
	}
	if existing, ok := tmpchat.Get(channelName); !ok {
		go c.Run()
		tmpchat.Set(channelName, c)
		return c
	} else {
		return existing
	}
}

type TURNCreds struct {
	Username string `json:"username"`
	Password string `json:"credential"`
}

func GetTURNCreds(userID string) TURNCreds {
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	turnUserName := fmt.Sprintf("%d:%s", expiresAt, userID)
	h := hmac.New(sha1.New, []byte(os.Getenv("TURN_AUTH_SECRET")))
	h.Write([]byte(turnUserName))
	password := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return TURNCreds{turnUserName, password}
}

func (c *Channel) Run() {
	for msg := range c.Messages {
		switch msg.Type {
		case TURNCredRequest:
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
		case RTCOffer, RTCAnswer, RTCICECandidate:
			if member, ok := c.Members.Get(msg.ToUserID); ok {
				msg.SendTo(member)
			}
		}
	}
}

func (c *Channel) Close() {
	close(c.Messages)
	tmpchat.Delete(c.Name)
	log.Println(fmt.Sprintf("cleaned up empty channel %s", c.Name))
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
	Name string `json:"name,omitempty"`
}

type EventType int

const (
	Entrance EventType = iota
	Exit
	RTCOffer
	RTCAnswer
	RTCICECandidate
	TURNCredRequest
	TURNCredResponse
)

func (m Message) Reply(msg Message) {
	response, _ := json.Marshal(msg)
	if err := m.fromConn.WriteMessage(1, response); err != nil {
		log.Println("write:", err)
	}
}

func (m Message) SendTo(member *websocket.Conn) {
	message, _ := json.Marshal(m)
	if err := member.WriteMessage(1, message); err != nil {
		log.Println("write:", err)
	}
}

func (c *Channel) Broadcast(msg Message) {
	message, _ := json.Marshal(msg)
	c.Members.Range(func(member *websocket.Conn) bool {
		if err := member.WriteMessage(1, message); err != nil {
			log.Println("write:", err)
		}
		return true
	})
}

func signalingHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	var channelName, userID string
	if vals, ok := params["channelName"]; ok {
		channelName = vals[0]
	}
	if vals, ok := params["userID"]; ok {
		userID = vals[0]
	}
	if ok := tmpchat.Turnstile.Admit(userID); !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool {
			var origin string
			if vals, ok := r.Header["Origin"]; ok {
				origin = vals[0]
			}
			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}
			return originURL.String() == os.Getenv("TMPCHAT_URL")
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	Materialize(channelName).Members.Set(userID, conn)
	for {
		_, rawSignal, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			if ch, ok := tmpchat.Get(channelName); ok {
				ch.Broadcast(Message{Type: Exit, FromUser: User{ID: userID}})
				ch.Members.Delete(userID)
				if ch.Members.Count() == 0 {
					ch.Close()
				}
			}
			break
		}
		message := Message{}
		if err := json.Unmarshal(rawSignal, &message); err != nil {
			continue
		}
		message.fromConn = conn
		if ch, ok := tmpchat.Get(message.ChannelName); ok {
			ch.Messages <- message
		}
	}
}

type Turnstile struct {
	sync.RWMutex
	m map[string]struct{}
}

func (t *Turnstile) Register(userID string) bool {
	t.RLock()
	_, ok := t.m[userID]
	t.RUnlock()
	if ok {
		return false
	}
	t.Lock()
	t.m[userID] = struct{}{}
	t.Unlock()
	return true
}

func (t *Turnstile) Admit(userID string) bool {
	t.RLock()
	_, ok := t.m[userID]
	t.RUnlock()
	if ok {
		t.Lock()
		delete(t.m, userID)
		t.Unlock()
		return true
	}
	return false
}

type tmpchatPageData struct {
	ChannelName  string
	UserID       string
	AppURL       string
	SignalingURL string
}

func tmpchatHandler(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Path[1:] // Omit leading slash in path
	var tmpl *template.Template
	fm := template.FuncMap{
		"safeURL": func(u string) template.URL { return template.URL(u) },
	}
	if channelName == "" {
		tmpl = getTemplate("tmpchat-index", fm)
	} else {
		tmpl = getTemplate("tmpchat-channel", fm)
	}
	userID := uuid.New().String()
	if ok := tmpchat.Turnstile.Register(userID); !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	d := tmpchatPageData{
		channelName,
		userID,
		os.Getenv("TMPCHAT_URL"),
		os.Getenv("TMPCHAT_SIGNALING_URL"),
	}
	_ = tmpl.Execute(w, d)
}
