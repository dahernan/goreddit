package streaming

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/dahernan/goreddit/api"
)

// no very meanfull test, just experimenting
func TestBasicStreamming(t *testing.T) {

	consumerKey := os.Getenv("CONSUMER_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	reddit := api.NewReddit(&http.Client{}, "go reddit test", consumerKey, secretKey)

	v := url.Values{}
	v.Set("limit", "5")
	stream := NewRedditStream(reddit.Hot, "trailers", v)

	items := stream.Stream()

	i := 0
	for it := range items {
		i++
		fmt.Println("New Item: ", i, it)
	}

}
