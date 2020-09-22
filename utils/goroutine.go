package utils

import (
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	// DefaultAntsPoolSize sets up the capacity of worker pool, 256 * 1024.
	DefaultAntsPoolSize = 10 * 1024

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 5 * time.Second

	// Nonblocking decides what to do when submitting a new job to a full worker pool: waiting for a available worker
	// or returning nil directly.
	Nonblocking = true
)

//func init() {
//// It releases the default pool from ants.
//ants.Release()
//
//options := ants.Options{ExpiryDuration: ExpiryDuration, Nonblocking: Nonblocking}
//Go, _ = ants.NewPool(DefaultAntsPoolSize, ants.WithOptions(options))
//
//}

// Pool is the alias of ants.Pool.
type Pool = ants.Pool

//var Go *Pool

// Default instantiates a non-blocking *WorkerPool with the capacity of DefaultAntsPoolSize.
//func Go() *Pool {
//	options := ants.Options{ExpiryDuration: ExpiryDuration, Nonblocking: Nonblocking}
//	defaultAntsPool, _ := ants.NewPool(DefaultAntsPoolSize, ants.WithOptions(options))
//	return defaultAntsPool
//}
