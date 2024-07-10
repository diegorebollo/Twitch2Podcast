package ffmpeg

import (
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"log"
	"os/exec"
	"time"
)

var transcodeQueue []*models.Video
var currentJobs []*models.Video

var IsTranscodeQueueRunning = false

func AddToTranscodeQueue(vod *models.Video) {
	transcodeQueue = append(transcodeQueue, vod)
}

func removeElement(array []*models.Video, r *models.Video) []*models.Video {

	var newArray []*models.Video

	for _, e := range array {
		if e != r {
			newArray = append(newArray, e)
		}
	}
	return newArray
}

func transcodeComplete(vod *models.Video) {
	log.Printf("%d.mp3 saved", vod.ID)
	currentJobs = removeElement(currentJobs, vod)
}

func RunTranscodeQueue() {

	fmt.Println("transcode queue")
	IsTranscodeQueueRunning = true
	const MaxJobs = 4

	for len(transcodeQueue) >= 1 {

		fmt.Println("Current num of jobs:", len(currentJobs))

		var numJobs int

		if len(transcodeQueue) < MaxJobs-len(currentJobs) {
			numJobs = len(transcodeQueue)
		} else {
			numJobs = MaxJobs - len(currentJobs)
		}

		for i := range numJobs {
			vod := transcodeQueue[i]
			currentJobs = append(currentJobs, vod)
			transcodeQueue = removeElement(transcodeQueue, vod)
			saveMp3(vod)
		}

		time.Sleep(5 * time.Second)

	}

	IsTranscodeQueueRunning = false

}

func saveMp3(vod *models.Video) {

	log.Printf("Transcoding VOD '%d'", vod.ID)

	filename := fmt.Sprintf("%d.mp3", vod.ID)

	cmd := exec.Command("ffmpeg", "-i", *vod.AudioURL, "-codec:a", "libmp3lame", "-qscale:a", "5", filename)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		cmd.Wait()
		transcodeComplete(vod)
	}()
}
