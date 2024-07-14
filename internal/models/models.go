package models

import (
	"time"
)

type Search struct {
	Channel *Channel
	Url     *string
}

type Channel struct {
	Id              int
	Login           string
	DisplayName     string
	Description     string
	CreatedAt       string
	LastSearch      string
	ProfileImageURL string
}

type Video struct {
	ID                  int
	ChannelId           int
	Title               string
	Description         string
	Language            string
	CreatedAt           string
	LengthSeconds       int
	BroadcastType       string
	AudioURL            *string
	PreviewThumbnailURL string
	IsPublic            bool
	IsTranscoded        bool
}

type Rss struct {
	ChannelId  int
	Rss        string
	LastUpdate string
	LastSearch string
}

type Episode struct {
	ChannelId int
	VideoId   int
	Language  string
	Data      string
}

type TranscodeQueueDb struct {
	Id        int
	VideoId   int
	ChannelId int
	CreatedAt string
	Video     string
}

type TranscodeQueue struct {
	Video []*Video
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
