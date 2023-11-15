import {
  predeploys as v1Predeploys,
  getDeployedContractDefinition,
} from '@ethan-bedrock/contracts'

import { predeploys as bedrockPredeploys } from '@ethan-bedrock/contracts-bedrock'

import {
  L1ChainID,
  L2ChainID,
  OEContractsLike,
  OEL1ContractsLike,
  OEL2ContractsLike,
  BridgeAdapterData,
} from '../interfaces'
import {
  StandardBridgeAdapter,
  DAIBridgeAdapter,
  ECOBridgeAdapter,
} from '../adapters'

export const DEPOSIT_CONFIRMATION_BLOCKS: {
  [ChainID in L2ChainID]: number
} = {
  [L2ChainID.MANTLE]: 50 as const,
  [L2ChainID.MANTLE_TESTNET]: 50 as const,
  [L2ChainID.MANTLE_GOERLIQA]: 12 as const,
  [L2ChainID.MANTLE_KOVAN]: 12 as const,
  [L2ChainID.MANTLE_HARDHAT_LOCAL]: 2 as const,
  [L2ChainID.MANTLE_HARDHAT_DEVNET]: 2 as const,
  [L2ChainID.MANTLE_V2_LOCAL_DEVNET]: 2 as const,

}

export const CHAIN_BLOCK_TIMES: {
  [ChainID in L1ChainID]: number
} = {
  [L1ChainID.MAINNET]: 13 as const,
  [L1ChainID.GOERLI]: 15 as const,
  [L1ChainID.HARDHAT_LOCAL]: 1 as const,
  [L1ChainID.BEDROCK_LOCAL_DEVNET]: 15 as const,
}

/**
 * Full list of default L2 contract addresses.
 * TODO(tynes): migrate to predeploys from contracts-bedrock
 */
export const DEFAULT_L2_CONTRACT_ADDRESSES: OEL2ContractsLike = {
  L2CrossDomainMessenger: v1Predeploys.L2CrossDomainMessenger,
  L2ToL1MessagePasser: v1Predeploys.BVM_L2ToL1MessagePasser,
  L2StandardBridge: v1Predeploys.L2StandardBridge,
  OVM_L1BlockNumber: v1Predeploys.BVM_L1BlockNumber,
  OVM_L2ToL1MessagePasser: v1Predeploys.BVM_L2ToL1MessagePasser,
  OVM_DeployerWhitelist: v1Predeploys.BVM_DeployerWhitelist,
  OVM_ETH: v1Predeploys.BVM_ETH,
  OVM_GasPriceOracle: v1Predeploys.BVM_GasPriceOracle,
  OVM_SequencerFeeVault: v1Predeploys.BVM_SequencerFeeVault,
  WETH: v1Predeploys.WETH9,
  BedrockMessagePasser: bedrockPredeploys.L2ToL1MessagePasser,
  BVM_MANTLE: v1Predeploys.LegacyERC20Mantle,
  TssRewardContract: v1Predeploys.TssRewardContract,

}

/**
 * Loads the L1 contracts for a given network by the network name.
 *
 * @param network The name of the network to load the contracts for.
 * @returns The L1 contracts for the given network.
 */
const getL1ContractsByNetworkName = (network: string): OEL1ContractsLike => {
  const getDeployedAddress = (name: string) => {
    return getDeployedContractDefinition(name, network).address
  }

  return {
    AddressManager: getDeployedAddress('Lib_AddressManager'),
    L1CrossDomainMessenger: getDeployedAddress(
      'Proxy__OVM_L1CrossDomainMessenger'
    ),
    L1StandardBridge: getDeployedAddress('Proxy__OVM_L1StandardBridge'),
    StateCommitmentChain: getDeployedAddress('StateCommitmentChain'),
    CanonicalTransactionChain: getDeployedAddress('CanonicalTransactionChain'),
    BondManager: getDeployedAddress('BondManager'),
    OptimismPortal: '0x5b47E1A08Ea6d985D6649300584e6722Ec4B1383' as const,
    L2OutputOracle: '0xE6Dfba0953616Bacab0c9A8ecb3a9BBa77FC15c0' as const,
    //TODO : unknown rollup address
    Rollup: '0xE6Dfba0953616Bacab0c9A8ecb3a9BBa77FC15c0' as const
  }
}

