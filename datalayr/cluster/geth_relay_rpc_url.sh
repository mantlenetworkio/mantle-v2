name=$(kubectl get pod relay-0 -n relay -o=jsonpath='{.spec.nodeName}')
address=$(kubectl get node -o json | jq --arg v "$name" -r '.items[] | select(.metadata.name | test($v)).status.addresses' | jq -r '.[] | select(.type | test("ExternalIP")).address')
port=$(kubectl get svc relay-svc-0 -n relay -o yaml | yq -r '.spec.ports[] | select(.name == "http").nodePort')

echo $address:$port