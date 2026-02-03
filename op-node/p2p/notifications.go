package p2p

import (
	"strings"

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
	lookupENR func(peer.ID) (string, uint64)
}

func (notif *notifications) Listen(n network.Network, a ma.Multiaddr) {
	notif.log.Info("started listening network address", "addr", a)
}

func (notif *notifications) ListenClose(n network.Network, a ma.Multiaddr) {
	notif.log.Info("stopped listening network address", "addr", a)
}

func (notif *notifications) Connected(n network.Network, v network.Conn) {
	notif.m.IncPeerCount()
	source, opstackID, enr, enrSource, addrCount, addrSample, isStatic := notif.peerDetails(n, v.RemotePeer(), v.Stat().Direction)
	notif.log.Debug("connected to peer",
		"peer", v.RemotePeer(),
		"addr", v.RemoteMultiaddr(),
		"dir", v.Stat().Direction,
		"source", source,
		"is_static", isStatic,
		"opstack_id", opstackID,
		"enr", enr,
		"enr_source", enrSource,
		"peerstore_addrs", addrSample,
		"peerstore_addr_count", addrCount,
	)
}

func (notif *notifications) Disconnected(n network.Network, v network.Conn) {
	notif.m.DecPeerCount()
	source, opstackID, enr, enrSource, addrCount, addrSample, isStatic := notif.peerDetails(n, v.RemotePeer(), v.Stat().Direction)
	notif.log.Debug("disconnected from peer",
		"peer", v.RemotePeer(),
		"addr", v.RemoteMultiaddr(),
		"dir", v.Stat().Direction,
		"source", source,
		"is_static", isStatic,
		"opstack_id", opstackID,
		"enr", enr,
		"enr_source", enrSource,
		"peerstore_addrs", addrSample,
		"peerstore_addr_count", addrCount,
	)
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

func (notif *notifications) peerDetails(n network.Network, id peer.ID, dir network.Direction) (string, uint64, string, string, int, string, bool) {
	isStatic := notif.isStatic != nil && notif.isStatic(id)
	opstackID := uint64(0)
	enr := ""
	enrSource := ""

	if notif.md != nil {
		md, err := notif.md.GetPeerMetadata(id)
		if err == nil && (md.ENR != "" || md.OPStackID != 0) {
			opstackID = md.OPStackID
			enr = md.ENR
			enrSource = "peerstore"
		}
	}
	if enrSource == "" && notif.lookupENR != nil {
		if foundENR, foundID := notif.lookupENR(id); foundENR != "" {
			enr = foundENR
			opstackID = foundID
			enrSource = "discv5_table"
		}
	}

	if isStatic {
		return "static", opstackID, enr, enrSource, peerAddrsCount(n, id), peerAddrsSample(n, id), true
	}
	if opstackID != 0 || enr != "" {
		return "discovery", opstackID, enr, enrSource, peerAddrsCount(n, id), peerAddrsSample(n, id), false
	}
	switch dir {
	case network.DirInbound:
		return "inbound", 0, "", "", peerAddrsCount(n, id), peerAddrsSample(n, id), false
	case network.DirOutbound:
		return "outbound", 0, "", "", peerAddrsCount(n, id), peerAddrsSample(n, id), false
	default:
		return "unknown", 0, "", "", peerAddrsCount(n, id), peerAddrsSample(n, id), false
	}
}

func NewNetworkNotifier(log log.Logger, m metrics.Metricer, isStatic func(peer.ID) bool, md store.MetadataStore, lookupENR func(peer.ID) (string, uint64)) network.Notifiee {
	if m == nil {
		m = metrics.NoopMetrics
	}
	return &notifications{log: log, m: m, isStatic: isStatic, md: md, lookupENR: lookupENR}
}

func peerAddrsCount(n network.Network, id peer.ID) int {
	if n == nil {
		return 0
	}
	return len(n.Peerstore().Addrs(id))
}

func peerAddrsSample(n network.Network, id peer.ID) string {
	if n == nil {
		return ""
	}
	addrs := n.Peerstore().Addrs(id)
	if len(addrs) == 0 {
		return ""
	}
	limit := 3
	if len(addrs) < limit {
		limit = len(addrs)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, addrs[i].String())
	}
	return strings.Join(out, ",")
}
