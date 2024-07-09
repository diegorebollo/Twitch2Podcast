package jobs

import (
	"database/sql"
	"drebollo/twitchtopodcast/internal/dbmanager"
	"log"
	"time"
)

func UpdateVods(db *sql.DB) {

	log.Print("Update VODS Enable")

	for {
		time.Sleep(10 * time.Second)
		log.Print("Updating Vods...")
		channels := dbmanager.GetAllChannelIds(db)

		for i := 0; i < len(channels); i++ {
			channel := channels[i]
			dbmanager.UpdateAllVods(channel, db)
		}

	}

}
