package config

import (
	"bytes"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

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
		return nil, err
	}

	config := new(App)

	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadDefault(c string) (*App, error) {
	config := new(App)
	dec := yaml.NewDecoder(bytes.NewReader([]byte(c)))
	dec.KnownFields(true)
	err := dec.Decode(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
