package lrucache

import (
	"github.com/hashicorp/golang-lru"
	"time"
)

type Item struct {
	Value      []byte
	Expiration int64
}

type Storage struct {
	lru *lru.Cache
}

func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

func NewStorage(cacheLimit int) (*Storage, error) {
	var c *lru.Cache
	var err error
	if c, err = lru.New(cacheLimit); err != nil {
		return nil, err
	}
	return &Storage{
		lru: c,
	}, nil
}

func (s Storage) Get(key string) []byte {
	if item, ok := s.lru.Get(key); ok {
		it := item.(Item)
		if it.Expired() {
			s.lru.Remove(key)
			return nil
		}
		return it.Value
	}
	return nil

}

func (s Storage) Set(key string, value []byte, duration time.Duration) {
	item := Item{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
	s.lru.Add(key, item)
}
