package models

import (
	"time"
)

const DateLayout = "2006-01-02 15:04:05.999999 -0700"

type Search struct {
	Channel *Channel
}

type Channel struct {
	Id              int
	Login           string
	DisplayName     string
	Description     string
	CreatedAt       time.Time
	LastSearch      time.Time
	ProfileImageURL string
}

type Video struct {
	ID                  int
	ChannelId           int
	Title               string
	Description         string
	Language            string
	CreatedAt           time.Time
	LengthSeconds       int
	BroadcastType       string
	AudioURL            *string
	PreviewThumbnailURL string
	IsPublic            bool
}

type Rss struct {
	ChannelId  int
	Rss        string
	LastUpdate time.Time
	LastSearch time.Time
}

type Episode struct {
	ChannelId int
	VideoId   int
	Language  string
	Data      string
}

type ApiVideos struct {
	Edges []ApiEdges `json:"edges"`
}

type ApiEdges struct {
	Node ApiNode `json:"node"`
}

type ApiNode struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Language            string    `json:"language"`
	CreatedAt           time.Time `json:"createdAt"`
	LengthSeconds       int       `json:"lengthSeconds"`
	BroadcastType       string    `json:"broadcastType"`
	PreviewThumbnailURL string    `json:"previewThumbnailURL"`
}

type ApiUser struct {
	ID              string     `json:"id"`
	Login           *string    `json:"login"`
	DisplayName     *string    `json:"displayName"`
	Description     *string    `json:"description"`
	CreatedAt       *time.Time `json:"createdAt"`
	ProfileImageURL *string    `json:"profileImageURL"`
	Videos          *ApiVideos `json:"videos"`
}

type ApiVideoPlaybackAccessToken struct {
	Signature string `json:"signature"`
	Value     string `json:"value"`
}

type ApiData struct {
	User                     *ApiUser                     `json:"user"`
	VideoPlaybackAccessToken *ApiVideoPlaybackAccessToken `json:"videoPlaybackAccessToken"`
}

type ApiExtensions struct {
	DurationMilliseconds int    `json:"durationMilliseconds"`
	RequestID            string `json:"requestID"`
}

type ApiResponse struct {
	Data       ApiData       `json:"data"`
	Extensions ApiExtensions `json:"extensions"`
}
