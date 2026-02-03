package p2p

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/p2p/store"

	"github.com/ethereum/go-ethereum/log"
)

type NotificationsMetricer interface {
	IncPeerCount()
	DecPeerCount()
	IncStreamCount()
	DecStreamCount()
}

type notifications struct {
	log log.Logger
	m   NotificationsMetricer
	isStatic func(peer.ID) bool
	md       store.MetadataStore
}

func (notif *notifications) Listen(n network.Network, a ma.Multiaddr) {
	notif.log.Info("started listening network address", "addr", a)
}

func (notif *notifications) ListenClose(n network.Network, a ma.Multiaddr) {
	notif.log.Info("stopped listening network address", "addr", a)
}

func (notif *notifications) Connected(n network.Network, v network.Conn) {
	notif.m.IncPeerCount()
	source, opstackID := notif.peerSource(v.RemotePeer(), v.Stat().Direction)
	notif.log.Debug("connected to peer", "peer", v.RemotePeer(), "addr", v.RemoteMultiaddr(), "dir", v.Stat().Direction, "source", source, "opstack_id", opstackID)
}

func (notif *notifications) Disconnected(n network.Network, v network.Conn) {
	notif.m.DecPeerCount()
	source, opstackID := notif.peerSource(v.RemotePeer(), v.Stat().Direction)
	notif.log.Debug("disconnected from peer", "peer", v.RemotePeer(), "addr", v.RemoteMultiaddr(), "dir", v.Stat().Direction, "source", source, "opstack_id", opstackID)
}

func (notif *notifications) OpenedStream(n network.Network, v network.Stream) {
	notif.m.IncStreamCount()
	c := v.Conn()
	notif.log.Trace("opened stream", "protocol", v.Protocol(), "peer", c.RemotePeer(), "addr", c.RemoteMultiaddr())
}

func (notif *notifications) ClosedStream(n network.Network, v network.Stream) {
	notif.m.DecStreamCount()
	c := v.Conn()
	notif.log.Trace("opened stream", "protocol", v.Protocol(), "peer", c.RemotePeer(), "addr", c.RemoteMultiaddr())
}

func (notif *notifications) peerSource(id peer.ID, dir network.Direction) (string, uint64) {
	if notif.isStatic != nil && notif.isStatic(id) {
		return "static", 0
	}
	if notif.md != nil {
		md, err := notif.md.GetPeerMetadata(id)
		if err == nil && (md.ENR != "" || md.OPStackID != 0) {
			return "discovery", md.OPStackID
		}
	}
	switch dir {
	case network.DirInbound:
		return "inbound", 0
	case network.DirOutbound:
		return "outbound", 0
	default:
		return "unknown", 0
	}
}

func NewNetworkNotifier(log log.Logger, m metrics.Metricer, isStatic func(peer.ID) bool, md store.MetadataStore) network.Notifiee {
	if m == nil {
		m = metrics.NoopMetrics
	}
	return &notifications{log: log, m: m, isStatic: isStatic, md: md}
}