/**
 * Mapping of L1 chain IDs to the appropriate contract addresses for the OE deployments to the
 * given network. Simplifies the process of getting the correct contract addresses for a given
 * contract name.
 */
export const CONTRACT_ADDRESSES: {
  [ChainID in L2ChainID]: OEContractsLike
} = {
  [L2ChainID.MANTLE]: {
    l1: getL1ContractsByNetworkName('mainnet'),
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
  [L2ChainID.MANTLE_TESTNET]: {
    l1: getL1ContractsByNetworkName('goerli'),
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
  [L2ChainID.MANTLE_HARDHAT_LOCAL]: {
    l1: {
      AddressManager: '0x5FbDB2315678afecb367f032d93F642f64180aa3' as const,
      L1CrossDomainMessenger:
        '0x8A791620dd6260079BF849Dc5567aDC3F2FdC318' as const,
      L1StandardBridge: '0x610178dA211FEF7D417bC0e6FeD39F05609AD788' as const,
      StateCommitmentChain:
        '0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9' as const,
      CanonicalTransactionChain:
        '0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9' as const,
      BondManager: '0x5FC8d32690cc91D4c39d9d3abcBD16989F875707' as const,
      OptimismPortal: '0x0000000000000000000000000000000000000000' as const,
      L2OutputOracle: '0x0000000000000000000000000000000000000000' as const,
      Rollup: '0x0000000000000000000000000000000000000000' as const

    },
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
  [L2ChainID.MANTLE_HARDHAT_DEVNET]: {
    l1: {
      AddressManager:
        process.env.ADDRESS_MANAGER_ADDRESS ||
        ('0x92aBAD50368175785e4270ca9eFd169c949C4ce1' as const),
      L1CrossDomainMessenger:
        process.env.L1_CROSS_DOMAIN_MESSENGER_ADDRESS ||
        ('0x7959CF3b8ffC87Faca8aD8a1B5D95c0f58C0BEf8' as const),
      L1StandardBridge:
        process.env.L1_STANDARD_BRIDGE_ADDRESS ||
        ('0x8BAccFF561FDe61D6bC8B6f299fFBa561d2189B9' as const),
      StateCommitmentChain:
        process.env.STATE_COMMITMENT_CHAIN_ADDRESS ||
        ('0xd9e2F450525079e1e29fB23Bc7Caca6F61f8fD4a' as const),
      CanonicalTransactionChain:
        process.env.CANONICAL_TRANSACTION_CHAIN_ADDRESS ||
        ('0x0090171f848B2aa86918E5Ef2406Ab3d424fdd83' as const),
      BondManager:
        process.env.BOND_MANAGER_ADDRESS ||
        ('0x9faB987C9C469EB23Da31B7848B28aCf30905eA8' as const),
      Rollup:
        process.env.Rollup ||
        ('0x9faB987C9C469EB23Da31B7848B28aCf30905eA8' as const),
      OptimismPortal: '0x0000000000000000000000000000000000000000' as const,
      L2OutputOracle: '0x0000000000000000000000000000000000000000' as const,
    },
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
  [L2ChainID.MANTLE_V2_LOCAL_DEVNET]: {
    l1: {
      AddressManager: '0x6900000000000000000000000000000000000005' as const,
      L1CrossDomainMessenger:
        '0x6900000000000000000000000000000000000002' as const,
      L1StandardBridge: '0x6900000000000000000000000000000000000003' as const,
      StateCommitmentChain:
        '0x0000000000000000000000000000000000000000' as const,
      CanonicalTransactionChain:
        '0x0000000000000000000000000000000000000000' as const,
      BondManager: '0x0000000000000000000000000000000000000000' as const,
      OptimismPortal: '0x6900000000000000000000000000000000000001' as const,
      L2OutputOracle: '0x6900000000000000000000000000000000000000' as const,
      //TODO : Rollup contracts
      Rollup: '0x0000000000000000000000000000000000000000' as const
    },
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
  [L2ChainID.MANTLE_KOVAN]: {
    l1: {
      AddressManager: '0x100Dd3b414Df5BbA2B542864fF94aF8024aFdf3a' as const,
      L1CrossDomainMessenger:
        '0x4361d0F75A0186C05f971c566dC6bEa5957483fD' as const,
      L1StandardBridge: '0x22F24361D548e5FaAfb36d1437839f080363982B' as const,
      StateCommitmentChain:
        '0xD7754711773489F31A0602635f3F167826ce53C5' as const,
      CanonicalTransactionChain:
        '0xf7B88A133202d41Fe5E2Ab22e6309a1A4D50AF74' as const,
      BondManager: '0xc5a603d273E28185c18Ba4d26A0024B2d2F42740' as const,
      Rollup:
        process.env.Rollup ||
        ('0x9faB987C9C469EB23Da31B7848B28aCf30905eA8' as const),

      //bedrock part
      OptimismPortal:"0x6900000000000000000000000000000000000001" as const,
      L2OutputOracle:"0x6900000000000000000000000000000000000000" as const

    },
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
  [L2ChainID.MANTLE_GOERLIQA]: {
    l1: {
      AddressManager:
        process.env.ADDRESS_MANAGER_ADDRESS ||
        ('0x327903410307971Ca7Ba8A6CB2291D3b8825d7F5' as const),
      L1CrossDomainMessenger:
        process.env.L1_CROSS_DOMAIN_MESSENGER_ADDRESS ||
        ('0x3f41DAcb2dB659e45826126d004ad3E0C8eA680e' as const),
      L1StandardBridge:
        process.env.L1_STANDARD_BRIDGE_ADDRESS ||
        ('0x4cf99b9BC9B2Da64033D1Fb65146Ea60fbe8AD4B' as const),
      StateCommitmentChain:
        process.env.STATE_COMMITMENT_CHAIN_ADDRESS ||
        ('0x88EC574e2ef0EcF9043373139099f7E535F94dBC' as const),
      CanonicalTransactionChain:
        process.env.CANONICAL_TRANSACTION_CHAIN_ADDRESS ||
        ('0x258e80D5371fD7fFdDFE29E60b366f9FC44844c8' as const),
      BondManager:
        process.env.BOND_MANAGER_ADDRESS ||
        ('0xc723Cb5f3337c2F6Eab9b29E78CE42a28B8661d1' as const),
      Rollup:
        process.env.Rollup ||
        ('0x9faB987C9C469EB23Da31B7848B28aCf30905eA8' as const),
      //bedrock part
      OptimismPortal:"0x6900000000000000000000000000000000000001" as const,
      L2OutputOracle:"0x6900000000000000000000000000000000000000" as const
    },
    l2: DEFAULT_L2_CONTRACT_ADDRESSES,
  },
}

/**
 * Mapping of L1 chain IDs to the list of custom bridge addresses for each chain.
 */
export const BRIDGE_ADAPTER_DATA: {
  [ChainID in L2ChainID]?: BridgeAdapterData
} = {
  [L2ChainID.MANTLE]: {
    BitBTC: {
      Adapter: StandardBridgeAdapter,
      l1Bridge: '0xaBA2c5F108F7E820C049D5Af70B16ac266c8f128' as const,
      l2Bridge: '0x158F513096923fF2d3aab2BcF4478536de6725e2' as const,
    },
    DAI: {
      Adapter: StandardBridgeAdapter,
      l1Bridge: '0x10E6593CDda8c58a1d0f14C5164B376352a55f2F' as const,
      l2Bridge: '0x467194771dAe2967Aef3ECbEDD3Bf9a310C76C65' as const,
    },
  },
  [L2ChainID.MANTLE_KOVAN]: {
    wstETH: {
      Adapter: StandardBridgeAdapter,
      l1Bridge: '0xa88751C0a08623E11ff38c6B70F2BbEe7865C17c' as const,
      l2Bridge: '0xF9C842dE4381a70eB265d10CF8D43DceFF5bA935' as const,
    },
    BitBTC: {
      Adapter: StandardBridgeAdapter,
      l1Bridge: '0x0b651A42F32069d62d5ECf4f2a7e5Bd3E9438746' as const,
      l2Bridge: '0x0CFb46528a7002a7D8877a5F7a69b9AaF1A9058e' as const,
    },
    USX: {
      Adapter: StandardBridgeAdapter,
      l1Bridge: '0x40E862341b2416345F02c41Ac70df08525150dC7' as const,
      l2Bridge: '0xB4d37826b14Cd3CB7257A2A5094507d701fe715f' as const,
    },
    DAI: {
      Adapter: StandardBridgeAdapter,
      l1Bridge: '0xb415e822C4983ecD6B1c1596e8a5f976cf6CD9e3' as const,
      l2Bridge: '0x467194771dAe2967Aef3ECbEDD3Bf9a310C76C65' as const,
    },
  },
}
