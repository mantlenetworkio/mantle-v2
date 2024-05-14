#!/bin/bash

kubectl exec graph-0 -c graph-node -- curl -s -X POST  -H "Content-Type: application/json" -d  '{"query": "query {operators { socket} }" }' http://127.0.0.1:8000/subgraphs/name/datalayr | jq '.data.operators' > analysis/socket-0.log

