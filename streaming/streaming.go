package streaming

import (
	"net/url"

	"github.com/dahernan/goreddit/api"
)

type RedditStreamer interface {
	Stream(redditFunc RedditListingFunc) chan api.Item
}

func (r *RedditStreamer) Stream(redditFunc RedditListingFunc) chan api.Item {

	stream := make(chan api.Item)

	v := url.Values{}
	v.Set("limit", "20")

	go func() {
		items, err := redditFunc("trailers", v)
		for item := range items {
			stream <- item
		}
	}()

	return stream

}
