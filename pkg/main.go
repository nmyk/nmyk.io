package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
)

type IndexData struct {
	AnimationDelay string
}

func main() {
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			data := IndexData{
				AnimationDelay: fmt.Sprintf("%ds", -rand.Intn(400)),
			}
			tmpl.Execute(w, data)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
