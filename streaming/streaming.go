package streaming

import (
	"fmt"
	"net/url"
	"time"

	"golang.org/x/net/context"

	"github.com/dahernan/goreddit/api"
)

type RedditStreamer interface {
	Stream() chan api.Item
}

type RedditStream struct {
	redditFunc api.RedditListingFunc
	subreddit  string
	params     url.Values
	stream     chan api.Item
	cancelFunc context.CancelFunc
}

func NewRedditStream(redditFunc api.RedditListingFunc, subreddit string, params url.Values) *RedditStream {

	stream := make(chan api.Item, 100)

	rs := &RedditStream{
		redditFunc: redditFunc,
		subreddit:  subreddit,
		params:     params,
		stream:     stream,
	}
	ctx, cancel := context.WithCancel(context.Background())
	rs.cancelFunc = cancel
	go rs.run(ctx)

	return rs
}

func (r *RedditStream) Stream() chan api.Item {
	return r.stream
}

func (r *RedditStream) run(ctx context.Context) {
	for {
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("Fetching items ...")
			r.fetchItems()

		case <-ctx.Done():
			fmt.Println("Done Run!!")
			return
		}
	}
}

func (r *RedditStream) fetchItems() {
	items, _ := r.redditFunc(r.subreddit, r.params)

	fmt.Println("Sending items!!", len(items))

	for i, item := range items {
		fmt.Println("Sending item-", i)
		r.stream <- item
	}

}
