package ffmpeg

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/dbmanager"
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"log"
	"os/exec"
	"time"
)

func removeElement(array []*models.Video, r *models.Video) []*models.Video {

	var newArray []*models.Video

	for _, e := range array {
		if e != r {
			newArray = append(newArray, e)
		}
	}
	return newArray
}

func transcodeComplete(vod *models.Video, db *sql.DB) {
	fmt.Println(vod.ID, ".mp3 Saved")
	dbmanager.RemoveFromJobsQueue(vod, db)

}

func RunTranscodeQueue(db *sql.DB) {

	fmt.Println("transcode queue")
	const MaxJobs = 4

	for {

		jobsToDo := dbmanager.GetTranscodeQueue(db).Video

		if len(jobsToDo) < 1 {
			break
		}

		currentJobs := dbmanager.GetJobsQueue(db).Video

		numOfJobsToDo := MaxJobs - len(currentJobs)

		if len(jobsToDo) < MaxJobs-len(currentJobs) {
			numOfJobsToDo = len(jobsToDo)
		}

		fmt.Println(numOfJobsToDo)

		for i := range numOfJobsToDo {
			vod := jobsToDo[i]
			dbmanager.InsertToJobsQueue(vod, db)
			saveMp3(vod, db)
			dbmanager.RemoveFromTranscodeQueue(vod, db)
		}

		time.Sleep(2 * time.Second)
	}

}

func saveMp3(vod *models.Video, db *sql.DB) {

	log.Printf("Transcoding VOD '%d'", vod.ID)

	filename := fmt.Sprintf("%d.mp3", vod.ID)

	cmd := exec.Command("ffmpeg", "-i", *vod.AudioURL, "-codec:a", "libmp3lame", "-qscale:a", "5", filename)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		cmd.Wait()
		transcodeComplete(vod, db)
	}()
}
