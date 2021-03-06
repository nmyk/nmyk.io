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
	Members  *Members
	Messages chan Message
}

type Members struct {
	sync.RWMutex
	m map[string]*websocket.Conn
}

func (m *Members) Get(id string) (*websocket.Conn, bool) {
	m.RLock()
	member, ok := m.m[id]
	m.RUnlock()
	return member, ok
}

func (m *Members) Set(id string, member *websocket.Conn) {
	m.Lock()
	m.m[id] = member
	m.Unlock()
}

func (m *Members) Delete(id string) {
	m.Lock()
	delete(m.m, id)
	m.Unlock()
}

func (m *Members) Count() int {
	var n int
	m.Range(func(member *websocket.Conn) bool {
		if member != nil {
			n++
		}
		return true
	})
	return n
}

func (m *Members) Range(f func(*websocket.Conn) bool) {
	m.RLock()
	defer m.RUnlock()
	for _, member := range m.m {
		if next := f(member); !next {
			return
		}
	}
}

func Materialize(channelName string) *Channel {
	c := &Channel{
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

func Collect(channelName string, userID string) {
	if ch, ok := tmpchat.Get(channelName); ok {
		ch.Broadcast(Message{Type: Exit, From: userID})
		ch.Members.Delete(userID)
		if ch.Members.Count() == 0 {
			close(ch.Messages)
			tmpchat.Delete(channelName)
		}
	}
}

func GetTURNCreds(userID string) TURNCreds {
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	turnUserName := fmt.Sprintf("%d:%s", expiresAt, userID)
	h := hmac.New(sha1.New, []byte(os.Getenv("TURN_AUTH_SECRET")))
	h.Write([]byte(turnUserName))
	password := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return TURNCreds{turnUserName, password}
}

type TURNCreds struct {
	Username string `json:"username"`
	Password string `json:"credential"`
}

func (c *Channel) Run() {
	for msg := range c.Messages {
		switch msg.Type {
		case TURNCredRequest:
			if member, ok := c.Members.Get(msg.From); ok {
				Message{
					Type:    TURNCredResponse,
					Content: GetTURNCreds(msg.From),
				}.SendTo(member)
			}
			c.Broadcast(
				Message{
					Type: Entrance,
					Content: struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					}{
						msg.From,
						msg.Content.(string),
					},
				})
		case RTCOffer, RTCAnswer, RTCICECandidate:
			if member, ok := c.Members.Get(msg.To); ok {
				msg.SendTo(member)
			}
		}
	}
}

type Message struct {
	To      string      `json:"to,omitempty"`
	From    string      `json:"from,omitempty"`
	Type    EventType   `json:"type"`
	Content interface{} `json:"content"`
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
	defer Collect(channelName, userID)
	for {
		_, rawSignal, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		message := Message{}
		if err := json.Unmarshal(rawSignal, &message); err != nil {
			continue
		}
		if ch, ok := tmpchat.Get(channelName); ok {
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
