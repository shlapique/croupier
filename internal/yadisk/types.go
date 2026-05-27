package yadisk

import "time"

type ResourceType string

const (
	ResourceTypeFile ResourceType = "file"
	ResourceTypeDir  ResourceType = "dir"
)

type Resource struct {
	Name            string            `json:"name"`
	Path            string            `json:"path"`
	Created         time.Time         `json:"created"`
	Modified        time.Time         `json:"modified"`
	Type            ResourceType      `json:"type"`
	
	PublicKey       *string           `json:"public_key,omitempty"`
	PublicURL       *string           `json:"public_url,omitempty"`

	CustomProperties map[string]string `json:"custom_properties,omitempty"`

	Embedded        *ResourceList     `json:"_embedded,omitempty"`

	MD5             *string           `json:"md5,omitempty"`
	MimeType        *string           `json:"mime_type,omitempty"`
	Size            *int64            `json:"size,omitempty"`
	Preview         *string           `json:"preview,omitempty"`

	OriginPath      *string           `json:"origin_path,omitempty"`
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
