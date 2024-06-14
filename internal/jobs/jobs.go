package jobs

import (
	"drebollo/twitchtopodcast/internal/dbmanager"
	"log"
	"time"
)

func UpdateVods() {

	log.Print("Update VODS Enable")

	for {
		time.Sleep(5 * time.Minute)
		log.Print("Updating Vods...")
		channels := dbmanager.GetAllChannelIds(dbmanager.ConnectDb())

		for i := 0; i < len(channels); i++ {
			channel := channels[i]
			dbmanager.UpdateAllVods(channel)
		}

	}

}
