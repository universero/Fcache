package fcache

import (
	"fmt"
	"log"
	"testing"
)

// Simulate a time-consuming db with map
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	fcache := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key:", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key %s not found", key)
		}))
	for k, v := range db {
		if view, err := fcache.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get the key:", k)
		}
		if _, err := fcache.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatal("failed to get the key:", k)
		}
	}

	if view, err := fcache.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
