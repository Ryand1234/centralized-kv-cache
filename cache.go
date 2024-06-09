package main;

import (
//	"fmt"
	"sync"
	"net/http"
	"github.com/gin-gonic/gin"
)

type CacheEntry struct {
	Value string
};

type Cache struct {
	data map[string]CacheEntry
	mu sync.RWMutex
};

func Init() *Cache {
	return &Cache{
		data: make(map[string]CacheEntry),
	}
}

func (c *Cache) set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = CacheEntry{Value: value}
}

func (c *Cache) get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key].Value
}

func main() {
	r := gin.Default()
	cache := Init()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is get request",
			"detail": "I am coming from server",
		})
	});
	r.GET("/set", func(c *gin.Context) {
		key := c.Query("key")
		value := c.Query("value")
		go cache.set(key, value)
		c.JSON(http.StatusOK, gin.H{
			"message": "Key updated in cache",
		})
	})
	r.GET("/get", func(c *gin.Context) {
		key := c.Query("key")
		value := cache.get(key)
		c.JSON(http.StatusOK, gin.H{
			"value": value,
		})
	})
	r.Run("0.0.0.0:5000")
}
