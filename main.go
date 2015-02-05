package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dahernan/goreddit/api"
	"github.com/unrolled/render"
)

var (
	client      *http.Client
	r           *render.Render
	consumerKey string
	secretKey   string
)

func init() {
	client = NewHttpClientWithTimeout(2 * time.Second)
	r = render.New(render.Options{
		Directory:  "templates",                // Specify what path to load the templates from.
		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
	})

}

func Index(w http.ResponseWriter, req *http.Request) {
	reddit := api.NewReddit(client, "go reddit test", consumerKey, secretKey)

	v := url.Values{}
	v.Set("limit", "20")
	items, err := reddit.Hot("trailers", v)
	if err != nil {
		log.Fatalln(err)
	}

	r.HTML(w, http.StatusOK, "index", items)

}

func main() {

	consumerKey = os.Getenv("CONSUMER_KEY")
	secretKey = os.Getenv("SECRET_KEY")

	mux := http.NewServeMux()

	mux.HandleFunc("/index.html", Index)

	//mux.Handle("/", http.FileServer(http.Dir("public")))

	http.ListenAndServe(":3000", mux)

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
