package twichapi

import (
	"bufio"
	"bytes"
	"drebollo/twitchtopodcast/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var clientId = "kimne78kx3ncx6brgo4mv6wki5h1ko"

func makeRequest(bodyContent string) []uint8 {
	url := "https://gql.twitch.tv/gql"
	contentType := "application/json"
	data := []byte(bodyContent)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Client-ID", clientId)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	return body
}

func ChannelData(channel string) models.ApiData {

	bodyContent := fmt.Sprintf(`{"query": "{ user(login: \"%s\") { id login displayName description createdAt profileImageURL(width: 300) }}"}`, channel)
	resp := makeRequest(bodyContent)

	var userResp models.ApiResponse

	err := json.Unmarshal(resp, &userResp)

	if err != nil {
		fmt.Println(err)
	}

	return userResp.Data

}

func GetAllVodsFromChannel(channelId int) []models.ApiEdges {

	bodyContent := fmt.Sprintf(`{"query": "{ user(id: \"%s\") { id videos(first: 100, type: ARCHIVE) { edges { node { id title description language createdAt lengthSeconds broadcastType previewThumbnailURL(height: 720, width: 1280) } } } }}"}`, fmt.Sprint(channelId))
	resp := makeRequest(bodyContent)

	var vodsResp models.ApiResponse
	err := json.Unmarshal(resp, &vodsResp)

	if err != nil {
		fmt.Println(err)
	}
	return vodsResp.Data.User.Videos.Edges
}

func GetAudioUrl(vodId int) *string {
	playlist := string(getPlayList(vodId))

	if !strings.Contains(playlist, "#EXTM3U") {
		log.Printf("Vod:'%d' Private M3U List", vodId)
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(playlist))
	var url string
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "/audio_only/") {
			url = scanner.Text()
		}
	}
	return &url
}

func getPlayList(vodId int) []uint8 {
	accessToken := getAccessToken(vodId)

	if accessToken == nil {
		log.Printf("Vod:'%d' NOT exist", vodId)
		return nil
	}

	url := fmt.Sprintf(`https://usher.ttvnw.net/vod/%d.m3u8?client_id=%s&token=%s&sig=%s&allow_source=true&allow_audio_only=true`, vodId, clientId, accessToken.Value, accessToken.Signature)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
	}

	return body

}

func getAccessToken(vodId int) *models.ApiVideoPlaybackAccessToken {
	bodyContent := fmt.Sprintf(`{"query": "{ videoPlaybackAccessToken(id: \"%d\", params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: \"site\"}) { signature value }}"}`, vodId)
	resp := makeRequest(bodyContent)

	var accessTokenResp models.ApiResponse
	err := json.Unmarshal(resp, &accessTokenResp)

	if err != nil {
		fmt.Println(err)
	}
	return accessTokenResp.Data.VideoPlaybackAccessToken
}
