package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	UrlBase     string  `json:"urlBase"`
	OutputPath  string  `json:"outputPath"`
	PackageType string  `json:"packageType"`
	UserList    []*User `json:"userList"`
	NamingStyle string  `json:"namingStyle"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

var ConfigInstance *Config = Load()

func Load() *Config {
	content, err := os.ReadFile("config.json")
	if err != nil {
		return &Config{
			UrlBase:     "mangacopy.com",
			OutputPath:  "./",
			PackageType: "cbz",
			NamingStyle: "title",
			UserList:    []*User{},
		}
	}

	var config *Config = &Config{}
	err = json.Unmarshal(content, config)
	if err != nil {
		fmt.Println("Error deserializing config:", err)
		return &Config{
			UrlBase:     "mangacopy.com",
			OutputPath:  "./",
			PackageType: "cbz",
			UserList:    []*User{},
			NamingStyle: "title",
		}
	}

	return config
}

func (c *Config) Save() {
	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("Error serializing config:", err)
		return
	}

	err = os.WriteFile("config.json", content, 0644)
	if err != nil {
		fmt.Println("Error writing config file:", err)
	}
}

// SaveConfig  .
func (c *Config) SaveConfig(config *Config) {
	config.UserList = ConfigInstance.UserList
	*ConfigInstance = *config
	ConfigInstance.Save()
}

func (c *Config) GetConfig() *Config {
	return ConfigInstance
}
