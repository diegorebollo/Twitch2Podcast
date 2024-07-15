package dbmanager

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/models"
	"drebollo/twitchtopodcast/internal/rss"
	twichapi "drebollo/twitchtopodcast/internal/twitchapi"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eduncan911/podcast"
	_ "github.com/mattn/go-sqlite3"
)

func UsersDb() *sql.DB {
	db, err := sql.Open("sqlite3", "file:db/users.sqlite?temp_store=memory&_journal_mode=WAL&_synchronous=normal&_mmap_size=30000000000&_busy_timeout=30000")

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func ServerDb() *sql.DB {
	db, err := sql.Open("sqlite3", "file:db/server.sqlite?temp_store=memory&_journal_mode=WAL&_synchronous=normal&_mmap_size=30000000000&_busy_timeout=30000")

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func initUsersDb() {

	db := UsersDb()

	defer db.Close()

	query := `CREATE TABLE IF NOT EXISTS channel (
		id INTEGER PRIMARY KEY,
		login TEXT,
		displayName TEXT,
		description TEXT,
		createdAt TEXT,
		lastSearch TEXT,
		profileImageURL TEXT
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
		isTranscoded BOOL,
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
		videoId INTEGER PRIMARY KEY,
		language  TEXT,
		data BLOB,
		FOREIGN KEY(channelId) REFERENCES channel(id)
	)`
	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

}

func initServerDb() {

	db := ServerDb()
	defer db.Close()

	query := `CREATE TABLE IF NOT EXISTS transcode_queue (
		videoId INTEGER PRIMARY KEY,
		channelId INTEGER,
		createdAt TEXT,
		video TEXT		
	)`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

	query = `CREATE TABLE IF NOT EXISTS job_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		videoId INTEGER,
		channelId INTEGER,	
		video TEXT		
	)`

	_, err = db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

}

func InitDb() {

	folderPath := "db"

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	initUsersDb()
	initServerDb()

}

func SearchChannel(loginChannel string, db *sql.DB) models.Search {

	var channel models.Channel
	err := db.QueryRow("SELECT * FROM channel WHERE login = $1", loginChannel).Scan(&channel.Id, &channel.Login, &channel.DisplayName, &channel.Description, &channel.CreatedAt, &channel.LastSearch, &channel.ProfileImageURL)

	baseUrl := os.Getenv("BASE_URL")

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Channel '%s' NOT found in DB", loginChannel)
			apiLookup := twichapi.ChannelData(loginChannel)
			if apiLookup.User == nil {
				log.Printf("Channel '%s' NOT found on Twitch", loginChannel)
				search := models.Search{Channel: nil}
				return search
			}
			channel = insertChannel(*apiLookup.User, db)
			urlFeed := fmt.Sprintf("%s/feed/%s", baseUrl, channel.Login)
			search := models.Search{Channel: &channel, Url: &urlFeed}
			return search
		}
		log.Fatal(err)
	}

	currentTime := time.Time.String(time.Now())
	_, err = db.Exec(`UPDATE channel SET lastSearch = $1 WHERE id = $2`, currentTime, channel.Id)

	if err != nil {
		log.Fatal(err)
	}

	urlFeed := fmt.Sprintf("%s/feed/%s", baseUrl, channel.Login)
	search := models.Search{Channel: &channel, Url: &urlFeed}

	return search
}

func getChannel(channelId int, db *sql.DB) *models.Channel {

	var channel models.Channel
	err := db.QueryRow("SELECT * FROM channel WHERE id = $1", channelId).Scan(&channel.Id, &channel.Login, &channel.DisplayName, &channel.Description, &channel.CreatedAt, &channel.LastSearch, &channel.ProfileImageURL)

	if err == sql.ErrNoRows {
		log.Printf("Channel ID '%d' do NOT exist", channelId)
		return nil
	}

	if err != nil {
		log.Fatal(err)
	}
	return &channel

}

func GetAllChannelIds(db *sql.DB) []int {

	rows, err := db.Query("SELECT id FROM channel")

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var channels []int

	for rows.Next() {
		var channel int
		err := rows.Scan(&channel)

		if err != nil {
			log.Fatal(err)
		}

		channels = append(channels, channel)
	}
	return channels

}

