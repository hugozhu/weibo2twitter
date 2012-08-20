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
	Accounts      map[string]Account //never change
	Last_weibo_id int64
}

type Config struct {
	file  string
	users map[string]*User //may change during runtime
}

func (this Config) Users() map[string]*User {
	return this.users
}

func (this Config) Save() {
	f, err := os.OpenFile(this.file, os.O_RDWR, 0755)
	defer f.Close()

	if err != nil {
		log.Fatalf("Failed to open %s %v\n", this.file, err)
	}

	data, _ := json.MarshalIndent(&this.users, "", "    ")
	f.WriteString(string(data))
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
