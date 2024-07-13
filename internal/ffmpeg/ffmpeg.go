package ffmpeg

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func transcodeComplete(vod *models.Video, serverDbCon *sql.DB, usersDbCon *sql.DB) {
	log.Printf("'%d'.mp3 Saved", vod.ID)
	dbmanager.RemoveFromJobsQueue(vod, serverDbCon)
	dbmanager.SetVodTranscodedTrue(vod, usersDbCon)

}

func RunTranscodeQueue(serverDbCon *sql.DB, usersDbCon *sql.DB) {
	maxJobsEnv := os.Getenv("MAX_TRANSCODE_JOBS")
	MaxJobs, err := strconv.Atoi(maxJobsEnv)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Running Transcode Queue")

	for {
		jobsToDo := dbmanager.GetTranscodeQueue(serverDbCon).Video

		if len(jobsToDo) < 1 {
			break
		}

		currentJobs := dbmanager.GetJobsQueue(serverDbCon).Video

		numOfJobsToDo := MaxJobs - len(currentJobs)

		if len(jobsToDo) < MaxJobs-len(currentJobs) {
			numOfJobsToDo = len(jobsToDo)
		}

		for i := range numOfJobsToDo {
			vod := jobsToDo[i]
			dbmanager.InsertToJobsQueue(vod, serverDbCon)
			saveMp3(vod, serverDbCon, usersDbCon)
			dbmanager.RemoveFromTranscodeQueue(vod, serverDbCon)
		}

		time.Sleep(5 * time.Second)
	}

}

func saveMp3(vod *models.Video, serverDbCon *sql.DB, usersDbCon *sql.DB) {

	userPath := fmt.Sprintf("audios/%d", vod.ChannelId)

	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err := os.MkdirAll(userPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	filePath := fmt.Sprintf("%s/%d.mp3", userPath, vod.ID)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error al eliminar el archivo:", err)
	}

	log.Printf("Transcoding VOD '%d'", vod.ID)

	cmd := exec.Command("ffmpeg", "-i", *vod.AudioURL, "-codec:a", "libmp3lame", "-qscale:a", "5", filePath)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		cmd.Wait()
		transcodeComplete(vod, serverDbCon, usersDbCon)
	}()
}