func insertChannel(user models.ApiUser, db *sql.DB) models.Channel {

	createdAt := time.Time.String(*user.CreatedAt)
	lastSearch := time.Time.String(time.Now())

	var channel models.Channel

	var description string
	if user.Description == nil || len(*user.Description) == 0 {
		description = fmt.Sprintf("%s's Twitch Channel", *user.DisplayName)
	} else {
		description = *user.Description
	}

	query := `INSERT INTO channel (id, login, displayName, description, createdAt, lastSearch, profileImageURL)
				VALUES ($1, $2, $3, $4, $5, $6, $7)	RETURNING *`

	err := db.QueryRow(query, user.ID, user.Login, user.DisplayName, description, createdAt, lastSearch, user.ProfileImageURL).Scan(&channel.Id, &channel.Login, &channel.DisplayName, &channel.Description, &channel.CreatedAt, &channel.LastSearch, &channel.ProfileImageURL)

	if err != nil {
		log.Fatal(err)
	}

	go insertRss(channel.Id, db)
	go saveAllVods(channel.Id, db)

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
	return channelId
}

func saveAllVods(channelId int, db *sql.DB) {

	numVodSave := 0
	vods := twichapi.GetAllVodsFromChannel(channelId)

	if len(vods) == 0 {
		log.Printf("Channel ID '%d' has NO VODs", channelId)
	}

	var wg sync.WaitGroup

	for _, vod := range vods {
		wg.Add(1)
		go insertVod(channelId, vod, db, &wg)
		numVodSave++

	}

	go func() {
		wg.Wait()
		updateRss(channelId, db)
		log.Printf("Total VODs saved from Channel ID '%d': %d ", channelId, numVodSave)
	}()
}

func insertVod(channelId int, vod models.ApiEdges, db *sql.DB, wg *sync.WaitGroup) *models.Video {

	defer wg.Done()

	if strings.Contains(vod.Node.PreviewThumbnailURL, "404_processing") {
		log.Printf("VOD '%s' is still a livestream", vod.Node.ID)
		return nil
	}

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

	isTranscoded := false

	var video models.Video

	query := `INSERT INTO vod (id,channelId,title,description,language,createdAt,lengthSeconds,broadcastType, audioURL, previewThumbnailURL, isPublic, isTranscoded)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *`

	err = db.QueryRow(query, vodId, channelId, vod.Node.Title, vod.Node.Description, vod.Node.Language, time.Time.String(vod.Node.CreatedAt), vod.Node.LengthSeconds, vod.Node.BroadcastType, audioUrl, vod.Node.PreviewThumbnailURL, isPublic, isTranscoded).Scan(&video.ID, &video.ChannelId, &video.Title, &video.Description, &video.Language, &video.CreatedAt, &video.LengthSeconds, &video.BroadcastType, &video.AudioURL, &video.PreviewThumbnailURL, &video.IsPublic, &video.IsTranscoded)

	if err != nil {
		log.Fatal(err)

	}

	if video.IsPublic {
		insertEpisode(video, db)
		InsertToTranscodeQueue(&video, ServerDb())
	}

	return &video
}

func SetVodTranscodedTrue(vod *models.Video, db *sql.DB) {

	var updatedVod models.Video

	err := db.QueryRow(`UPDATE vod SET isTranscoded = $1 WHERE id = $2 RETURNING *`, true, vod.ID).Scan(&updatedVod.ID, &updatedVod.ChannelId, &updatedVod.Title, &updatedVod.Description, &updatedVod.Language, &updatedVod.CreatedAt, &updatedVod.LengthSeconds, &updatedVod.BroadcastType, &updatedVod.AudioURL, &updatedVod.PreviewThumbnailURL, &updatedVod.IsPublic, &updatedVod.IsTranscoded)

	if err == sql.ErrNoRows {
		log.Fatal("ww", err)

	}

	if err != nil {
		log.Fatal(err)
	}

	updateEpisode(&updatedVod, db)
	updateRss(vod.ChannelId, db)
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

	return true
}

func GetVod(vodId int, db *sql.DB) *models.Video {

	var video models.Video
	err := db.QueryRow("SELECT * FROM vod WHERE id = $1", vodId).Scan(&video.ID, &video.ChannelId, &video.Title, &video.Description, &video.Language, &video.CreatedAt, &video.LengthSeconds, &video.BroadcastType, &video.AudioURL, &video.PreviewThumbnailURL, &video.IsPublic, &video.IsTranscoded)

	if err == sql.ErrNoRows {
		return nil
	}

	if err != nil {
		log.Fatal(err)
	}

	return &video

}

