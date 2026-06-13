package yadisk

import (
	"time"
)

type ResourceType string

const (
	ResourceTypeFile ResourceType = "file"
	ResourceTypeDir  ResourceType = "dir"
)

type Resource struct {
	Name     string       `json:"name"`
	Path     string       `json:"path"`
	Created  time.Time    `json:"created"`
	Modified time.Time    `json:"modified"`
	Type     ResourceType `json:"type"`

	PublicKey *string `json:"public_key,omitempty"`
	PublicURL *string `json:"public_url,omitempty"`

	CustomProperties map[string]string `json:"custom_properties,omitempty"`

	Embedded *ResourceList `json:"_embedded,omitempty"`

	MD5      *string `json:"md5,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`
	Size     *int64  `json:"size,omitempty"`
	Preview  *string `json:"preview,omitempty"`

	OriginPath *string `json:"origin_path,omitempty"`
}

type ResourceList struct {
	Items     []Resource `json:"items"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	Total     int        `json:"total"`
	Path      string     `json:"path"`
	Sort      *string    `json:"sort,omitempty"`
	PublicKey *string    `json:"public_key,omitempty"`
}

type GetLinkResponse struct {
	Href      string
	Method    string
	templated bool
}

type Page struct {
	Files []File
}

// Resource subset
type File struct {
    Id   string  `json:"id"`
    Name string  `json:"name"`
    Path string  `json:"path"`
    MD5  *string `json:"md5"`
    Href string  `json:"href"`
}

func (f *File) GetID() string {
	return f.Id
}

func (f *File) GetName() string {
	return f.Name
}

func (f *File) GetPath() string {
	return f.Path
}

func (f *File) GetHref() string {
	return f.Href
}

func (f *File) GetMD5() *string {
	return f.MD5
}
