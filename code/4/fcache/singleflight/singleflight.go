package singleflight

import "sync"

// call represent the doing or done request
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// Group manage different key request
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do make sure all required key call fn once at once time
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	// protected m
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// if some goroutine has required key, wait for value
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// if the key is first required, create a new call and add wait group lock
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// only this call do fn
	c.val, c.err = fn()
	c.wg.Done()

	// delete the map of key to call
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
