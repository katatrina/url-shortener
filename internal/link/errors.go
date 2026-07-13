package link

import "errors"

var (
	ErrSlugExists        = errors.New("slug already exists")
	ErrLinkNotFound      = errors.New("link not found")
	ErrLinkQuotaExceeded = errors.New("link quota exceeded")
)
