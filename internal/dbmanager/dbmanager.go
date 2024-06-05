package dbmanager

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/models"
	"drebollo/twitchtopodcast/internal/rss"
	twichapi "drebollo/twitchtopodcast/internal/twitchapi"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func ConnectDb() *sql.DB {
	db, err := sql.Open("sqlite3", "file:db.sqlite")

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func InitDb(db *sql.DB) {

	query := `CREATE TABLE IF NOT EXISTS channel (
		id INTEGER PRIMARY KEY,
		login TEXT,
		displayName TEXT,
		description TEXT,
		createdAt TEXT,
		lastSearch TEXT
	)`
	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

	query = `CREATE TABLE IF NOT EXISTS vod (
		id INTEGER PRIMARY KEY,	
		channelId INTEGER,			
		title TEXT,
		description TEXT,
		language TEXT,
		createdAt TEXT,
		lengthSeconds INTEGER,
		broadcastType TEXT,
		audioURL TEXT,
		previewThumbnailURL TEXT,
		isPublic BOOL, 
		FOREIGN KEY(channelId) REFERENCES channel(id)
	)`

	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

	query = `CREATE TABLE IF NOT EXISTS rss (
		channelId INTEGER,	
		rss BLOB,
		lastUpdate TEXT,
		lastSearch TEXT,
		FOREIGN KEY(channelId) REFERENCES channel(id)
	)`
	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

	query = `CREATE TABLE IF NOT EXISTS episode (
		channelId INTEGER,	
		videoId INTEGER,
		data BLOB,
		FOREIGN KEY(channelId) REFERENCES channel(id)
	)`
	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

}

func SearchChannel(loginChannel string, db *sql.DB) models.Search {

	var channel models.Channel
	err := db.QueryRow("SELECT * FROM channel WHERE login = $1", loginChannel).Scan(&channel.Id, &channel.Login, &channel.DisplayName, &channel.Description, &channel.CreatedAt, &channel.LastSearch)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Channel '%s' NOT found in DB", loginChannel)
			apiLookup := twichapi.ChannelData(loginChannel)
			if apiLookup.User == nil {
				log.Printf("Channel '%s' NOT found on Twitch", loginChannel)
				search := models.Search{Channel: nil}
				return search
			}
			channel = insertChannel(*apiLookup.User, ConnectDb())
			search := models.Search{Channel: &channel}
			return search
		}
		log.Fatal(err)
	}

	currentTime := time.Time.String(time.Now())
	_, err = db.Exec(`UPDATE channel SET lastSearch = $1 WHERE id = $2`, currentTime, channel.Id)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	search := models.Search{Channel: &channel}
	return search
}

func insertChannel(user models.ApiUser, db *sql.DB) models.Channel {

	createdAt := time.Time.String(*user.CreatedAt)
	lastSearch := time.Time.String(time.Now())

	var channel models.Channel

	var description string
	if user.Description == nil {
		description = fmt.Sprintf("%s's Twitch Channel", *user.DisplayName)
	} else {
		description = *user.Description
	}

	query := `INSERT INTO channel (id, login, displayName, description, createdAt, lastSearch)
				VALUES ($1, $2, $3, $4, $5, $6)	RETURNING *`

	err := db.QueryRow(query, user.ID, user.Login, user.DisplayName, description, createdAt, lastSearch).Scan(&channel.Id, &channel.Login, &channel.DisplayName, &channel.Description, &channel.CreatedAt, &channel.LastSearch)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	go saveAllVods(channel.Id)

	log.Printf("Channel '%s' CREATED in DB", *user.Login)
	return channel
}

func GetChannelId(loginChannel string, db *sql.DB) *int {

	var channelId *int
	err := db.QueryRow("SELECT id FROM channel WHERE login = $1", loginChannel).Scan(&channelId)

	if err == sql.ErrNoRows {
		return nil
	}

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	return channelId
}

