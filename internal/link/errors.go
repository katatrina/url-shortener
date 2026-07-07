package link

import "errors"

var (
	ErrSlugExists   = errors.New("slug already exists")
	ErrLinkNotFound = errors.New("link not found")
)
