import argparse
import logging
import os
import subprocess
import json
import socket
import time
import shutil
import copy
from urllib.request import urlopen, Request
from urllib.error import URLError

import testnet.log_setup
from testnet.genesis import GENESIS_TMPL

parser = argparse.ArgumentParser(description='Bedrock testnet launcher')
parser.add_argument('--monorepo-dir', help='Directory of the monorepo', default=os.getcwd())
parser.add_argument('--rpc-url', help='RPC URL of the L1 network', default='http://localhost:8545')



log = logging.getLogger()


def main():
    args = parser.parse_args()
    rpc_url = args.rpc_url

    pjoin = os.path.join
    monorepo_dir = os.path.abspath(args.monorepo_dir)
    testnet_dir = pjoin(monorepo_dir, '.testnet')
    ops_bedrock_dir = pjoin(monorepo_dir, 'ops-bedrock')
    contracts_bedrock_dir = pjoin(monorepo_dir, 'packages', 'contracts-bedrock')
    deployment_dir = pjoin(contracts_bedrock_dir, 'deployments')
    op_node_dir = pjoin(args.monorepo_dir, 'op-node')
    genesis_l2_path = pjoin(testnet_dir, 'genesis-l2.json')
    addresses_json_path = pjoin(testnet_dir, 'addresses.json')
    sdk_addresses_json_path = pjoin(testnet_dir, 'sdk-addresses.json')
    rollup_config_path = pjoin(testnet_dir, 'rollup.json')
    os.makedirs(testnet_dir, exist_ok=True)

    # Fetch chain ID from localhost:8545 AFTER L1 is up
    log.info(f'Fetching chain ID from l1 endpoint: {rpc_url}...')
    chain_id = get_chain_id_from_rpc(rpc_url)
    if chain_id is None:
        log.error(f'Failed to fetch chain ID from l1 endpoint: {rpc_url} after L1 startup')
        raise Exception('Cannot determine L1 chain ID')
    log.info(f'Using chain ID {chain_id} from l1 endpoint: {rpc_url}')

    # Update deployment_json_path with actual chain_id
    deployment_json_path = pjoin(deployment_dir, f'{chain_id}-deploy.json')

    log.info('Generating network config.')
    testnet_cfg_orig = pjoin(contracts_bedrock_dir, 'deploy-config', 'mantle-testnet.json')
    testnet_cfg_backup = pjoin(testnet_dir, 'mantle-testnet.json.bak')
    shutil.copy(testnet_cfg_orig, testnet_cfg_backup)
    deploy_config = read_json(testnet_cfg_orig)
    deploy_config['l1GenesisBlockTimestamp'] = GENESIS_TMPL['timestamp']
    deploy_config['l1StartingBlockTag'] = 'earliest'
    log_print(deploy_config)
    write_json(testnet_cfg_orig, deploy_config)

    if os.path.exists(addresses_json_path):
        log.info('Contracts already deployed.')
        addresses = read_json(addresses_json_path)
    else:
        log.info('Deploying contracts.')
        run_command(
          [
            'forge', 'script', 'scripts/deploy/Deploy.s.sol',
            '--rpc-url', rpc_url,
            '--private-key', '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80',
            '--broadcast'
          ],
          env={'DEPLOY_CONFIG_PATH': testnet_cfg_orig},
          cwd=contracts_bedrock_dir
        )
        addresses = read_json(deployment_json_path)
        sdk_addresses = {}
        sdk_addresses.update({
            'AddressManager': '0x0000000000000000000000000000000000000000',
            'StateCommitmentChain': '0x0000000000000000000000000000000000000000',
            'CanonicalTransactionChain': '0x0000000000000000000000000000000000000000',
            'BondManager': '0x0000000000000000000000000000000000000000',
        })
        sdk_addresses['L1CrossDomainMessenger'] = addresses['L1CrossDomainMessengerProxy']
        sdk_addresses['L1StandardBridge'] = addresses['L1StandardBridgeProxy']
        sdk_addresses['OptimismPortal'] = addresses['OptimismPortalProxy']
        sdk_addresses['L2OutputOracle'] = addresses['L2OutputOracleProxy']
        write_json(addresses_json_path, addresses)
        write_json(sdk_addresses_json_path, sdk_addresses)

    if os.path.exists(genesis_l2_path):
        log.info('L2 genesis and rollup configs already generated.')
    else:
        log.info('Generating L2 genesis and rollup configs.')
        run_command([
            'go', 'run', 'cmd/main.go', 'genesis', 'l2',
            '--l1-rpc', rpc_url,
            '--deploy-config', testnet_cfg_orig,
            '--l1-deployments', deployment_json_path,
            '--outfile.l2', pjoin(testnet_dir, 'genesis-l2.json'),
            '--outfile.rollup', pjoin(testnet_dir, 'rollup.json')
        ], cwd=op_node_dir)

    # rollup_config = read_json(rollup_config_path)

    # if os.path.exists(testnet_cfg_backup):
    #     shutil.move(testnet_cfg_backup, testnet_cfg_orig)

    # log.info('Bringing up L2.')
    # run_command(['mockdockercompose', 'up', '-d', 'l2'], cwd=ops_bedrock_dir, env={
    #     'PWD': ops_bedrock_dir
    # })
    # wait_up(9545)

    # log.info('Bringing up everything else.')
    # run_command(['mockdockercompose', 'up', '-d', 'op-node', 'op-proposer', 'op-batcher'], cwd=ops_bedrock_dir, env={
    #     'PWD': ops_bedrock_dir,
    #     'L2OO_ADDRESS': addresses['L2OutputOracleProxy'],
    #     'SEQUENCER_BATCH_INBOX_ADDRESS': rollup_config['batch_inbox_address']
    # })

    # log.info('Testnet ready.')


def get_chain_id_from_rpc(rpc_url='http://localhost:8545'):
    """Fetch chain ID from an Ethereum RPC endpoint."""
    try:
        data = json.dumps({
            'jsonrpc': '2.0',
            'method': 'eth_chainId',
            'params': [],
            'id': 1
        }).encode('utf-8')

        req = Request(rpc_url, data=data, headers={'Content-Type': 'application/json'})
        with urlopen(req, timeout=5) as response:
            result = json.loads(response.read().decode('utf-8'))
            chain_id_hex = result.get('result')
            if chain_id_hex:
                return int(chain_id_hex, 16)
            else:
                log.error(f'No chain ID in response: {result}')
                return None
    except Exception as e:
        log.error(f'Failed to fetch chain ID from {rpc_url}: {e}')
        return None


def run_command(args, check=True, shell=False, cwd=None, env=None):
    env = env if env else {}
    return subprocess.run(
        args,
        check=check,
        shell=shell,
        env={
            **os.environ,
            **env
        },
        cwd=cwd
    )


def wait_up(port, retries=10, wait_secs=1):
    for i in range(0, retries):
        log.info(f'Trying 127.0.0.1:{port}')
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            s.connect(('127.0.0.1', int(port)))
            s.shutdown(2)
            log.info(f'Connected 127.0.0.1:{port}')
            return True
        except Exception:
            time.sleep(wait_secs)

    raise Exception(f'Timed out waiting for port {port}.')


def write_json(path, data):
    with open(path, 'w+') as f:
        json.dump(data, f, indent='  ')

def log_print(msg):
    print('=' * 20)
    print(msg)
    print('=' * 20)

def read_json(path):
    with open(path, 'r') as f:
        return json.load(f)
