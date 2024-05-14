#!/bin/sh

function addPeer {
	SEALER_ENODE=$1
	i=0;
	sealer_quote=$(printf \"%s\" ${SEALER_ENODE})
	add_peer_cmd=$(printf 'admin.addPeer(%s)' enode)
	echo $add_peer_cmd >> /add_peer_cmd
	while true; do
		geth attach --exec "enode = ${sealer_quote} ; ${add_peer_cmd}" /relay-data/geth.ipc
		peerCount=$(geth attach --exec 'net.peerCount' /relay-data/geth.ipc)
		if [ "${peerCount}" -gt 0 ]; then
			break;
		fi
		echo "geth attach --exec ${add_peer_cmd} /relay-data/geth.ipc" > "/cmd-$i"
		sleep 3
		i=$(( i + 1 ))
	done
}

P2P_PORT=$(cat /p2p_port)
SEALER_ENODE=$(cat /sealer_enode)

geth --datadir=/relay-data init /chain-genesis/genesis.json
addPeer ${SEALER_ENODE} &
geth --networkid=40525 --nodiscover --http.vhosts=*  --syncmode=full --gcmode=archive --http --http.addr=0.0.0.0 --http.api net,eth,web3,txpool,debug  --datadir=/relay-data --bootnodes=${SEALER_ENODE} --port=${P2P_PORT}

