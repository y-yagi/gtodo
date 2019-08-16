package gtodo

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/y-yagi/goext/osext"
)

type Cache struct {
	path string
}

func NewCache(path string) *Cache {
	cache := &Cache{path: path}
	return cache
}

func (c *Cache) Read(key string) ([]byte, error) {
	file := filepath.Join(c.path, key)
	if !osext.IsExist(file) {
		return nil, nil
	}
	return ioutil.ReadFile(file)
}

func (c *Cache) Write(key string, value []byte) error {
	file := filepath.Join(c.path, key)
	return ioutil.WriteFile(file, value, 0644)
}

func (c *Cache) Delete(key string) error {
	file := filepath.Join(c.path, key)
	if !osext.IsExist(file) {
		return nil
	}

	return os.Remove(file)
}
