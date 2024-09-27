package image

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"gioui.org/app"
)

// CacheCapacity set how many network images can be cached.
// Set
var CacheCapacity int = 100

// the global image cache.
var imageCache *remoteImageCache

type remoteImageCache struct {
	fileDir string
	cache   *lruCache[string]
}

func newImageCache(capacity int) *remoteImageCache {
	fileDir, err := os.MkdirTemp("", "gioview-"+app.ID)
	if err != nil {
		panic(err)
	}
	return &remoteImageCache{
		fileDir: fileDir,
		cache: newLruCache[string](capacity, func(key, val string) {
			os.Remove(val)
		}),
	}
}

func (rc *remoteImageCache) get(url string) ([]byte, error) {
	path := rc.cache.Get(url)
	if path == "" {
		return nil, errors.New("no cached image")
	}

	return os.ReadFile(path)
}

func (rc *remoteImageCache) put(url string, filename string, img io.Reader) error {
	dest, err := os.Create(filepath.Join(rc.fileDir, filename))
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, img)
	if err != nil {
		return err
	}

	rc.cache.Put(url, dest.Name())
	return nil
}

func (rc *remoteImageCache) clear() {
	rc.cache.Clear()
}

func initImageCache() {
	// max cache capacity per application.
	imageCache = newImageCache(CacheCapacity)
}

// Clear the remote image cache if it is used.
func ClearCache() {
	if imageCache != nil {
		imageCache.clear()
	}
}
