package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// bgAnimationDurationSeconds is how long a full cycle of the color-changing
// background animation will take. It should be short enough that folks will
// eventually notice the colors changing, but long enough so that when they
// notice, it feels like discovering a hidden treat.
const bgAnimationDurationSeconds = 400

type bgData struct {
	BgAnimationDuration int
	BgAnimationDelay    int
}

func getTemplate(desc string, funcMap template.FuncMap) *template.Template {
	t, _ := ioutil.ReadFile(
		fmt.Sprintf("web/templates/%s.gohtml", desc),
	)
	return template.Must(template.New(desc).Funcs(funcMap).Parse(string(t)))
}

func getBgData() bgData {
	return bgData{
		// If we do this on the frontend, users will see a split second of the background at
		// 0s delay before it updates. This way the background looks the same for everyone
		// _and_ there are no visible transient states that make the site look cheap while
		// it's loading. If I didn't care, I'd just use Squarespace ;)
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
	tmpchatMux.HandleFunc("/", tmpchatHandler)
	go func() {
		log.Fatal(http.ListenAndServe(":8081", tmpchatMux))
	}()

	nmykMux := http.NewServeMux()
	nmykMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := getTemplate("index", template.FuncMap{})
		_ = tmpl.Execute(w, getBgData())
	})
	log.Fatal(http.ListenAndServe(":8080", nmykMux))
}
