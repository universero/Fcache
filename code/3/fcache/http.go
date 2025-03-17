package fcache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_fcache/"

// HttpPool implements PeerPicker for a pool of Http peer
type HttpPool struct {
	// this peer's base URL, e.g. https://example.com:8080
	self     string
	basePath string
}

// NewHttpPool return a HttpPool with defaultBasePath
func NewHttpPool(self string) *HttpPool {
	return &HttpPool{self: self, basePath: defaultBasePath}
}

// Log with server name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all request
// get the value by /<basePath>/<groupName>/<key>
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check the path
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HttpPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	fmt.Printf("try to get [group] %s [key] %s\n", groupName, key)

	// get the according group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// get value of the key
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write value as response
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(view.ByteSlice())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
