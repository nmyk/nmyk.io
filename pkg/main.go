package main

import (
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

type IndexData struct {
	BgAnimationDuration int
	BgAnimationDelay    int
}

func main() {
	tmpl := template.Must(template.ParseFiles("web/templates/index.gohtml"))
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d := IndexData{
			BgAnimationDuration: bgAnimationDurationSeconds,
			BgAnimationDelay:    int(-time.Now().Unix() % bgAnimationDurationSeconds),
		}
		tmpl.Execute(w, d)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