func saveAllVods(channelId int) {

	numVodSave := 0
	vods := twichapi.GetAllVodsFromChannel(channelId)

	if len(vods) == 0 {
		log.Printf("Channel ID '%d' has NO VODs", channelId)
		return
	}

	for i := 0; i < len(vods); i++ {
		vod := vods[i]
		vodId, err := strconv.Atoi(vod.Node.ID)

		if err != nil {
			log.Fatal(err)
		}

		if !vodExist(vodId, ConnectDb()) {
			if !strings.Contains(vod.Node.PreviewThumbnailURL, "404_processing") {
				go insertVod(channelId, vod, ConnectDb())
				numVodSave++
			}
		}
	}
	// go insertRss(channelId)
	log.Printf("Total VODs saved from Channel ID '%d': %d ", channelId, numVodSave)

}

func insertVod(channelId int, vod models.ApiEdges, db *sql.DB) models.Video {

	vodId, err := strconv.Atoi(vod.Node.ID)

	if err != nil {
		log.Fatal(err)
	}

	audioUrl := twichapi.GetAudioUrl(vodId)

	var isPublic bool

	if audioUrl == nil {
		isPublic = false
	} else {
		isPublic = true
	}

	var video models.Video

	query := `INSERT INTO vod (id,channelId,title,description,language,createdAt,lengthSeconds,broadcastType, audioURL, previewThumbnailURL, isPublic)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *`

	err = db.QueryRow(query, vodId, channelId, vod.Node.Title, vod.Node.Description, vod.Node.Language, time.Time.String(vod.Node.CreatedAt), vod.Node.LengthSeconds, vod.Node.BroadcastType, audioUrl, vod.Node.PreviewThumbnailURL, isPublic).Scan(&video.ID, &video.ChannelId, &video.Title, &video.Description, &video.Language, &video.CreatedAt, &video.LengthSeconds, &video.BroadcastType, &video.AudioURL, &video.PreviewThumbnailURL, &video.IsPublic)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	go insertEpisode(channelId, video, ConnectDb())
	return video
}

func vodExist(vodId int, db *sql.DB) bool {

	var videoId int
	err := db.QueryRow("SELECT id FROM vod WHERE id = $1", vodId).Scan(&videoId)

	if err == sql.ErrNoRows {
		return false
	}

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	return true
}

func GetRss(channelId *int, db *sql.DB) *models.Rss {
	var rss models.Rss
	err := db.QueryRow("SELECT * FROM rss WHERE channelId = $1", channelId).Scan(&rss.ChannelId, &rss.Rss, &rss.LastUpdate, &rss.LastSearch)

	if err == sql.ErrNoRows {
		return nil
	}

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	return &rss

}

func InsertRss(channelId int) {

	episodes := getAllEpisodes(channelId, ConnectDb())

	if len(episodes) == 0 {
		log.Printf("ChannelId: '%d' has NOT any Episodes", channelId)
		return
	}

	for i := 0; i < len(episodes); i++ {
		fmt.Println(episodes[i].Data)
	}

	//  CONTINUAR AQUI... Parsea el json de epsiodes.data

}

func insertEpisode(channelId int, vod models.Video, db *sql.DB) {

	episode := rss.GenerateEpisode(vod)

	jsonData, err := json.Marshal(episode)

	if err != nil {
		fmt.Println(err)
		return
	}

	query := `INSERT INTO episode (channelId, videoId, data)
	VALUES (?, ?, ?)`

	result, err := db.Exec(query, channelId, vod.ID, string(jsonData))

	if err != nil {
		log.Fatal(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows != 1 {
		log.Fatalf("expected to affect 1 row, affected %d", rows)
	}

	defer db.Close()

}

func getAllEpisodes(channelId int, db *sql.DB) []models.Episode {

	rows, err := db.Query("SELECT * FROM episode WHERE channelId = ?", channelId)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var episodes []models.Episode

	for rows.Next() {
		var episode models.Episode
		err := rows.Scan(&episode.ChannelId, &episode.VideoId, &episode.Data)

		if err != nil {
			log.Fatal(err)
		}

		episodes = append(episodes, episode)
	}
	return episodes
}
