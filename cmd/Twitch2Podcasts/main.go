package main

import (
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
)

func main() {
	fmt.Println("App Running")

	dbmanager.InitDb(dbmanager.ConnectDb())

	indexHandler := func(w http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, nil)
	}

	userHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			http.Redirect(w, req, "/", http.StatusMovedPermanently)
			return
		}

		channel := strings.ToLower(req.PostFormValue("channel"))
		var search models.Search

		if len(channel) > 1 && len(channel) <= 25 {
			search = dbmanager.SearchChannel(channel, dbmanager.ConnectDb())
		}

		if search.Channel == nil {
			htmlStr := fmt.Sprintf("<h2>'%s' Channel not exist </h2>", channel)
			tmpl, _ := template.New("t").Parse(htmlStr)
			tmpl.Execute(w, nil)
		} else {
			tmpl := template.Must(template.ParseFiles("templates/channel.html"))
			tmpl.Execute(w, search)
		}
	}

	feedHandler := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/xml")

		channel := req.PathValue("channel")

		var channelId *int

		if len(channel) > 1 && len(channel) <= 25 {
			channelId = dbmanager.GetChannelId(channel, dbmanager.ConnectDb())

			if channelId == nil {
				http.Redirect(w, req, "/", http.StatusMovedPermanently)
				return
			}
			rss := dbmanager.GetRss(channelId, dbmanager.ConnectDb())

			if rss == nil {
				http.Redirect(w, req, "/", http.StatusMovedPermanently)
				return
			}
			w.Write([]byte(rss.Rss))
		}
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/channel/", userHandler)
	http.HandleFunc("/feed/{channel}", feedHandler)

	fmt.Println("WebServer Running")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
