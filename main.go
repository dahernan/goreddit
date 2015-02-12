package main

import (
	"encoding/json"
	"fmt"
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
)

func init() {
	client = NewHttpClientWithTimeout(2 * time.Second)
	r = render.New(render.Options{
		Directory:  "templates",                // Specify what path to load the templates from.
		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
	})

	consumerKey = os.Getenv("CONSUMER_KEY")
	secretKey = os.Getenv("SECRET_KEY")

	reddit := api.NewReddit(client, "go reddit test", consumerKey, secretKey)
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

func Event(w http.ResponseWriter, req *http.Request) {
	r.HTML(w, http.StatusOK, "event", nil)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/index.html", Index)
	mux.HandleFunc("/test", Event)

	broker := sse.NewBroker()
	mux.Handle("/events", broker)

	//mux.Handle("/", http.FileServer(http.Dir("public")))

	go StreamingBroker(broker)

	http.ListenAndServe(":3000", mux)

}

func Stream(reddit api.RedditListing) streaming.RedditStreamer {
	v := url.Values{}
	v.Set("limit", "40")
	stream := streaming.NewRedditStream(reddit.New, "videos", v)

	return stream

}

func StreamingBroker(broker *sse.Broker) {

	items := stream.Stream()
	for it := range items {

		data, err := ItemToJson(it)
		if err != nil {
			log.Println("Error marshal item:", it, err)
		}

		broker.Send(data)
		fmt.Println("New Item: ", data)
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
