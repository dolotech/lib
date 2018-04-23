package common

import (
	"sort"
	"sync"
)

type Cache struct {
	sync.Mutex
	cache      [][]byte
	size       int
	statistics bool
	count      int
}

func NewCache(s int) *Cache {
	return &Cache{size: s}
}

func (this *Cache) Statistics() {
	this.Lock()
	this.statistics = true
	this.count = 0
	this.Unlock()
}

func (this *Cache) CloseStatistics() int {
	this.Lock()
	this.statistics = false
	count := this.count
	this.count = 0
	this.Unlock()
	return count
}

func (this *Cache) Get(n int) (r []byte) {
	this.Lock()
	defer this.Unlock()
	if this.statistics {
		this.count += n
	}
	lens := len(this.cache)
	if lens == 0 {
		return make([]byte, n, overCommit(n))
	}
	i := sort.Search(lens, func(x int) bool { return len(this.cache[x]) >= n })
	if i == lens {
		i--
		this.cache[i] = make([]byte, n, overCommit(n))
	}
	r = this.cache[i][:n]
	copy(this.cache[i:], this.cache[i+1:])
	this.cache[lens-1] = nil
	this.cache = this.cache[:lens-1]
	return
}

func (this *Cache) Put(b []byte) {
	this.Lock()
	defer this.Unlock()
	b = b[:cap(b)]
	lenb := len(b)
	if lenb == 0 {
		return
	}
	lens := len(this.cache)
	if lens >= this.size {
		return
	}
	i := sort.Search(lens, func(x int) bool { return len(this.cache[x]) >= lenb })
	this.cache = append(this.cache, nil)
	copy(this.cache[i+1:], this.cache[i:])
	this.cache[i] = b
}

func overCommit(n int) int {
	switch {
	case n < 8:
		return 8
	case n < 1e5:
		return 2 * n
	case n < 1e6:
		return 3 * n / 2
	default:
		return n
	}
}
