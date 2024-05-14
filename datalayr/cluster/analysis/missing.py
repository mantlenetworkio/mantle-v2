#!/usr/bin/env python3

import sys
import json

num=1
node = {}
for i in range(num):
    with open('analysis/socket-' + str(i) + '.log') as f:
        data = json.load(f)
        operators = {}
        for d in data:
            operators[d['socket']] = True
        node[i] = operators

registered = {}
registered_ports = {}
for i, operators in node.items():
    for d in operators:
        a = d[d.index('&')+1: d.index('@')]
        b = d[d.index('@')+1:]
        registered_ports[a] = True
        registered_ports[b] = True

for i in range(num):
    ops = set(registered.keys()).difference(set(node[i].keys()))
    print('node', str(i), 'does not contain', ops)

sorted_keys = sorted(list(registered_ports.keys()))
print(sorted_keys)
p=32013
for i in range(200):
    if str(p+i) not in sorted_keys:
        print(p+i)
