package main

import (
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/jobs"
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("App Running")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbmanager.InitDb()
	usersDbCon := dbmanager.UsersDb()
	serverDbCon := dbmanager.ServerDb()

	go jobs.UpdateVods(usersDbCon)

	enableTranscoding := os.Getenv("ENABLE_TRANSCODE")

	if enableTranscoding == "true" {
		go jobs.EnableTranscoding(serverDbCon, usersDbCon)

	}

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))
	audioHandler := http.StripPrefix("/audios/", http.FileServer(http.Dir("audios/")))

	indexHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Print("index ", req.UserAgent())
		tmpl := template.Must(template.ParseFiles("templates/index.html"))

		tmpl.Execute(w, nil)
	}

	faqHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Print("faq ", req.UserAgent())
		tmpl := template.Must(template.ParseFiles("templates/faq.html"))

		tmpl.Execute(w, nil)
	}

	userHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			http.Redirect(w, req, "/", http.StatusMovedPermanently)
			return
		}

		channel := strings.ToLower(req.PostFormValue("channel"))
		channel = strings.TrimSpace(channel)

		if strings.Contains(channel, "twitch.tv/") {
			split := strings.Split(channel, "/")

			channel = split[len(split)-1]

			if len(channel) == 0 {
				channel = split[len(split)-2]
			}
		}

		var search models.Search

		if len(channel) > 1 && len(channel) <= 25 {
			search = dbmanager.SearchChannel(channel, usersDbCon)
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
			channelId = dbmanager.GetChannelId(channel, usersDbCon)

			if channelId == nil {
				http.Redirect(w, req, "/", http.StatusMovedPermanently)
				return
			}

			rss := dbmanager.GetRss(channelId, usersDbCon)
			if rss == nil {
				http.Redirect(w, req, "/", http.StatusMovedPermanently)
				return
			}
			w.Write([]byte(rss.Rss))
		}
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/faq", faqHandler)
	http.Handle("/static/{file}", staticHandler)
	http.Handle("/audios/{channelId}/{file}", audioHandler)
	http.HandleFunc("/channel/", userHandler)
	http.HandleFunc("/feed/{channel}", feedHandler)

	log.Println("WebServer Running")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))

}
