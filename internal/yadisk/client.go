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

type ResourceType string

const (
	ResourceTypeFile ResourceType = "file"
	ResourceTypeDir  ResourceType = "dir"
)

type Resource struct {
	// Common fields for both files and folders
	Name            string            `json:"name"`
	Path            string            `json:"path"`
	Created         time.Time         `json:"created"`
	Modified        time.Time         `json:"modified"`
	Type            ResourceType      `json:"type"`
	
	// Public resource fields (only present if the resource is published)
	PublicKey       *string           `json:"public_key,omitempty"`
	PublicURL       *string           `json:"public_url,omitempty"`

	// Custom metadata (only if set via meta-add request)
	CustomProperties map[string]string `json:"custom_properties,omitempty"` // User-defined attributes

	// For folders only (present in the "_embedded" field)
	Embedded        *ResourceList     `json:"_embedded,omitempty"`

	// For files only
	MD5             *string           `json:"md5,omitempty"`
	MimeType        *string           `json:"mime_type,omitempty"`
	Size            *int64            `json:"size,omitempty"`
	Preview         *string           `json:"preview,omitempty"`

	// Fields for resources in Trash
	OriginPath      *string           `json:"origin_path,omitempty"` // Path before being moved to Trash
}

type ResourceList struct {
	Items       []Resource `json:"items"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	Total       int        `json:"total"`
	Path        string     `json:"path"`
	Sort        *string    `json:"sort,omitempty"`
	PublicKey   *string    `json:"public_key,omitempty"`
}


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
