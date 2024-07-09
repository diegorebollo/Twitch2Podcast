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

	dbCon := dbmanager.ConnectDb()

	// go jobs.UpdateVods(dbCon)

	// go dbmanager.SearchChannel("chiclanafriends", dbTest)
	// go dbmanager.SearchChannel("pokimane", dbTest)
	// go dbmanager.SearchChannel("ibai", dbTest)
	// go dbmanager.SearchChannel("illojuan", dbTest)
	// go dbmanager.SearchChannel("orslok", dbTest)
	// go dbmanager.SearchChannel("el_yuste", dbTest)
	// go dbmanager.SearchChannel("elxokas", dbTest)
	// go dbmanager.SearchChannel("ikurotime", dbTest)
	// go dbmanager.SearchChannel("alexelcapo", dbTest)
	// go dbmanager.SearchChannel("chicocartera", dbTest)
	// go dbmanager.SearchChannel("jujalag", dbTest)
	// go dbmanager.SearchChannel("knekro", dbTest)
	// go dbmanager.SearchChannel("eurogamer_es", dbTest)
	// go dbmanager.SearchChannel("viviendoenlacalle", dbTest)
	// go dbmanager.SearchChannel("rubius", dbTest)
	// go dbmanager.SearchChannel("auronplay", dbTest)
	// go dbmanager.SearchChannel("littleragergirl", dbTest)
	// go dbmanager.SearchChannel("anujbost", dbTest)
	// go dbmanager.SearchChannel("5ro4", dbTest)
	// go dbmanager.SearchChannel("japanwolf", dbTest)
	// go dbmanager.SearchChannel("llunaclark", dbTest)
	// go dbmanager.SearchChannel("pazos64", dbTest)
	// go dbmanager.SearchChannel("pandarina", dbTest)
	// go dbmanager.SearchChannel("rickyedit", dbTest)
	// go dbmanager.SearchChannel("elrichmc", dbTest)
	// go dbmanager.SearchChannel("kaicenat", dbTest)

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))

	indexHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Print("index ", req.UserAgent())
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, nil)
	}

	userHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			http.Redirect(w, req, "/", http.StatusMovedPermanently)
			return
		}

		channel := strings.ToLower(req.PostFormValue("channel"))

		if strings.Contains(channel, "twitch.tv/") {
			split := strings.Split(channel, "/")

			channel = split[len(split)-1]

			if len(channel) == 0 {
				channel = split[len(split)-2]
			}
		}

		var search models.Search

		if len(channel) > 1 && len(channel) <= 25 {
			search = dbmanager.SearchChannel(channel, dbCon)
		}

		if search.Channel == nil {
			htmlStr := "<h3 class='channel-not-found'>Channel does not exist</h3>"

			if len(channel) > 25 {
				htmlStr = "<h3 class='channel-not-found'>Channel not valid</h3>"
			}

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

		log.Printf("feed channel:%s %s", channel, req.UserAgent())

		var channelId *int

		if len(channel) > 1 && len(channel) <= 25 {
			channelId = dbmanager.GetChannelId(channel, dbCon)

			if channelId == nil {
				http.Redirect(w, req, "/", http.StatusMovedPermanently)
				return
			}

			rss := dbmanager.GetRss(channelId, dbCon)
			if rss == nil {
				http.Redirect(w, req, "/", http.StatusMovedPermanently)
				return
			}
			w.Write([]byte(rss.Rss))
		}
	}

	http.HandleFunc("/", indexHandler)
	http.Handle("/static/", staticHandler)
	http.HandleFunc("/channel/", userHandler)
	http.HandleFunc("/feed/{channel}", feedHandler)

	fmt.Println("WebServer Running")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
