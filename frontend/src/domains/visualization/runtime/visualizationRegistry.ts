import type {
  VisualizationCatalogEntry,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import type { VisualizationModuleOption } from '@/types'
import { getAmmActions } from './visualizationAmmActions'
import { getAttackActions } from './visualizationAttackActions'
import { getBlockchainActions } from './visualizationBlockchainActions'
import { getConsensusActions } from './visualizationConsensusActions'
import { getHashActions } from './visualizationCryptoActions'
import { getBridgeActions } from './visualizationCrosschainActions'
import { getDiscoveryActions } from './visualizationDiscoveryActions'
import { getCallTraceActions } from './visualizationEvmActions'
import { getTopologyActions } from './visualizationTopologyActions'

const ALL_VISUALIZATION_MODULE_KEYS = [
  'attacks/access_control',
  'attacks/attack_51',
  'attacks/bribery',
  'attacks/bridge_attack',
  'attacks/delegatecall_attack',
  'attacks/dos',
  'attacks/fake_deposit',
  'attacks/flashloan',
  'attacks/frontrun',
  'attacks/governance',
  'attacks/infinite_mint',
  'attacks/integer_overflow',
  'attacks/liquidation',
  'attacks/long_range',
  'attacks/nothing_at_stake',
  'attacks/oracle_manipulation',
  'attacks/reentrancy',
  'attacks/sandwich',
  'attacks/selfdestruct',
  'attacks/selfish_mining',
  'attacks/signature_replay',
  'attacks/timestamp_manipulation',
  'attacks/tx_origin',
  'attacks/weak_randomness',
  'blockchain/block_structure',
  'blockchain/chain_sync',
  'blockchain/difficulty',
  'blockchain/gas',
  'blockchain/light_client',
  'blockchain/mempool',
  'blockchain/merkle_tree',
  'blockchain/mining',
  'blockchain/state_trie',
  'blockchain/transaction',
  'blockchain/utxo',
  'blockchain/wallet',
  'consensus/dag',
  'consensus/dpos',
  'consensus/fork_choice',
  'consensus/hotstuff',
  'consensus/pbft',
  'consensus/pos',
  'consensus/pow',
  'consensus/raft',
  'consensus/tendermint',
  'consensus/vrf',
  'crosschain/atomic_swap',
  'crosschain/bridge',
  'crosschain/data_availability',
  'crosschain/finality',
  'crosschain/ibc',
  'crosschain/light_client',
  'crosschain/optimistic_rollup',
  'crosschain/oracle_bridge',
  'crosschain/plasma',
  'crosschain/relay',
  'crosschain/state_channel',
  'crosschain/zk_rollup',
  'crypto/asymmetric',
  'crypto/bls',
  'crypto/commitment',
  'crypto/elliptic_curve',
  'crypto/encoding',
  'crypto/hash',
  'crypto/kdf',
  'crypto/merkle_proof',
  'crypto/mpc',
  'crypto/randomness',
  'crypto/ring_sig',
  'crypto/signature',
  'crypto/symmetric',
  'crypto/threshold_sig',
  'crypto/zkp',
  'defi/amm',
  'defi/concentrated_liquidity',
  'defi/governance',
  'defi/insurance',
  'defi/interest_model',
  'defi/lending',
  'defi/liquidation',
  'defi/liquidity_pool',
  'defi/options',
  'defi/perpetual',
  'defi/stablecoin',
  'defi/ve_token',
  'defi/yield_aggregator',
  'evm/abi_codec',
  'evm/call_trace',
  'evm/create_create2',
  'evm/delegatecall',
  'evm/disassembler',
  'evm/event_log',
  'evm/evm_executor',
  'evm/gas_profiler',
  'evm/proxy_pattern',
  'evm/state_diff',
  'evm/storage_layout',
  'network/bgp_hijack',
  'network/block_propagation',
  'network/discovery',
  'network/eclipse_attack',
  'network/gossip',
  'network/kademlia',
  'network/nat_traversal',
  'network/partition',
  'network/sybil_attack',
  'network/topology',
] as const

const CATEGORY_LABELS: Record<string, string> = {
  attacks: '攻击演示',
  blockchain: '区块链基础',
  consensus: '共识算法',
  crosschain: '跨链机制',
  crypto: '密码学',
  defi: 'DeFi 机制',
  evm: 'EVM 执行',
  network: 'P2P 网络',
}

const SPECIAL_MODULE_LABELS: Record<string, string> = {
  'attacks/reentrancy': '重入攻击演示器',
  'blockchain/block_structure': '区块结构演示器',
  'blockchain/merkle_tree': 'Merkle 树演示器',
  'consensus/pbft': 'PBFT 共识实验',
  'consensus/raft': 'Raft 共识实验',
  'crypto/hash': '哈希算法演示器',
  'crosschain/bridge': '跨链桥演示器',
  'defi/amm': 'AMM 机制演示器',
  'evm/call_trace': '调用跟踪演示器',
  'network/discovery': '节点发现演示器',
  'network/topology': '网络拓扑演示器',
}

function formatModuleName(rawId: string): string {
  return rawId
    .split('_')
    .map((part) => {
      const upper = part.toUpperCase()
      return part === upper ? part : part.charAt(0).toUpperCase() + part.slice(1)
    })
    .join(' ')
}

function buildDefaultModuleLabel(moduleKey: string): string {
  const [category = '', rawId = moduleKey] = moduleKey.split('/')
  const categoryLabel = CATEGORY_LABELS[category] || '可视化模块'
  return `${formatModuleName(rawId)} ${categoryLabel}`
}

function buildDefaultModuleDescription(moduleKey: string): string {
  const [category = '', rawId = moduleKey] = moduleKey.split('/')
  const categoryLabel = CATEGORY_LABELS[category] || '可视化内容'
  return `通过 ${formatModuleName(rawId)} 场景观察 ${categoryLabel} 的核心过程、状态变化与事件流转。`
}

function createEntry(
  option: VisualizationModuleOption,
  runtime: Omit<VisualizationRuntimeSpec, 'moduleKey'>,
  getActions?: VisualizationCatalogEntry['getActions'],
): VisualizationCatalogEntry {
  return {
    option,
    runtime,
    getActions,
  }
}

function inferRenderer(category: string): VisualizationRuntimeSpec['renderer'] {
  switch (category) {
    case 'consensus':
      return 'consensus'
    case 'attacks':
      return 'attack'
    case 'blockchain':
      return 'blockchain'
    case 'network':
      return 'network'
    case 'crypto':
      return 'crypto'
    case 'crosschain':
      return 'crosschain'
    case 'evm':
      return 'evm'
    case 'defi':
      return 'defi'
    default:
      return 'blockchain'
  }
}

function inferMode(simulatorId: string): VisualizationRuntimeSpec['mode'] | undefined {
  switch (simulatorId) {
    case 'block_structure':
      return 'chain'
    case 'merkle_tree':
      return 'merkle'
    case 'state_trie':
      return 'state'
    case 'transaction':
      return 'transaction'
    case 'wallet':
      return 'wallet'
    case 'mining':
      return 'mining'
    case 'mempool':
      return 'mempool'
    case 'light_client':
      return 'light_client'
    default:
      return undefined
  }
}

function inferScenario(simulatorId: string): VisualizationRuntimeSpec['scenario'] | undefined {
  switch (simulatorId) {
    case 'topology':
      return 'topology'
    case 'discovery':
      return 'discovery'
    case 'gossip':
      return 'gossip'
    case 'partition':
      return 'partition'
    case 'kademlia':
      return 'kademlia'
    case 'block_propagation':
      return 'block_propagation'
    case 'eclipse_attack':
      return 'eclipse_attack'
    case 'sybil_attack':
      return 'sybil_attack'
    case 'bgp_hijack':
      return 'bgp_hijack'
    case 'nat_traversal':
      return 'nat_traversal'
    default:
      return undefined
  }
}

function getEnhancedActions(moduleKey: string, runtime: Omit<VisualizationRuntimeSpec, 'moduleKey'>) {
  if (runtime.renderer === 'consensus' && runtime.algorithm) {
    return () => getConsensusActions({ moduleKey, ...runtime }, true)
  }

  if (runtime.renderer === 'attack') {
    return () => getAttackActions({ moduleKey, ...runtime }, true)
  }

  switch (moduleKey) {
    case 'defi/amm':
      return () => getAmmActions()
    case 'blockchain/block_structure':
      return () => getBlockchainActions('chain')
    case 'blockchain/merkle_tree':
      return () => getBlockchainActions('merkle')
    case 'network/topology':
      return () => getTopologyActions()
    case 'network/discovery':
      return () => getDiscoveryActions()
    case 'crypto/hash':
      return () => getHashActions()
    case 'crosschain/bridge':
      return () => getBridgeActions()
    case 'evm/call_trace':
      return () => getCallTraceActions()
    default:
      return undefined
  }
}

function buildDefaultEntry(moduleKey: string): VisualizationCatalogEntry {
  const [category = '', simulatorId = moduleKey] = moduleKey.split('/')
  const runtime = {
    simulatorId,
    renderer: inferRenderer(category),
    algorithm: category === 'consensus' ? (simulatorId as VisualizationRuntimeSpec['algorithm']) : undefined,
    protocol: category === 'defi' ? (simulatorId as VisualizationRuntimeSpec['protocol']) : undefined,
    mode: category === 'blockchain' ? inferMode(simulatorId) : undefined,
    scenario: category === 'network' ? inferScenario(simulatorId) : undefined,
  } satisfies Omit<VisualizationRuntimeSpec, 'moduleKey'>

  return createEntry(
    {
      key: moduleKey,
      label: SPECIAL_MODULE_LABELS[moduleKey] || buildDefaultModuleLabel(moduleKey),
      description: buildDefaultModuleDescription(moduleKey),
      simulator_id: simulatorId,
    },
    runtime,
    getEnhancedActions(moduleKey, runtime),
  )
}

export const VISUALIZATION_CATALOG: Record<string, VisualizationCatalogEntry> = Object.fromEntries(
  ALL_VISUALIZATION_MODULE_KEYS.map((moduleKey) => [moduleKey, buildDefaultEntry(moduleKey)]),
)

export const VISUALIZATION_MODULE_OPTIONS: VisualizationModuleOption[] = Object.values(VISUALIZATION_CATALOG)
  .map((entry) => entry.option)
  .sort((left, right) => left.label.localeCompare(right.label, 'zh-CN'))

export function getVisualizationModuleOption(moduleKey?: string): VisualizationModuleOption | undefined {
  if (!moduleKey) {
    return undefined
  }
  return VISUALIZATION_CATALOG[moduleKey]?.option
}

export function resolveVisualizationModule(moduleKey?: string): VisualizationRuntimeSpec {
  const normalizedKey = moduleKey || 'blockchain/block_structure'
  const catalogEntry = VISUALIZATION_CATALOG[normalizedKey]

  if (!catalogEntry) {
    const [category = '', simulatorId = normalizedKey] = normalizedKey.split('/')
    return {
      moduleKey: normalizedKey,
      simulatorId,
      renderer: inferRenderer(category),
    }
  }

  return {
    moduleKey: normalizedKey,
    ...catalogEntry.runtime,
  }
}
