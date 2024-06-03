package models

import "time"

type Search struct {
	Channel *Channel
}

type Channel struct {
	Id          int
	Login       string
	DisplayName string
	Description *string
	CreatedAt   string
	LastSearch  string
}

type Video struct {
	ChannelId           int
	ID                  int
	Title               string
	Description         string
	Language            string
	CreatedAt           string
	LengthSeconds       int
	BroadcastType       string
	AudioURL            string
	PreviewThumbnailURL string
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
	ID          string     `json:"id"`
	Login       *string    `json:"login"`
	DisplayName *string    `json:"displayName"`
	Description *string    `json:"description"`
	CreatedAt   *time.Time `json:"createdAt"`
	Videos      *ApiVideos `json:"videos"`
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
