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
	BgAnimationDelay int
}

func main() {
	tmpl := template.Must(template.ParseFiles("web/templates/index.gohtml"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Set params for the color-changing background here rather than
			// on the client so we don't have to see a split second of wrong
			// background before the body element's `.onload` event handler
			// updates the element's `animation-delay` style property.
			d := IndexData{
				BgAnimationDuration: bgAnimationDurationSeconds,
				// Setting `animation-delay` to -X seconds begins the animation
				// as though it had already been playing for X seconds. We use
				// a function of time.Now() to make it so that everyone sees the
				// same thing, as though it were a natural phenomenon.
				BgAnimationDelay: int(-time.Now().Unix() % bgAnimationDurationSeconds),
			}
			tmpl.Execute(w, d)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
