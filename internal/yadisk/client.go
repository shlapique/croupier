package yadisk

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	// "os"
	"context"
	// "os/signal"
	"time"
	"strconv"
)

const baseURL = "https://cloud-api.yandex.net/v1/disk/resources"

type Client struct {
	token   string
	http    *http.Client
	baseURL *url.URL
}

type Config struct {
	Token string
	Timeout time.Duration
}

func New(config Config) *Client {
	pURL, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	return &Client{
		token:   config.Token,
		http:    &http.Client{Timeout: config.Timeout * time.Second},
		baseURL: pURL,
	}
}

// <REQ>
// path: path to resource relative to / (root) of a disk
// <RESP>
// items: list of resources for this 'path'
func (c *Client) GetMeta(ctx context.Context, path string, limit int, offset int) (*Resource, error) {
	u := *c.baseURL

	fmt.Printf("Full path: %s\n", u.Path)

	q := u.Query()
	q.Set("path", path)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	fmt.Printf("Full encoded queuery: %s\n", u.RawQuery)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+c.token)

	resp, err := c.http.Do(req)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

	var resource Resource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, err
	}

	return &resource, nil
}

// map: Resource array -> Names array
func MapNames[T any](items *[]T, getName func(T) string) []string {
	names := make([]string, 0, len(*items))
	for _, item := range *items {
		names = append(names, getName(item))
	}
	return names
}
