package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("/Users/hugozhu/Projects/hugozhu/weibo2twitter/config.json")
	if config == nil {
		t.Fatalf("Config is nil")
	}
	config.Save()
}
