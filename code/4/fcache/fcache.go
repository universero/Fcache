package fcache

import (
	"fmt"
	pb "github.com/univero/fcache/fcache/cachepb"
	"github.com/univero/fcache/fcache/singleflight"
	"log"
	"sync"
)

// A Getter loads data for a key
type Getter interface {
	// Get is a callback function,
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function.
// interface function, only one function in the interface can be used.
// with it, we can use both struct and func as the parameter.
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup initialise a group, and set it in the map called groups
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the group according to the name with read only mutex
func GetGroup(name string) *Group {
	// Read only lock
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get returns the value according to the key
// if the key is empty, it will return a new ByteView and log an error
// if the key doesn't exist, try to load it from the other data source
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Fcache] Fcache hit", key, "get", v)
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remoter peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeer called more than once")
	}
	g.peers = peers
}

// load data from other data source
// it will be expanded latter
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Fcache] Failed to get from peer", peer, "with key", key)
			}
		}
		return g.getLocally(key)
	})
	if err != nil {
		return viewi.(ByteView), err
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	resp := &pb.Response{}
	err := peer.Get(req, resp)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: resp.Value}, nil
}

// getLocally uses the getter to load the missing key
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: bytes}
	g.populateCache(key, value)
	return value, nil
}

// populateCache adds the new key-value in the cache
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
