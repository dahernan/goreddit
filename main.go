package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/unrolled/render"

	"github.com/dahernan/goreddit/api"
	"github.com/dahernan/goreddit/sse"
	"github.com/dahernan/goreddit/streaming"
)

var (
	client *http.Client
	r      *render.Render

	stream      streaming.RedditStreamer
	consumerKey string
	secretKey   string
	port        string
	agent       string
)

func init() {
	client = NewHttpClientWithTimeout(2 * time.Second)
	r = render.New(render.Options{
		Directory:  "templates",                // Specify what path to load the templates from.
		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
	})

	consumerKey = os.Getenv("CONSUMER_KEY")
	secretKey = os.Getenv("SECRET_KEY")
	port = os.Getenv("PORT")
	agent = os.Getenv("AGENT")

	reddit := api.NewReddit(client, agent, consumerKey, secretKey)
	stream = Stream(reddit)

}

func Index(w http.ResponseWriter, req *http.Request) {

	items := stream.Items()

	first := items[0]
	items = items[1:5]

	data := map[string]interface{}{
		"First": first,
		"Items": items,
	}

	r.HTML(w, http.StatusOK, "index", data)

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", Index)

	broker := sse.NewBroker()
	mux.Handle("/events", broker)

	go StreamingBroker(broker)
	http.ListenAndServe(":"+port, mux)

}

func Stream(reddit api.RedditListing) streaming.RedditStreamer {
	v := url.Values{}
	v.Set("limit", "40")
	stream := streaming.NewRedditStream(reddit.New, "videos", v)

	return stream

}

func StreamingBroker(broker *sse.Broker) {

	items := stream.Stream()

	for {
		select {
		case it := <-items:
			data, err := ItemToJson(it)
			if err != nil {
				log.Println("Error marshal item:", it, err)
			}

			broker.Send(data)
			//fmt.Println("New Item: ", data)

		// heroku proxy timesout if the connection does not have activity
		// so send something to prevent it
		case <-time.After(5 * time.Second):
			broker.Send("{}")
		}
	}

}

func ItemToJson(it api.Item) (string, error) {
	data, err := json.Marshal(it)
	return string(data), err
}

func NewHttpClientWithTimeout(timeout time.Duration) *http.Client {
	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}
	return &client
}
