package rss

import (
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"time"

	"github.com/eduncan911/podcast"
)

func Generator(channel models.Channel) podcast.Podcast {

	title := channel.DisplayName
	link := fmt.Sprintf("https://www.twitch.tv/%s", channel.Login)
	description := channel.Description
	lastBuildDate := time.Now()

	const dateLayout = "2006-01-02 15:04:05.999999 -0700 MST"

	pubDate, err := time.Parse(dateLayout, channel.CreatedAt)
	if err != nil {
		panic(err)
	}

	podcast := podcast.New(title, link, description, &pubDate, &lastBuildDate)
	return podcast
}

func GenerateEpisode(vod models.Video) podcast.Item {
	episode := podcast.Item{Title: vod.Title, Description: vod.Description, Link: vod.AudioURL}
	return episode
}
