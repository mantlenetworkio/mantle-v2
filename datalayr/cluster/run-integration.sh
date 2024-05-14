set -e

if [ "$#" -ne 1 ]; then
    echo "need cluster name"
    exit 0
fi

# Get disperser external ip
echo Dispersers
name=$(kubectl get pod dl-disperser-0 -o=jsonpath='{.spec.nodeName}')
address=$(kubectl get node -o json | jq --arg v "$name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
port=$(kubectl get cm dl-disperser-0 -o yaml | yq -r '.data.DL_DISPERSER_GRPC_PORT')
dis_socket=$address:$port
./wait-for $dis_socket -t 180
echo "Disperser socket: $dis_socket"

# Get retriever external ip
echo Retrievers
name=$(kubectl get pod dl-retriever-0 -o=jsonpath='{.spec.nodeName}')
address=$(kubectl get node -o json | jq --arg v "$name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
port=$(kubectl get cm dl-retriever-0 -o yaml | yq -r '.data.DL_RETRIEVER_GRPC_PORT')
ret_socket=$address:$port
./wait-for $ret_socket -t 180
echo "Retriever socket: $ret_socket"

# Get node external ips
echo Nodes

nodes=$(kubectl get pods -l app=dl-node --no-headers -o custom-columns=":metadata.name")
for entry in $nodes; do
    node_name=$(kubectl get pod $entry -o=jsonpath='{.spec.nodeName}')
    address=$(kubectl get node -o json | jq --arg v "$node_name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
    port=$(kubectl get cm $entry -o yaml | yq -r '.data.DL_NODE_GRPC_PORT')
    node_socket=$address:$port
    echo "query" $node_socket
    ./wait-for $node_socket -t 180
    echo $entry: $node_socket
done

pushd ..
    ginkgo ./integration -- --new-experiment=false --config=../cluster/kustomize/overlay/$1 --dis-socket=$dis_socket --ret-socket=$ret_socket
popd
