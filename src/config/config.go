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
	Enabled  bool
	Accounts map[string]Account
}

type Config struct {
	file     string
	users    map[string]User
	last_ids map[string]int64
}

func (this Config) GetUser(name string) User {
	return this.users[name]
}

func (this Config) SaveLastIds() {
	f, err := os.OpenFile(this.file+".id", os.O_RDWR|os.O_CREATE, 0775)
	if err != nil {
		log.Fatalf("Failed to open %s %v\n", this.file, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	err = enc.Encode(&this.last_ids)
	if err != nil {
		log.Fatalf("Failed to save %s %v\n", this.file, err)
	}
}

func (this User) GetAccount(name string) Account {
	return this.Accounts[name]
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
