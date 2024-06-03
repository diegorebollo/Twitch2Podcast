package rss

import (
	"time"

	"github.com/eduncan911/podcast"
)

func Generator(channelId *int) string {

	p := podcast.New("dddd", "cc", "aaaaa", &time.Time{}, &time.Time{})

	rss := string(p.Bytes())

	return rss
}
