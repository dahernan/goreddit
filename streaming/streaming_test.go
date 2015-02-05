package streaming

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/dahernan/goreddit/api"
)

func TestBasicStreamming(t *testing.T) {

	consumerKey := os.Getenv("CONSUMER_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	reddit := api.NewReddit(&http.Client{}, "go reddit test", consumerKey, secretKey)

	v := url.Values{}
	v.Set("limit", "20")
	stream := NewRedditStream(reddit.Hot, "trailers", v)

	items := stream.Stream()

	for it := range items {
		fmt.Println("NEW:::: ", it)
	}

}
