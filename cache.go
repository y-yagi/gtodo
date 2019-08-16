package gtodo

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/y-yagi/goext/osext"
)

// Cache is a type for file cache.
type Cache struct {
	path string
}

// NewCache create a new Cache.
func NewCache(path string) *Cache {
	cache := &Cache{path: path}
	return cache
}

// Read cache.
func (c *Cache) Read(key string) ([]byte, error) {
	file := filepath.Join(c.path, key)
	if !osext.IsExist(file) {
		return nil, nil
	}
	return ioutil.ReadFile(file)
}

// Write create a new cache.
func (c *Cache) Write(key string, value []byte) error {
	file := filepath.Join(c.path, key)
	return ioutil.WriteFile(file, value, 0644)
}

// Delete delete cache.
func (c *Cache) Delete(key string) error {
	file := filepath.Join(c.path, key)
	if !osext.IsExist(file) {
		return nil
	}

	return os.Remove(file)
}
