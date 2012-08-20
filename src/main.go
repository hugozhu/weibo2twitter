package main

import (
	"encoding/json"
	"github.com/bsdf/twitter"
	"io/ioutil"
	"log"
	"net/http"
	// "strconv"
	"config"
	"fmt"
	"strconv"
	"unicode/utf8"
)

//https://api.weibo.com/oauth2/authorize?client_id=3558864612&redirect_uri=http://cn.myalert.info/test.php&response_type=token
//curl --data "client_id=3558864612&client_secret=b50f7e096d4048ab39f151888e628345&grant_type=authorization_code&code=ce4359430f05c73bf0c11f45ab8d0621&redirect_uri=http://cn.myalert.info/test.php" https://api.weibo.com/oauth2/access_token

//https://api.weibo.com/2/statuses/user_timeline.json?access_token=2.008TkTLDIQdqsD2b947e60590yQenc&screen_name=hugozhu

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

//3480730133033379
func Timeline(screen_name string, since_id int64) []Post {
	url := "https://api.weibo.com/2/statuses/user_timeline.json?access_token=2.008TkTLDIQdqsDbfb2b62864h6jrPC"
	url += fmt.Sprintf("&screen_name=%s&since_id=%d", screen_name, since_id)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)

	var posts = []Post{}
	if resp.StatusCode == 200 {
		var data map[string]interface{}
		json.Unmarshal(bytes, &data)

		for _, entry := range data["statuses"].([]interface{}) {
			entry := entry.(map[string]interface{})
			id, _ := strconv.ParseInt(entry["idstr"].(string), 10, 64)
			text := entry["text"].(string)
			link := ""
			if entry["original_pic"] != nil {
				link = " ✈ " + entry["original_pic"].(string)
			}

			if entry["retweeted_status"] != nil {
				retweeted := entry["retweeted_status"].(map[string]interface{})
				re_user := retweeted["user"].(map[string]interface{})
				text = text + " //RT @" + re_user["name"].(string) + ": " + retweeted["text"].(string)

				if retweeted["original_pic"] != nil {
					link = " ✈ " + retweeted["original_pic"].(string)
				}
			}
			len1 := utf8.RuneCount([]byte(text))
			len2 := utf8.RuneCount([]byte(link))
			if len1+len2 > 140 {
				text = SubStringByChar(text, 140-len2) + link
			} else {
				text = text + link
			}

			posts = append(posts, Post{id, text})
		}
	}
	return posts
}

func main() {
	config := config.NewConfig("/Users/hugozhu/Projects/hugozhu/weibo2twitter/config.json")

	t := twitter.Twitter{
		ConsumerKey:      "5QrdEQ39A1yZJcMAuc2mwg",
		ConsumerSecret:   "bJattIehGRzbe67ei6dgSx8KGHYuj4KbI0lqVBQMQ",
		OAuthToken:       "9729212-d1ehMo6rIHWZKC9VRb7bzgrckdOCtvNDXseExsIWyT",
		OAuthTokenSecret: "KEbc80BDoH9ovlXq6y9jSiuXjHZuBO3qa9Stl1MZCCg",
	}
	log.Println(t)
	posts := Timeline("hugozhu", 3480730133033379)
	for i := len(posts) - 1; i >= 0; i-- {
		p := posts[i]
		log.Println(p)
		// _, err := t.Tweet(p.Text)
		// if err != nil {
		// 	log.Println("[error]", err, p)
		// }
	}

}
