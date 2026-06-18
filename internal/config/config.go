package config

import (
	"bytes"
	"io/ioutil"
	"fmt"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

var defaultConfig = `
---
client:
  path: "disk:/kindle/"
  page_size: 15
  timeout: 15
preloader:
  timeout: 15
  window_size: 5
  window_lag: 2
  workers_num: 2
downloader:
  path: "./"
  max_concurrent_files: 50
  workers_num: 2
server:
  port: "1234"
`

type Client struct {
	Path     string `yaml:"path"`
	PageSize int    `yaml:"page_size"`
	Timeout  int    `yaml:"timeout"`
}

type Preloader struct {
	Timeout    int `yaml:"timeout"`
	WindowSize int `yaml:"window_size"`
	WindowLag  int `yaml:"window_lag"`
	WorkersNum int `yaml:"workers_num"`
}

type Downloader struct {
	Path               string `yaml:"path"`
	MaxConcurrentFiles int    `yaml:"max_concurrent_files"`
	WorkersNum         int    `yaml:"workers_num"`
}

type Server struct {
	Port string `yaml:"port"`
}

type App struct {
	Client     Client
	Preloader  Preloader
	Downloader Downloader
	Server     Server
}

func Load(path string) (*App, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Unable to load config file [%s]: %s\n", path, err)
		fmt.Printf("Loading default config\n")
		return loadDefault()
	}

	config := new(App)

	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}
	return Mrg(config)
}

// here we merge not full (optional) user config from config.yml
// with defaultConfig
func Mrg(config *App) (*App, error) {
	defaultApp, err := loadDefault()
	if err != nil {
		fmt.Printf("Unable to load DEFAULT config: %s\n", err)
		return nil, err
	}
	mergo.Merge(&config, defaultApp)
	return config, nil
}

func loadDefault() (*App, error) {
	c := defaultConfig
	config := new(App)
	dec := yaml.NewDecoder(bytes.NewReader([]byte(c)))
	dec.KnownFields(true)
	err := dec.Decode(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
