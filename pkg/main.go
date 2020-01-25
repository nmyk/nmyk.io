package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type IndexData struct {
	AnimationDelay string
}

func getAnimationDelay() string {
	return fmt.Sprintf("%ds", -(time.Now().Unix() % 400))
}

func main() {
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			data := IndexData{
				AnimationDelay: getAnimationDelay(),
			}
			tmpl.Execute(w, data)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
