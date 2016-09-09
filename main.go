package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/codegangsta/negroni"
	"github.com/gophergala2016/thunderbird"
	"github.com/gorilla/mux"
)

var homeTempl = template.Must(template.ParseFiles("home.html"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	scheme := "ws"
	if os.Getenv("GO_ENV") == "production" {
		scheme = "wss"
	}
	url := fmt.Sprintf("%s://%s/ws", scheme, r.Host)

	homeTempl.Execute(w, url)
}

type RoomChannel struct {
	tb *thunderbird.Thunderbird
}

func (rc *RoomChannel) Received(event thunderbird.Event) {
	switch event.Type {
	case "message":
		rc.tb.Broadcast(event.Channel, event.Body)
	}
}

func main() {
	tb := thunderbird.New()
	ch := &RoomChannel{tb}
	tb.HandleChannel("room", ch)

	router := mux.NewRouter()
	router.HandleFunc("/", serveHome).Methods("GET")
	router.Handle("/ws", tb.HTTPHandler())

	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir("../client/lib")), // serve thunderbird.js
		negroni.NewStatic(http.Dir("public")),        // serve other assets
	)
	n.UseHandler(router)

	n.Run(":" + os.Getenv("PORT"))
}
