package jobs

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/ffmpeg"
	"log"
	"time"
)

func UpdateVods(db *sql.DB) {

	log.Print("Update VODS Enable")

	for {
		time.Sleep(1 * time.Minute)
		log.Print("Updating Vods...")
		channels := dbmanager.GetAllChannelIds(db)

		for i := 0; i < len(channels); i++ {
			channel := channels[i]
			dbmanager.UpdateAllVods(channel, db)
		}

	}

}

func EnableTranscoding() {

	log.Print("Transcoding enable")

	for {
		time.Sleep(10 * time.Second)
		if !ffmpeg.IsTranscodeQueueRunning {
			ffmpeg.RunTranscodeQueue()
		}
	}
}
