package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

type OAuthToken interface {
	Token(urlOAuth string, consumerKey string, secretKey string) (*Token, error)
}

type AppOnlyOAuthToken struct {
	client    *http.Client
	userAgent string
}

func NewAppOnlyOAuthToken(client *http.Client, userAgent string) OAuthToken {
	return &AppOnlyOAuthToken{
		client:    client,
		userAgent: userAgent,
	}
}

// Implements Application-only OAuth https://github.com/reddit/reddit/wiki/OAuth2
// It does not need for a user context and always has to be over https
func (app *AppOnlyOAuthToken) Token(urlOAuth string, consumerKey string, secretKey string) (*Token, error) {

	encodedKeys := encodeKeys(consumerKey, secretKey)
	body := []byte("grant_type=client_credentials")

	req, err := http.NewRequest("POST", urlOAuth, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", app.userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	req.Header.Add("Authorization", strings.Join([]string{"Basic", encodedKeys}, " "))

	resp, err := app.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		bb, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bb))
	}

	dec := json.NewDecoder(resp.Body)
	if err != nil {
		return nil, err
	}

	token := Token{}
	err = dec.Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func encodeKeys(consumerKey string, secretKey string) string {
	key := url.QueryEscape(consumerKey)
	secret := url.QueryEscape(secretKey)
	credentials := []string{key, secret}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(credentials, ":")))
}
