package main

import (
	"config"
	"fmt"
	"github.com/bsdf/twitter"
	"github.com/hugozhu/goweibo"
	"log"
	"os"
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
		Timeline(weibo_account.Name, user.Last_weibo_id)
		posts := Timeline(weibo_account.Name, user.Last_weibo_id)
		t := twitter.Twitter{
			ConsumerKey:      config.Twitter_ConsumerKey,
			ConsumerSecret:   config.Twitter_ConsumerSecret,
			OAuthToken:       twitter_account.Oauth_token_key,
			OAuthTokenSecret: twitter_account.Oauth_token_secret,
		}
		for i := len(posts) - 1; i >= 0; i-- {
			post := posts[i]
			if post.Id > user.Last_weibo_id {
				user.Last_weibo_id = post.Id
				tweet, err := t.Tweet(post.Text)
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

func init() {

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
