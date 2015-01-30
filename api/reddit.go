package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dahernan/goreddit/oauth"
)

const (
	oauthUrl = "https://www.reddit.com/api/v1/access_token"
	baseUrl  = "https://oauth.reddit.com"
)

var (
	DefaultRedditListing RedditListing
)

type Response struct {
	Kind string
	Data struct {
		Children []Item
	}
}

type Item struct {
	Kind string
	Data ItemData
}

type ItemData struct {
	Id        string
	Title     string
	URL       string
	Domain    string
	Thumbnail string
}

type RedditListing interface {
	New(subreddit string, params url.Values) ([]Item, error)
	Hot(subreddit string, params url.Values) ([]Item, error)
	Top(subreddit string, params url.Values) ([]Item, error)
}

type RedditListingFunc func(subreddit string, params url.Values) ([]Item, error)

type Reddit struct {
	client      *http.Client
	userAgent   string
	consumerKey string
	secretKey   string
}

func NewReddit(client *http.Client, userAgent, consumerKey, secretKey string) RedditListing {
	return &Reddit{
		client:      client,
		userAgent:   userAgent,
		consumerKey: consumerKey,
		secretKey:   secretKey,
	}
}

func (r *Reddit) New(subreddit string, params url.Values) ([]Item, error) {
	return r.Listing("new", subreddit, params)
}

func (r *Reddit) Hot(subreddit string, params url.Values) ([]Item, error) {
	return r.Listing("hot", subreddit, params)
}

func (r *Reddit) Top(subreddit string, params url.Values) ([]Item, error) {
	return r.Listing("top", subreddit, params)
}

func (r *Reddit) Listing(page string, subreddit string, params url.Values) ([]Item, error) {
	redditAuth := oauth.NewAppOnlyOAuthToken(r.client, r.userAgent)

	t, err := redditAuth.Token(oauthUrl, r.consumerKey, r.secretKey)
	if err != nil {
		return nil, err
	}

	redditUrl := fmt.Sprintf("%s/%s.json", baseUrl, page)
	if subreddit != "" {
		redditUrl = fmt.Sprintf("%s/r/%s/%s.json", baseUrl, subreddit, page)
	}

	u, err := url.Parse(redditUrl)
	if err != nil {
		return nil, err
	}

	u.RawQuery = params.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", r.userAgent)
	req.Header.Add("Authorization", strings.Join([]string{"bearer", t.AccessToken}, " "))

	resp, err := r.client.Do(req)

	log.Println(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bb, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bb))
	}

	dec := json.NewDecoder(resp.Body)

	redditResponse := Response{}
	dec.Decode(&redditResponse)

	return redditResponse.Data.Children, nil

}