func getAllVods(channelId int, db *sql.DB) []models.Video {

	rows, err := db.Query("SELECT * FROM vod WHERE channelId = ? ORDER BY createdAt DESC", channelId)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var vods []models.Video

	for rows.Next() {
		var video models.Video
		err := rows.Scan(&video.ID, &video.ChannelId, &video.Title, &video.Description, &video.Language, &video.CreatedAt, &video.LengthSeconds, &video.BroadcastType, &video.AudioURL, &video.PreviewThumbnailURL, &video.IsPublic, &video.IsTranscoded)

		if err != nil {
			log.Fatal(err)
		}

		vods = append(vods, video)
	}

	return vods
}

func getAllVodsId(channelId *int, db *sql.DB) []int {

	rows, err := db.Query("SELECT id FROM vod WHERE channelId = ? ORDER BY createdAt DESC", channelId)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var vodsId []int

	for rows.Next() {
		var videoId int
		err := rows.Scan(&videoId)

		if err != nil {
			log.Fatal(err)
		}

		vodsId = append(vodsId, videoId)
	}

	return vodsId

}

func removeVod(channelId int, vodId int, db *sql.DB) {

	query := `DELETE FROM vod WHERE	id = ?`

	result, err := db.Exec(query, vodId)

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

	removeEpisode(channelId, vodId, db)

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

	return &rss

}

func insertRss(channelId int, db *sql.DB) {

	channel := getChannel(channelId, db)
	rss := rss.Generator(*channel)

	query := `INSERT INTO rss (channelId, rss, lastUpdate, lastSearch)
	VALUES (?, ?, ?, ?)`

	time := time.Now()

	result, err := db.Exec(query, channelId, rss.String(), time, time)

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

}

func insertEpisode(vod models.Video, db *sql.DB) {

	episode := rss.GenerateEpisode(vod)
	jsonData, err := json.Marshal(episode)

	if err != nil {
		fmt.Println(err)
		return
	}

	query := `INSERT INTO episode (channelId, videoId, language, data)
	VALUES (?, ?, ?, ?)`

	result, err := db.Exec(query, vod.ChannelId, vod.ID, vod.Language, string(jsonData))

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
}

