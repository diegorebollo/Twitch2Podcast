package jobs

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/ffmpeg"
	"log"
	"os"
	"strconv"
	"time"
)

func UpdateVods(db *sql.DB) {

	updateIntervalEnv := os.Getenv("UPDATE_INTERVAL")
	updateIntervalInt, err := strconv.Atoi(updateIntervalEnv)

	if err != nil {
		log.Fatal(err)
	}

	updateInterval := time.Duration(updateIntervalInt) * time.Minute

	log.Print("Update VODS Enable")

	for {
		time.Sleep(updateInterval)
		log.Print("Updating Vods...")
		channels := dbmanager.GetAllChannelIds(db)

		for i := 0; i < len(channels); i++ {
			channel := channels[i]
			dbmanager.UpdateAllVods(channel, db)
		}

	}

}

func EnableTranscoding(serverDbCon *sql.DB, usersDbCon *sql.DB) {

	log.Print("Transcoding enable")

	jobsQueue := dbmanager.GetJobsQueue(serverDbCon).Video
	if len(jobsQueue) > 0 {

		for _, job := range jobsQueue {
			dbmanager.InsertToTranscodeQueue(job, serverDbCon)
		}

		log.Println("Removing Jobs from Queue")
		dbmanager.RemoveAllFromJobsQueue(serverDbCon)
	}

	for {
		time.Sleep(5 * time.Second)
		transcodeQueue := dbmanager.GetTranscodeQueue(serverDbCon)
		if len(transcodeQueue.Video) >= 1 {
			ffmpeg.RunTranscodeQueue(serverDbCon, usersDbCon)
		}

	}
}
