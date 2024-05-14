set -e

if [ "$#" -ne 2 ]; then
    echo "need cluster name and region"
    exit 0
fi

cluster_name=$1
cluster_region=$2

aws eks update-kubeconfig --name $cluster_name --region $cluster_region --role-arn arn:aws:iam::844660611549:role/$cluster_name-datalayr-admin-access

for entry in $(kubectl get pods -o=name | grep dl-retriever); do
    name="${entry##*/}"
    node_name=$(kubectl get pod $name -o=jsonpath='{.spec.nodeName}')
    address=$(kubectl get node -o json | jq --arg v "$node_name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
    port=$(kubectl get cm $name -o yaml | yq -r '.data.DL_RETRIEVER_GRPC_PORT')
    socket=$address:$port
    echo $name: $socket
done

for entry in $(kubectl get pods -o=name | grep dl-disperser); do
    name="${entry##*/}"
    node_name=$(kubectl get pod $name -o=jsonpath='{.spec.nodeName}')
    address=$(kubectl get node -o json | jq --arg v "$node_name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
    port=$(kubectl get cm $name -o yaml | yq -r '.data.DL_DISPERSER_GRPC_PORT')
    socket=$address:$port
    echo $name: $socket
done

name=$(kubectl get pod graph-0 -o=jsonpath='{.spec.nodeName}')
address=$(kubectl get node -o json | jq --arg v "$name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
port=$(kubectl get svc graph-svc-0 -o yaml | yq -r '.spec.ports[] | select(.name | test("http")).nodePort')
socket=$address:$port/subgraphs/name/datalayr/
echo "graph-0: $socket"

aws eks update-kubeconfig --name chain --region us-west-2 --role-arn arn:aws:iam::844660611549:role/chain-datalayr-admin-access

name=$(kubectl get pod relay-0 -n relay -o=jsonpath='{.spec.nodeName}')
address=$(kubectl get node -n relay -o json | jq --arg v "$name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
port=$(kubectl get svc relay-svc-0 -n relay -o yaml | yq -r '.spec.ports[] | select(.name | test("http")).nodePort')
socket=$address:$port
echo "relay-0: $socket"

aws eks update-kubeconfig --name $cluster_name --region $cluster_region --role-arn arn:aws:iam::844660611549:role/$cluster_name-datalayr-admin-access