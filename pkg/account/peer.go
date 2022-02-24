package account

import (
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerType int

const (
	_ PeerType = iota
	UnregisteredPeer
	WhiteListPeer
	RegisteredPeer
)

type Info struct {
	PeerType     PeerType
	AccountLevel int
}

func FetchPeerType(peerID peer.ID, registry *registry.Registry) *Info {
	peerType := UnregisteredPeer
	peerAccountLevel, _ := registry.ProviderAccountLevel(peerID)
	if peerAccountLevel != -1 {
		peerType = RegisteredPeer
	}
	if registry.IsTrusted(peerID) {
		peerType = WhiteListPeer
	}
	return &Info{
		PeerType:     peerType,
		AccountLevel: peerAccountLevel,
	}
}
