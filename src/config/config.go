package config

import (
	"bufio"
	"encoding/json"
	// "io/ioutil"
	"log"
	"os"
)

type Account struct {
	Name               string
	Blogtype           string
	Oauth_token_key    string
	Oauth_token_secret string
}

type User struct {
	Enabled       bool
	Accounts      map[string]Account
	Last_weibo_id int64
}

type Config struct {
	file  string
	users map[string]User
}

func (this Config) Users() map[string]User {
	return this.users
}

func (this Config) GetUser(name string) User {
	return this.users[name]
}

func (this Config) Save() {
	data := json.Marshal(&this.users)
	log.Println(data)
}

func (this User) GetAccount(name string) Account {
	return this.Accounts[name]
}

func (this User) setLastId(id int64) {
	this.setLastId(id)
}

func NewConfig(file string) *Config {
	c := Config{file, nil}
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Failed to open %s %v\n", file, err)
	}
	defer f.Close()
	dec := json.NewDecoder(bufio.NewReader(f))
	err = dec.Decode(&c.users)
	if err != nil {
		log.Fatalf("Failed't parse %s %v\n", file, err)
	}
	return &c
}
