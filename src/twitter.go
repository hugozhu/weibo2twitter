package main

import (
	"config"
	"fmt"
	"github.com/hugozhu/goweibo"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"
)

func SubStringByByte(str string, len2 int) string {
	if len(str) <= len2 {
		return str
	}
	n := 0
	for c := range str {
		if c > len2 {
			return str[:n]
		}
		n = c
	}
	return str
}

func SubStringByChar(str string, len int) string {
	c, n := 0, 0
	for c = range str {
		n++
		if n > len {
			return str[0:c]
		}
	}
	return str
}

type Post struct {
	Id   int64
	Text string
}

func Timeline(screen_name string, since_id int64) []Post {
	var posts = []Post{}
	for _, p := range sina.TimeLine(0, screen_name, since_id, 20) {
		id := p.Id
		text := p.Text
		link := ""
		if p.Original_Pic != "" {
			link = " ✈ " + p.Original_Pic
		}
		if p.Retweeted_Status != nil {
			if p.Retweeted_Status.User != nil {
				re_user := p.Retweeted_Status.User
				text = text + " //RT @" + re_user.Name + ": " + p.Retweeted_Status.Text
			}

			if p.Retweeted_Status.Original_Pic != "" {
				link = " ✈ " + p.Retweeted_Status.Original_Pic
			}
		}
		len1 := utf8.RuneCount([]byte(text))
		len2 := utf8.RuneCount([]byte(link))
		if len1+len2 > 140 {
			link2 := fmt.Sprintf("http://weibo.com/%d/%s", p.User.Id, sina.QueryMid(id, 1))
			text = SubStringByChar(text, 140-38-len2) + link + " " + link2
		} else {
			text = text + link
		}
		posts = append(posts, Post{id, text})
	}
	return posts
}

func sync(name string, user *config.User) {
	if user.Enabled {
		weibo_account := user.GetAccount("tsina")
		twitter_account := user.GetAccount("twitter")
		posts := Timeline(weibo_account.Name, user.Last_weibo_id)

		oauth_user := oauth1a.NewAuthorizedConfig(twitter_account.Oauth_token_key, twitter_account.Oauth_token_secret)
		client := twittergo.NewClient(twitter_config, oauth_user)
		for i := len(posts) - 1; i >= 0; i-- {
			post := posts[i]
			if post.Id > user.Last_weibo_id {
				user.Last_weibo_id = post.Id
				tweet, err := Tweet(client, post.Text)
				log.Println(weibo_account.Name, post.Text, tweet)
				if err != nil {
					log.Println("[error]", tweet, err)
				}
			}
		}
	}
}

var complete_chan chan string

var sina = &weibo.Sina{
	AccessToken: weibo.ReadToken("token"),
}

var twitter_config = &oauth1a.ClientConfig{
	ConsumerKey:    config.Twitter_ConsumerKey,
	ConsumerSecret: config.Twitter_ConsumerSecret,
}

func init() {
	var debug = false
	weibo.SetDebugEnabled(&debug)
}

func Tweet(client *twittergo.Client, post string) (*twittergo.Tweet, error) {
	data := url.Values{}
	data.Set("status", post)
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", "/1.1/statuses/update.json", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.SendRequest(req)
	if err != nil {
		return nil, err
	}
	tweet := &twittergo.Tweet{}
	err = resp.Parse(tweet)
	if err != nil {
		return nil, err
	}
	return tweet, nil
}

func main() {
	conf := config.NewConfig(os.Getenv("PWD") + "/config.json")
	defer func() {
		conf.Save()
	}()

	n := len(conf.Users())
	complete_chan := make(chan string, n)
	for name, user := range conf.Users() {
		go func(n string, u *config.User) {
			sync(n, u)
			complete_chan <- n + " is done"
		}(name, user)
	}
	log.Println("wait for complete")
	for i := 0; i < n; i++ {
		log.Print(<-complete_chan)
	}
	log.Println("done")
}
