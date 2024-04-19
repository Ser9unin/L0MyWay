package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	storage "github.com/Ser9unin/L0MyWay/pkg/storage/db"
	lru "github.com/hashicorp/golang-lru"
)

type Cache struct {
	cache *lru.Cache
	size  int32
}

func NewCache(size int, db *sql.DB) (*Cache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}

	c := &Cache{
		cache: cache,
		size:  int32(size),
	}

	return c, nil
}

func (c *Cache) Store(key string, value json.RawMessage) {
	c.cache.ContainsOrAdd(key, value)
}

func (c *Cache) Get(ctx context.Context, db *sql.DB, key string) (json.RawMessage, bool, error) {
	var order_uid storage.DBOrder
	order_uid.OrderUID = key
	v, ok := c.cache.Get(key)
	if ok {
		log.Println("got cache hit")
	}
	if !ok {
		o, err := storage.GetOrderByID(ctx, db, order_uid)
		if errors.Is(err, sql.ErrNoRows) {
			return json.RawMessage{}, false, nil
		}
		if err != nil {
			return json.RawMessage{}, false, err
		}
		c.Store(o.OrderUID, o.Data)
		v = o.Data
	}

	order, _ := v.(json.RawMessage)

	return order, true, nil
}

func (c *Cache) Recover(ctx context.Context, db *sql.DB) error {
	params := storage.ListOrdersParams{
		ID:    0,
		Limit: c.size,
	}

	orders, err := storage.ListOrders(ctx, db, params)
	if err != nil {
		return err
	}

	for _, v := range orders {
		c.Store(v.OrderUID, v.Data)
	}

	return nil
}
