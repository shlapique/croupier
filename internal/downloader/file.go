package downloader

type File interface {
	GetID() string
	GetName() string
	GetPath() string
	GetHref() string
	GetMD5() *string
}
