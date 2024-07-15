package rss

import (
	"drebollo/twitchtopodcast/internal/models"
	"fmt"
	"strconv"
	"time"

	"github.com/eduncan911/podcast"
)

func Generator(channel models.Channel) podcast.Podcast {

	title := channel.DisplayName
	link := fmt.Sprintf("https://www.twitch.tv/%s", channel.Login)
	description := channel.Description
	lastBuildDate := time.Now()
	img := channel.ProfileImageURL

	const dateLayout = "2006-01-02 15:04:05.999999 -0700 MST"
	pubDate, err := time.Parse(dateLayout, channel.CreatedAt)

	if err != nil {
		panic(err)
	}

	podcast := podcast.New(title, link, description, &pubDate, &lastBuildDate)
	podcast.AddAuthor(channel.DisplayName, link)
	podcast.AddImage(img)
	podcast.Generator = "twitch2podcast.com | go podcast v1.3.1 (github.com/eduncan911/podcast)"
	return podcast
}

func GenerateEpisode(vod models.Video) *podcast.Item {

	vodLink := fmt.Sprintf("https://www.twitch.tv/videos/%d", vod.ID)

	if vod.IsPublic {
		episode := podcast.Item{Title: vod.Title, Description: "default description", Link: vodLink}

		const dateLayout = "2006-01-02 15:04:05.999999 -0700 MST"

		pubDate, err := time.Parse(dateLayout, vod.CreatedAt)

		if err != nil {
			panic(err)
		}

		audioUrl := *vod.AudioURL

		if vod.IsTranscoded {
			audioUrl = fmt.Sprintf("https://dev.twitch2podcast.com/audios/%d/%d.mp3", vod.ChannelId, vod.ID)
		}

		episode.AddPubDate(&pubDate)

		size := (vod.LengthSeconds * 100000) / 8

		episode.AddEnclosure(audioUrl, podcast.MP3, int64(size))
		episode.IDuration = strconv.Itoa(vod.LengthSeconds)
		episode.Description = fmt.Sprintf("<h1>%s</h1><img src='%s'><a href='%s'>%s</a><br><p>Podcast generated by <a href='https://www.twitch2podcast.com'>twitch2podcast.com</a></p>", vod.Title, vod.PreviewThumbnailURL, vodLink, vodLink)
		return &episode
	}
	return nil
}
