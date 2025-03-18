package fcache

import pb "github.com/univero/fcache/fcache/cachepb"

// PeerPicker is the interface that must be implemented to
// locate the peer that owns a specify key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer
// And get cache from a peer
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
