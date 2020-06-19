package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

// bgAnimationDurationSeconds is how long a full cycle of the color-changing
// background animation will take. It should be short enough that folks will
// eventually notice the colors changing, but long enough so that when they
// notice, it feels like discovering a hidden treat.
const bgAnimationDurationSeconds = 400

type BgData struct {
	BgAnimationDuration int
	BgAnimationDelay    int
}

type TmpchatData struct {
	BgData
	ChannelName string
	AppHost     string // So this works seamlessly in dev (localhost) and prod (tmpch.at)
}

func getTemplate(desc string) *template.Template {
	return template.Must(template.ParseFiles(fmt.Sprintf("web/templates/%s.gohtml", desc)))
}

func getBgData() BgData {
	return BgData{
		BgAnimationDuration: bgAnimationDurationSeconds,
		BgAnimationDelay:    int(-time.Now().Unix() % bgAnimationDurationSeconds),
	}
}

func main() {
	signalingMux := http.NewServeMux()
	signalingMux.HandleFunc("/", signalingHandler)
	go func() {
		log.Fatal(http.ListenAndServe(":7070", signalingMux))
	}()

	tmpchatMux := http.NewServeMux()
	tmpchatMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.Host)
		channelName := r.URL.Path[1:]
		var tmpl *template.Template
		if channelName == "" {
			tmpl = getTemplate("tmpchat-index")
		} else {
			tmpl = getTemplate("tmpchat-channel")
		}
		d := TmpchatData{getBgData(), channelName, r.Host}
		tmpl.Execute(w, d)
	})
	go func() {
		log.Fatal(http.ListenAndServe(":8081", tmpchatMux))
	}()

	nmykMux := http.NewServeMux()
	nmykMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := getTemplate("index")
		tmpl.Execute(w, getBgData())
	})
	log.Fatal(http.ListenAndServe(":8080", nmykMux))
}
