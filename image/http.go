package image

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	// Change this if you want to load larger images from the web. Default max size is 10M.
	MaxDownloadSize = 1024 * 1024 * 10

	// Timeout for http request.
	DownloadTimeout = time.Second * 10
)

var (
	// copied from https://github.com/imroc/req/blob/master/client_impersonate.go
	firefoxHeaders = map[string]string{
		"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:105.0) Gecko/20100101 Firefox/105.0",
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"accept-language":           "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"upgrade-insecure-requests": "1",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "same-origin",
		"sec-fetch-user":            "?1",
	}
)

// References: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
var allowedMIMETypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"binary/octet-stream",
}

// Because of network anti crawler policies applied by sites, we have to disguising
// out http client as browsers, see below link for details:
// https://req.cool/blog/supported-http-fingerprint-impersonation-to-bypass-anti-crawler-detection-effortlessly/
type HttpClient struct {
	client *http.Client
}

func newClient() *HttpClient {
	c := &http.Client{
		// wait for 10s to download the file
		Timeout: DownloadTimeout,
	}

	return &HttpClient{
		client: c,
	}
}

func (c *HttpClient) get(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range firefoxHeaders {
		request.Header.Set(key, value)
	}

	return c.client.Do(request)
}

func (c *HttpClient) Download(location string) (string, io.ReadCloser, error) {
	location = strings.TrimSpace(location)
	_, err := url.ParseRequestURI(location)
	if err != nil {
		return "", nil, err
	}

	resp, err := c.get(location)
	if err != nil {
		return "", nil, err
	}
	// let caller close the body
	//defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	if !slices.Contains[[]string, string](allowedMIMETypes, contentType) {
		return "", nil, fmt.Errorf("file MIME type: %s, not allowed to download", contentType)
	}

	if size > MaxDownloadSize {
		return "", nil, errors.New("file too large, will not download")
	}

	filename := filepath.Base(location)
	return filename, resp.Body, nil
}