func updateEpisode(vod *models.Video, db *sql.DB) {

	episode := rss.GenerateEpisode(*vod)
	jsonData, err := json.Marshal(episode)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec(`UPDATE episode SET data = $1 WHERE videoId = $2`, jsonData, vod.ID)
	if err == sql.ErrNoRows {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

}

func updateRss(channelId int, db *sql.DB) {

	channel := getChannel(channelId, db)
	rssData := rss.Generator(*channel)
	episodes := getAllEpisodes(channelId, db)

	if len(episodes) == 0 {
		log.Printf("ChannelId: '%d' has NOT any Episodes", channelId)

		time := time.Now()
		_, err := db.Exec(`UPDATE rss SET rss = $1, lastUpdate = $2, lastSearch = $3 WHERE channelId = $4`, rssData.String(), time, time, channel.Id)
		if err == sql.ErrNoRows {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	for i := len(episodes) - 1; i >= 0; i-- {
		episodeData := episodes[i].Data
		rssData.Language = episodes[i].Language

		var episode podcast.Item

		err := json.Unmarshal([]byte(episodeData), &episode)

		if err != nil {
			fmt.Println(err)
		}

		if _, err := rssData.AddItem(episode); err != nil {
			fmt.Println(episode.Title, ": error", err.Error())
			return
		}
	}

	time := time.Now()

	_, err := db.Exec(`UPDATE rss SET rss = $1, lastUpdate = $2, lastSearch = $3 WHERE channelId = $4`, rssData.String(), time, time, channel.Id)

	if err == sql.ErrNoRows {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

}

func removeEpisode(channelId int, vodId int, db *sql.DB) {

	query := `DELETE FROM episode WHERE	videoId = ?`

	result, err := db.Exec(query, vodId)

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

	// filePath := fmt.Sprintf("audios/%d/%d.mp3", channelId, vodId)

	// if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
	// 	fmt.Println("Error al eliminar el archivo:", err)
	// }

	updateRss(channelId, db)

}

func removeAllEpisodes(channelId int, db *sql.DB) {

	query := `DELETE FROM episode WHERE channelId = ?`

	result, err := db.Exec(query, channelId)

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

	updateRss(channelId, db)
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
		err := rows.Scan(&episode.ChannelId, &episode.VideoId, &episode.Language, &episode.Data)

		if err != nil {
			log.Fatal(err)
		}

		episodes = append(episodes, episode)
	}
	return episodes
}

func UpdateAllVods(channelId int, db *sql.DB) {

	needUpdate := false

	vodsApi := twichapi.GetAllVodsFromChannel(channelId)
	vodsDb := getAllVodsId(&channelId, db)

	if len(vodsApi) == 0 && len(vodsDb) == 0 {
		return
	}

	var vodsApiId []int

	var wg sync.WaitGroup

	// Check if Vod from Api is in DB

	if len(vodsApi) == 0 {
		log.Printf("Channel ID '%d' has NO VODs (API)", channelId)
	}

	for i := len(vodsApi) - 1; i >= 0; i-- {
		lastvodApi, err := strconv.Atoi(vodsApi[i].Node.ID)
		vodsApiId = append(vodsApiId, lastvodApi)

		if err != nil {
			log.Fatal(err)
		}

		isNew := true

		for _, id := range vodsDb {
			if id == lastvodApi {
				isNew = false
				break
			}
		}

		wg.Add(1)

		if isNew {
			needUpdate = true
			vod := insertVod(channelId, vodsApi[i], db, &wg)
			if vod != nil {
				log.Printf("New VOD (%d) from Channel ID '%d' added", lastvodApi, channelId)
			}
		}

		go func() {
			wg.Wait()
		}()
	}

	// Check if Vod from DB is in API

	for i := len(vodsDb) - 1; i >= 0; i-- {

		vodDb := vodsDb[i]

		stillOnTwitch := false

		for _, id := range vodsApiId {
			if id == vodDb {
				stillOnTwitch = true
				break
			}
		}

		if !stillOnTwitch {
			needUpdate = true
			removeVod(channelId, vodDb, db)
			log.Printf("VOD %d deleted", vodDb)
		}
	}

	if needUpdate {
		updateRss(channelId, db)
	}
}

func InsertToTranscodeQueue(vod *models.Video, db *sql.DB) {

	data, err := json.Marshal(vod)

	if err != nil {
		log.Fatal(err)
	}

	query := `INSERT INTO transcode_queue (channelId, videoId, createdAt, video)
	VALUES (?, ?, ?, ?)`

	result, err := db.Exec(query, vod.ChannelId, vod.ID, vod.CreatedAt, string(data))

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

}

func GetTranscodeQueue(db *sql.DB) models.TranscodeQueue {

	rows, err := db.Query("SELECT * FROM transcode_queue ORDER BY createdAt DESC")

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var transcodeQueue models.TranscodeQueue

	for rows.Next() {
		var vodDb models.TranscodeQueueDb
		err := rows.Scan(&vodDb.VideoId, &vodDb.ChannelId, &vodDb.CreatedAt, &vodDb.Video)

		if err != nil {
			log.Fatal(err)
		}

		var vod models.Video

		err = json.Unmarshal([]byte(vodDb.Video), &vod)

		if err != nil {
			fmt.Println(err)
		}

		transcodeQueue.Video = append(transcodeQueue.Video, &vod)
	}
	return transcodeQueue

}

func RemoveFromTranscodeQueue(vod *models.Video, db *sql.DB) {

	query := `DELETE FROM transcode_queue WHERE	videoId = ?`

	result, err := db.Exec(query, vod.ID)

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

}

func InsertToJobsQueue(vod *models.Video, db *sql.DB) {

	data, err := json.Marshal(vod)

	if err != nil {
		log.Fatal(err)
	}

	query := `INSERT INTO job_queue (channelId, videoId, video)
	VALUES (?, ?, ?)`

	result, err := db.Exec(query, vod.ChannelId, vod.ID, string(data))

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

}

func GetJobsQueue(db *sql.DB) models.TranscodeQueue {

	rows, err := db.Query("SELECT * FROM job_queue ORDER BY id ASC")

	if err != nil {
		log.Fatal(err)
	}

	var transcodeQueue models.TranscodeQueue

	for rows.Next() {
		var vodDb models.TranscodeQueueDb
		err := rows.Scan(&vodDb.Id, &vodDb.VideoId, &vodDb.ChannelId, &vodDb.Video)

		if err != nil {
			log.Fatal(err)
		}

		var vod models.Video

		err = json.Unmarshal([]byte(vodDb.Video), &vod)

		if err != nil {
			fmt.Println(err)
		}

		transcodeQueue.Video = append(transcodeQueue.Video, &vod)
	}
	return transcodeQueue

}

func RemoveFromJobsQueue(vod *models.Video, db *sql.DB) {

	query := `DELETE FROM job_queue WHERE videoId = ?`

	result, err := db.Exec(query, vod.ID)

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

}

func RemoveAllFromJobsQueue(db *sql.DB) {

	query := `DELETE FROM job_queue`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}

}
