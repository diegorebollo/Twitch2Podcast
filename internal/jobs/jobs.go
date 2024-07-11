package jobs

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/ffmpeg"
	"fmt"
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

func EnableTranscoding(db *sql.DB) {

	log.Print("Transcoding enable")

	for {
		time.Sleep(5 * time.Second)
		transcodeQueue := dbmanager.GetTranscodeQueue(db)
		if len(transcodeQueue.Video) >= 1 {
			ffmpeg.RunTranscodeQueue(db)
			transcodeQueue = dbmanager.GetTranscodeQueue(db)
			fmt.Println("NO HAY MAS TRABAJO: ", len(transcodeQueue.Video))

		}

	}
}
