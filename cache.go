package main;

import (
	"fmt"
	"sync"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

type CacheEntry struct {
	Value string
	Expiration int64
	Expiry bool
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

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration := time.Now().AddDate(1,0,0).UnixNano()
	c.data[key] = CacheEntry{Value: value, Expiration: expiration, Expiry: false}
}

func (c *Cache) SetT(key, value string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration := time.Now().Add(duration).UnixNano()
	c.data[key] = CacheEntry{Value: value, Expiration: expiration, Expiry: true}
}



func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, found := c.data[key]
	if !found || (entry.Expiry && entry.Expiration < time.Now().UnixNano()) {
		return "", false
	}
	return entry.Value, true
}

func (c *Cache) CleanUp() {
	for {
		time.Sleep(time.Minute)
		now := time.Now().UnixNano()
		c.mu.Lock()
		for key, entry := range c.data {
			if now > entry.Expiration && entry.Expiry {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

func main() {
	r := gin.Default()
	cache := Init()
	go cache.CleanUp()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is get request",
			"detail": "I am coming from server",
		})
	});
	r.GET("/set", func(c *gin.Context) {
		key := c.Query("key")
		value := c.Query("value")
		durationstr, exists := c.GetQuery("duration")
		if exists {
			fmt.Println(durationstr, exists)
			duration, err := time.ParseDuration(durationstr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err,
				})
				return
			}
			go cache.SetT(key, value, duration)
		} else {
			go cache.Set(key, value)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Key updated in cache",
		})
	})
	r.GET("/get", func(c *gin.Context) {
		key := c.Query("key")
		value, found := cache.Get(key)
		if !found {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "key not present in cache",
				"found": found,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"value": value,
				"found": found,
			})
		}
	})
	r.Run("0.0.0.0:5000")
}
