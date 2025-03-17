package main

import (
	"fmt"
	"github.com/univero/fcache/fcache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	fcache.NewGroup("scores", 2<<10, fcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := fcache.NewHttpPool(addr)
	log.Println("fcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
