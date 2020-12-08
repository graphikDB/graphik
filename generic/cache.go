package generic

import (
	"github.com/autom8ter/machine"
	"sync"
	"time"
)

type Cache struct {
	items sync.Map
}

type item struct {
	data    interface{}
	expires int64
}

func NewCache(m *machine.Machine, garbageCollect time.Duration) *Cache {
	cache := &Cache{
		items: sync.Map{},
	}
	m.Go(func(routine machine.Routine) {
		now := time.Now().UnixNano()
		cache.items.Range(func(key, value interface{}) bool {
			item := value.(item)
			if item.expires > 0 && now > item.expires {
				cache.items.Delete(key)
			}
			return true
		})
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(garbageCollect))))
	return cache
}

func (c *Cache) Get(key interface{}) (interface{}, bool) {
	obj, exists := c.items.Load(key)

	if !exists {
		return nil, false
	}

	item := obj.(item)

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		return nil, false
	}

	return item.data, true
}

func (c *Cache) Exists(key interface{}) bool {
	_, ok := c.Get(key)
	return ok
}

func (c *Cache) Set(key interface{}, value interface{}, duration time.Duration) {
	var expires int64

	if duration > 0 {
		expires = time.Now().Add(duration).UnixNano()
	}
	c.items.Store(key, item{
		data:    value,
		expires: expires,
	})
}

func (c *Cache) Range(f func(key, value interface{}) bool) {
	now := time.Now().UnixNano()

	fn := func(key, value interface{}) bool {
		item := value.(item)

		if item.expires > 0 && now > item.expires {
			return true
		}

		return f(key, item.data)
	}

	c.items.Range(fn)
}

func (c *Cache) Delete(key interface{}) {
	c.items.Delete(key)
}

func (c *Cache) Len() int {
	count := 0
	c.Range(func(key, value interface{}) bool {
		if value != nil {
			count++
		}
		return true
	})
	return count
}
