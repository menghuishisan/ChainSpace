import type { SimulatorState } from './simulation'

export type VisualizationRenderer =
  | 'consensus'
  | 'attack'
  | 'blockchain'
  | 'network'
  | 'crypto'
  | 'crosschain'
  | 'evm'
  | 'defi'

/**
 * 可视化运行时规格。
 * 由模块目录解析后得到，用来决定前端应该渲染哪个领域画布以及对应模式。
 */
export interface VisualizationRuntimeSpec {
  moduleKey: string
  simulatorId: string
  renderer: VisualizationRenderer
  algorithm?: 'pbft' | 'raft' | 'pow' | 'pos' | 'dpos' | 'hotstuff' | 'tendermint' | 'dag' | 'vrf' | 'fork_choice'
  protocol?: 'amm' | 'lending' | 'stablecoin' | 'governance' | 'liquidity_pool' | 'concentrated_liquidity' | 'insurance' | 'interest_model' | 'liquidation' | 'options' | 'perpetual' | 've_token' | 'yield_aggregator'
  mode?: 'chain' | 'block' | 'merkle' | 'fork' | 'state' | 'transaction' | 'wallet' | 'mining' | 'mempool' | 'light_client'
  scenario?: 'topology' | 'discovery' | 'gossip' | 'partition' | 'kademlia' | 'block_propagation' | 'eclipse_attack' | 'sybil_attack' | 'bgp_hijack' | 'nat_traversal'
}

export interface VisualizationCanvasState {
  state: SimulatorState
}
