package fcache

import (
	"fmt"
	pb "github.com/univero/fcache/fcache/cachepb"
	"github.com/univero/fcache/fcache/hash"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_fcache/"
	defaultReplicas = 50
)

// HttpPool implements PeerPicker for a pool of Http peer
type HttpPool struct {
	// this peer's base URL, e.g. https://example.com:8080
	self     string
	basePath string
	mu       sync.Mutex
	// use consistent hash to choose node with the key
	peers *hash.Map
	// map node to its httpGetter
	httpGetter map[string]*httpGetter
}

// NewHttpPool return a HttpPool with defaultBasePath
func NewHttpPool(self string) *HttpPool {
	return &HttpPool{self: self, basePath: defaultBasePath}
}

// Log with server name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// Set updates the pool's list of peers
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = hash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetter = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetter[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer get the correct node according to the key
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetter[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HttpPool)(nil)

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
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(body)
	if err != nil {
		return
	}
}

// httpGetter is a client
type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v%v", h.baseURL, url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()))

	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("close body err %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("unmarshal response err %v", err)
	}
	return nil
}

// To verify httpGetter has implemented PeerGetter
// The statement is usually used to check the interface implement in the compile period
var _ PeerGetter = (*httpGetter)(nil)
