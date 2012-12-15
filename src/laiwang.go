package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sqlite"
	"strconv"
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

func GetStringLength(str string) int {
	return utf8.RuneCount([]byte(str))
}

type Post struct {
	Id     int64
	Text   string
	PicUrl string
}

func sync_weibo_2_laiwang(user map[string]interface{}) {
	var last_weibo_id int64
	if user["pos"] != nil {
		last_weibo_id = user["pos"].(int64)
	}

	posts := Timeline(ACCESS_TOKEN, user["weibo_uid"].(int64), last_weibo_id)

	for i := len(posts) - 1; i >= 0; i-- {
		post := posts[i]
		if post.Id > last_weibo_id {
			last_weibo_id = post.Id
			post_laiwang(user, post.Text, post.PicUrl)
			log.Println(user["weibo_name"], post.Id, post.Text, post.PicUrl)
			db_execute(user["id"].(int64), last_weibo_id)
		}
	}
}

func db_execute(id int64, pos int64) {
	sqlite.Run(os.Getenv("PWD")+"/db.sqlite3", func(db *sqlite.DB) {
		sql := fmt.Sprintf("insert or replace into weibo_seq (id, pos) values (%d, %d)", id, pos)
		db.Execute(sql)
	})
}

func Timeline(access_token string, uid int64, since_id int64) []Post {
	url := "https://api.weibo.com/2/statuses/user_timeline.json?access_token=" + access_token
	url += fmt.Sprintf("&uid=%d&since_id=%d&count=20", uid, since_id)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(url, err)
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)

	log.Println(url)

	var posts = []Post{}
	if resp.StatusCode == 200 {
		var data map[string]interface{}
		json.Unmarshal(bytes, &data)

		// log.Println(string(bytes))

		if data["statuses"] != nil {
			for _, entry := range data["statuses"].([]interface{}) {
				entry := entry.(map[string]interface{})
				id, _ := strconv.ParseInt(entry["idstr"].(string), 10, 64)
				text := entry["text"].(string)
				link := ""
				if entry["original_pic"] != nil {
					link = entry["original_pic"].(string)
				}

				if entry["retweeted_status"] != nil {
					retweeted := entry["retweeted_status"].(map[string]interface{})
					if retweeted["user"] != nil {
						re_user := retweeted["user"].(map[string]interface{})
						text = text + " //RT @" + re_user["name"].(string) + ": " + retweeted["text"].(string)
					}

					if retweeted["original_pic"] != nil {
						link = retweeted["original_pic"].(string)
					}
				}
				posts = append(posts, Post{id, text, link})
			}
		}
	} else {
		log.Fatal(string(bytes))
	}
	return posts
}

var ACCESS_TOKEN string

//curl -d ""  "https://open.laiwang.com/v1/post/add?access_token=f4f55c77856768d983e1671bbcd195&content=Hello"
//curl -H "Expect:"  --form access_token=f4f55c77856768d983e1671bbcd195 --form content=hello  --form pic=@test.jpg  "https://open.laiwang.com/v1/post/addwithpic"

func post_laiwang(user map[string]interface{}, content string, pic_url string) {
	var resp *http.Response
	var err error

	token := user["laiwang_access_token"].(string)

	if pic_url != "" {
		buf := new(bytes.Buffer)
		w := multipart.NewWriter(buf)
		w.WriteField("access_token", token)
		w.WriteField("content", content)
		wr, _ := w.CreateFormFile("pic", "weibo.jpg")
		resp, err = http.Get(pic_url)
		if err == nil {
			io.Copy(wr, resp.Body)
		}
		w.Close()
		resp, err = http.Post("https://open.laiwang.com/v1/post/addwithpic", w.FormDataContentType(), buf)
	} else {
		resp, err = http.PostForm("https://open.laiwang.com/v1/post/add", url.Values{
			"access_token": {token},
			"content":      {content},
		})
	}
	bytes, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(bytes), err)
}

func fetch_img(url string) []byte {
	resp, err := http.Get(url)
	if err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		return bytes
	}
	return nil
}

var complete_chan chan string
var users map[int64]map[string]interface{}

func init() {
	users = make(map[int64]map[string]interface{})
	sqlite.Run(os.Getenv("PWD")+"/db.sqlite3", func(db *sqlite.DB) {
		results := db.Query("select b.*, a.pos from users b left join weibo_seq a on a.id=b.id")
		for _, row := range results {
			users[row["id"].(int64)] = row
		}
	})

	data, err := ioutil.ReadFile(os.Getenv("PWD") + "/token")
	if err != nil {
		log.Fatal(err)
	}
	ACCESS_TOKEN = strings.TrimSpace(string(data))
}

func main() {
	n := len(users)
	complete_chan := make(chan string, n)
	for id, user := range users {
		go func(id int64, u map[string]interface{}) {
			sync_weibo_2_laiwang(u)
			complete_chan <- fmt.Sprintf("%v is done", id)
		}(id, user)
	}
	log.Println("wait for completion")
	for i := 0; i < n; i++ {
		log.Print(<-complete_chan)
	}
	log.Println("done")
}
