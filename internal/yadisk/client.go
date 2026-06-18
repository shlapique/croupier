package yadisk

import (
	"context"
	// "encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	// "os"
	// "os/signal"
	// "crypto/md5"
	// "io"
	"log/slog"
	"strconv"
	"time"
)

const baseURL = "https://cloud-api.yandex.net/v1/disk/resources"

type Client struct {
	l *slog.Logger

	token   string
	http    *http.Client
	baseURL *url.URL
}

type Config struct {
	L       *slog.Logger
	Token   string
	Timeout time.Duration
}

func New(config Config) *Client {
	l := config.L
	if l == nil {
		l = slog.New(slog.Default().Handler())
	}
	l = l.With("pkg", "client")

	pURL, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	return &Client{
		l:       l,
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

	c.l.Debug("Full path", "path", u.Path)

	q := u.Query()
	q.Set("path", path)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	c.l.Debug("Full encoded queuery", "q", u.RawQuery)

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

// map: Resource array -> Resource Subset Struct
func MapSubset[T any, U any](items *[]T, f func(T) U) []U {
	out := make([]U, 0, len(*items))
	for _, item := range *items {
		out = append(out, f(item))
	}
	return out
}

// returns href to direct download with smth like curl
func (c *Client) GetDownloadLink(ctx context.Context, file File) (*GetLinkResponse, error) {
	c.l.Debug("Trying to get download link to a file", "path", file.Path)

	u := *c.baseURL

	u.Path = u.Path + "/download"
	q := u.Query()
	q.Set("path", file.Path)
	u.RawQuery = q.Encode()

	c.l.Debug("Full encoded queuery", "q", u.RawQuery)

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

	var getLinkResponse GetLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&getLinkResponse); err != nil {
		return nil, err
	}

	return &getLinkResponse, nil
}
