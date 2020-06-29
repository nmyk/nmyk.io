package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

func getTemplate(desc string, funcMap template.FuncMap) *template.Template {
	t, _ := ioutil.ReadFile(
		fmt.Sprintf("web/templates/%s.gohtml", desc),
	)
	return template.Must(template.New(desc).Funcs(funcMap).Parse(string(t)))
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
		_ = tmpl.Execute(w, nil)
	})
	log.Fatal(http.ListenAndServe(":8080", nmykMux))
}
