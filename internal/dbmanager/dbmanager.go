package dbmanager

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/models"
	twichapi "drebollo/twitchtopodcast/internal/twitchapi"
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
		channelId INTEGER,			
		id INTEGER PRIMARY KEY,
		title TEXT,
		description TEXT,
		language TEXT,
		createdAt TEXT,
		lengthSeconds INTEGER,
		broadcastType TEXT,
		previewThumbnailURL TEXT,
		audioURL TEXT,
		FOREIGN KEY(channelId) REFERENCES channel(id)
	)`

	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

	query = `CREATE TABLE IF NOT EXISTS rss (
		channelId INTEGER,	
		rss TEXT,
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

	go saveAllVods(channel.Id)

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

	query := `INSERT INTO channel (id, login, displayName, description, createdAt, lastSearch)
				VALUES ($1, $2, $3, $4, $5, $6)	RETURNING *`

	err := db.QueryRow(query, user.ID, user.Login, user.DisplayName, user.Description, createdAt, lastSearch).Scan(&channel.Id, &channel.Login, &channel.DisplayName, &channel.Description, &channel.CreatedAt, &channel.LastSearch)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	go saveAllVods(channel.Id)

	log.Printf("Channel '%s' CREATED in DB", *user.Login)
	return channel
}

func ChannelId(loginChannel string, db *sql.DB) *int {

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
	log.Printf("Total VODs saved from Channel ID '%d': %d ", channelId, numVodSave)
}

func insertVod(channelId int, vod models.ApiEdges, db *sql.DB) {

	vodId, err := strconv.Atoi(vod.Node.ID)

	log.Printf("Inserting VOD '%d' from Channel '%d'", vodId, channelId)

	if err != nil {
		log.Fatal(err)
	}

	audioUrl := twichapi.GetAudioUrl(vodId)

	query := `INSERT INTO vod (channelId,id,title,description,language,createdAt,lengthSeconds,broadcastType, audioURL, previewThumbnailURL)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, channelId, vodId, vod.Node.Title, vod.Node.Description, vod.Node.Language, time.Time.String(vod.Node.CreatedAt), vod.Node.LengthSeconds, vod.Node.BroadcastType, audioUrl, vod.Node.PreviewThumbnailURL)

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
