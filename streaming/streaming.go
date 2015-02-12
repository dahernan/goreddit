package streaming

import (
	"log"
	"net/url"
	"time"

	"golang.org/x/net/context"

	"github.com/dahernan/goreddit/api"
)

type RedditStreamer interface {
	Stream() chan api.Item
	Items() []api.Item
}

type RedditStream struct {
	// redit api params
	redditFunc api.RedditListingFunc
	subreddit  string
	params     url.Values

	// streaming channel and cancelation function
	stream     chan api.Item
	cancelFunc context.CancelFunc

	// cache the last request
	items []api.Item
	keys  map[string]bool

	// channels to copy the items
	itemsChan    chan struct{}
	itemsOutChan chan []api.Item
}

func NewRedditStream(redditFunc api.RedditListingFunc, subreddit string, params url.Values) *RedditStream {

	stream := make(chan api.Item, 100)

	// api params
	rs := &RedditStream{
		redditFunc: redditFunc,
		subreddit:  subreddit,
		params:     params,
	}

	// stream and cancel
	ctx, cancel := context.WithCancel(context.Background())
	rs.stream = stream
	rs.cancelFunc = cancel

	// channels to copy items
	rs.itemsChan = make(chan struct{})
	rs.itemsOutChan = make(chan []api.Item)

	// Start the event loop
	go rs.run(ctx)

	return rs
}

func (r *RedditStream) Stream() chan api.Item {
	return r.stream
}

func (r *RedditStream) Cancel() {
	r.cancelFunc()
}

func (r *RedditStream) run(ctx context.Context) {
	for {
		select {
		case <-time.After(5 * time.Second):
			r.fetchItems()

		case <-r.itemsChan:
			r.itemsOutChan <- r.doItems()

		case <-ctx.Done():
			log.Println("RedditStream: Finished run")
			return
		}
	}
}

func (r *RedditStream) fetchItems() {
	items, _ := r.redditFunc(r.subreddit, r.params)

	for _, item := range items {
		_, ok := r.keys[item.Data.Id]
		if !ok {
			r.stream <- item
		}
	}

	// rebuild the cache
	r.keys = make(map[string]bool)
	r.items = items
	for _, item := range items {
		r.keys[item.Data.Id] = true
	}
}

func (r *RedditStream) Items() []api.Item {
	r.itemsChan <- struct{}{}
	return <-r.itemsOutChan
}

func (r *RedditStream) doItems() []api.Item {
	size := len(r.items)
	items := make([]api.Item, size, size)
	if size == 0 {
		return items
	}
	copy(items, r.items)
	return items
}
