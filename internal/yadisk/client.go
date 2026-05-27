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
)

const baseURL = "https://cloud-api.yandex.net/v1/disk/resources"

type Client struct {
	token   string
	http    *http.Client
	baseURL *url.URL
}

func New(token string) *Client {
	pURL, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	return &Client{
		token:   token,
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: pURL,
	}
}

// <REQ>
// path: path to resource relative to / (root) of a disk
// <RESP>
// items: list of resources for this 'path'
func (c *Client) GetMeta(ctx context.Context, path string) (*Resource, error) {
	u := *c.baseURL
	// u.Path += "/resources"

	fmt.Printf("Full path: %s\n", u.Path)

	q := u.Query()
	q.Set("path", path)
	u.RawQuery = q.Encode()

	fmt.Printf("Full qeuery: %s\n", u.RawQuery)

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
