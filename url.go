package main

import (
	"fmt"
	"net/url"
)

// PublicURL builds the canonical public URL for an object in a public-read bucket.
// key is the raw object key (e.g. "20260619_153022_photo.jpg"); it gets URL-escaped
// to handle spaces, Chinese characters, parentheses, etc.
func PublicURL(bucket, region, key string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", bucket, region, url.PathEscape(key))
}
