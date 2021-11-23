package registry

import (
	"Pando/config"
	"Pando/internal/registry/discovery"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type mockDiscoverer struct {
	discoverRsp *discovery.Discovered
}

const (
	exceptID   = "12D3KooWK7CTS7cyWi51PeNE3cTjS2F2kDCZaQVU4A5xBmb9J1do"
	trustedID  = "12D3KooWSG3JuvEjRkSxt93ADTjQxqe4ExbBwSkQ9Zyk1WfBaZJF"
	trustedID2 = "12D3KooWKSNuuq77xqnpPLnU3fq1bTQW2TwSZL2Z4QTHEYpUVzfr"

	minerDiscoAddr = "stitest999999"
	minerAddr      = "/ip4/127.0.0.1/tcp/9999"
	minerAddr2     = "/ip4/127.0.0.2/tcp/9999"
)

var discoveryCfg = config.Discovery{
	Policy: config.Policy{
		Allow:       false,
		Except:      []string{exceptID, trustedID, trustedID2},
		Trust:       false,
		TrustExcept: []string{trustedID, trustedID2},
	},
	PollInterval:   config.Duration(time.Minute),
	RediscoverWait: config.Duration(time.Minute),
}

var aclCfg = config.AccountLevel{Threshold: []int{1, 10, 99}}

func newMockDiscoverer(t *testing.T, providerID string) *mockDiscoverer {
	peerID, err := peer.Decode(providerID)
	assert.NoError(t, err, "bad provider ID")

	maddr, err := multiaddr.NewMultiaddr(minerAddr)
	assert.NoError(t, err, "bad miner address")

	return &mockDiscoverer{
		discoverRsp: &discovery.Discovered{
			AddrInfo: peer.AddrInfo{
				ID:    peerID,
				Addrs: []multiaddr.Multiaddr{maddr},
			},
			Type: discovery.MinerType,
		},
	}
}

func (m *mockDiscoverer) Discover(ctx context.Context, peerID peer.ID, filecoinAddr string) (*discovery.Discovered, error) {
	if filecoinAddr == "bad1234" {
		return nil, errors.New("unknown miner")
	}

	return m.discoverRsp, nil
}

func TestNewRegistryDiscovery(t *testing.T) {
	mockDisco := newMockDiscoverer(t, exceptID)

	r, err := NewRegistry(&discoveryCfg, &aclCfg, nil, mockDisco)
	assert.NoError(t, err)

	t.Log("created new registry")

	peerID, err := peer.Decode(trustedID)
	assert.NoError(t, err, "bad provider ID")

	maddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3002")
	assert.NoErrorf(t, err, "Cannot create multiaddr: %s", err)
	info := &ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    peerID,
			Addrs: []multiaddr.Multiaddr{maddr},
		},
	}

	err = r.Register(info)
	if err != nil {
		t.Error("failed to register directly:", err)
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Check that 2nd call to Close is ok
	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDiscoveryAllowed(t *testing.T) {
	mockDisco := newMockDiscoverer(t, exceptID)

	r, err := NewRegistry(&discoveryCfg, &aclCfg, nil, mockDisco)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	t.Log("created new registry")

	peerID, err := peer.Decode(exceptID)
	if err != nil {
		t.Fatal("bad provider ID:", err)
	}

	err = r.Discover(peerID, minerDiscoAddr, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("discovered mock miner", minerDiscoAddr)

	info := r.ProviderInfoByAddr(minerDiscoAddr)
	if info == nil {
		t.Fatal("did not get provider info for miner")
	}
	t.Log("got provider info for miner")

	assert.Equal(t, info.AddrInfo.ID, peerID, "did not get correct porvider id")

	peerID, err = peer.Decode(trustedID)
	if err != nil {
		t.Fatal("bad provider ID:", err)
	}
	maddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3002")
	if err != nil {
		t.Fatalf("Cannot create multiaddr: %s", err)
	}
	info = &ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    peerID,
			Addrs: []multiaddr.Multiaddr{maddr},
		},
	}
	err = r.Register(info)
	if err != nil {
		t.Error("failed to register directly:", err)
	}

	infos := r.AllProviderInfo()
	if len(infos) != 2 {
		t.Fatal("expected 2 provider infos")
	}

}

func TestDiscoveryBlocked(t *testing.T) {
	mockDisco := newMockDiscoverer(t, exceptID)

	peerID, err := peer.Decode(exceptID)
	if err != nil {
		t.Fatal("bad provider ID:", err)
	}

	discoveryCfg.Policy.Allow = true
	defer func() {
		discoveryCfg.Policy.Allow = false
	}()

	r, err := NewRegistry(&discoveryCfg, &aclCfg, nil, mockDisco)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	t.Log("created new registry")

	err = r.Discover(peerID, minerDiscoAddr, true)
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatal("expected error:", ErrNotAllowed, "got:", err)
	}

	into := r.ProviderInfoByAddr(minerDiscoAddr)
	if into != nil {
		t.Error("should not have found provider info for miner")
	}
}

func TestDatastore(t *testing.T) {
	dataStorePath := t.TempDir()
	mockDisco := newMockDiscoverer(t, exceptID)

	peerID, err := peer.Decode(trustedID)
	if err != nil {
		t.Fatal("bad provider ID:", err)
	}
	maddr, err := multiaddr.NewMultiaddr(minerAddr)
	if err != nil {
		t.Fatal("bad miner address:", err)
	}
	info1 := &ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    peerID,
			Addrs: []multiaddr.Multiaddr{maddr},
		},
	}
	peerID, err = peer.Decode(trustedID2)
	if err != nil {
		t.Fatal("bad provider ID:", err)
	}
	maddr, err = multiaddr.NewMultiaddr(minerAddr2)
	if err != nil {
		t.Fatal("bad miner address:", err)
	}
	info2 := &ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    peerID,
			Addrs: []multiaddr.Multiaddr{maddr},
		},
	}

	// Create datastore
	dstore, err := leveldb.NewDatastore(dataStorePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRegistry(&discoveryCfg, &aclCfg, dstore, mockDisco)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("created new registry with datastore")

	err = r.Register(info1)
	if err != nil {
		t.Fatal("failed to register directly:", err)
	}
	err = r.Register(info2)
	if err != nil {
		t.Fatal("failed to register directly:", err)
	}

	pinfo := r.ProviderInfo(peerID)
	if pinfo == nil {
		t.Fatal("did not find registered provider")
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Create datastore
	dstore, err = leveldb.NewDatastore(dataStorePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	r, err = NewRegistry(&discoveryCfg, &aclCfg, dstore, mockDisco)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("re-created new registry with datastore")

	infos := r.AllProviderInfo()
	if len(infos) != 2 {
		t.Fatal("expected 2 provider infos")
	}

	for i := range infos {
		pid := infos[i].AddrInfo.ID
		if pid != info1.AddrInfo.ID && pid != info2.AddrInfo.ID {
			t.Fatalf("loaded invalid provider ID: %s", pid)
		}
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}
}
