import type { SimulatorState } from './simulation'

export interface VisualizationMetricCard {
  key: string
  label: string
  value: string
  hint?: string
}

export interface VisualizationEntityDetail {
  label: string
  value: string
}

export interface VisualizationEntityCard {
  id: string
  title: string
  subtitle?: string
  status?: string
  details: VisualizationEntityDetail[]
}

export interface VisualizationStateItem {
  label: string
  value: string
}

export interface VisualizationStateSection {
  key: string
  title: string
  items: VisualizationStateItem[]
}

export interface VisualizationEventItem {
  id: string
  title: string
  summary: string
  tick: number
  source?: string
  target?: string
}

export interface VisualizationDisturbanceItem {
  id: string
  type: string
  target: string
  label: string
  summary: string
}

/**
 * 共识类可视化相关类型定义。
 */
export interface ConsensusNode {
  id: string
  role: 'leader' | 'follower' | 'candidate' | 'validator' | 'byzantine'
  status: 'active' | 'faulty' | 'offline'
  label: string
  summary: string
  view?: number
  term?: number
  commitIndex?: number
  lastLogIndex?: number
  prepareVotes?: number
  commitVotes?: number
  votedFor?: string
  roleDescription?: string
  stateLabel?: string
}

export interface ConsensusMessage {
  id: string
  from: string
  to: string
  type: string
  phase?: string
  status: 'pending' | 'delivered' | 'dropped'
  description: string
}

export interface ConsensusPhase {
  name: string
  progress: number
  votes?: number
  required?: number
  explanation?: string
}

export interface ConsensusStats {
  requestCount?: number
  committedCount?: number
  successCount?: number
  failureCount?: number
  avgLatency?: number
  latestEvent?: string
  latestEventLabel?: string
  faultTolerance?: number
  view?: number
  term?: number
  leaderId?: string
  sequence?: number
  activeFaultCount?: number
  activeAttackCount?: number
}

export interface ConsensusTimelineItem {
  id: string
  title: string
  tick: number
  source: string
  target: string
  summary?: string
  phase?: string
}

export interface ConsensusPhaseGuide {
  meaning: string
  condition: string
  next: string
}

export interface ConsensusStageFlowItem {
  key: string
  title: string
  actor: string
  goal: string
}

/**
 * 共识主舞台输入。
 * 当节点数量过多时，画布会对相近角色进行聚合展示，避免画面过度拥挤。
 */
export interface ConsensusCanvasProps {
  state: SimulatorState
  algorithm?: 'pbft' | 'raft' | 'pow' | 'pos' | 'dpos' | 'hotstuff' | 'tendermint' | 'dag' | 'vrf' | 'fork_choice'
}

export type ConsensusAlgorithm =
  | 'pbft'
  | 'raft'
  | 'pow'
  | 'pos'
  | 'dpos'
  | 'hotstuff'
  | 'tendermint'
  | 'dag'
  | 'vrf'
  | 'fork_choice'

export type ConsensusMechanism =
  | 'bft'
  | 'leader_replication'
  | 'committee'
  | 'mining'
  | 'dag'

export interface ConsensusStageSceneProps {
  algorithm: ConsensusAlgorithm
  nodes: ConsensusNode[]
  messages: ConsensusMessage[]
  phase?: ConsensusPhase
  stats: ConsensusStats
  timeline: ConsensusTimelineItem[]
}

/**
 * 区块链基础类可视化相关类型定义。
 */
export interface ChainBlock {
  id: string
  number: number
  hash: string
  prevHash: string
  txCount: number
  status: 'current' | 'confirmed'
  explanation: string
}

export interface BlockField {
  key: string
  name: string
  value: string
}

export interface BlockTransaction {
  id: string
  hash: string
  from: string
  to: string
  value: string
  status: string
}

export interface CurrentBlockSummary {
  hash: string
  prevHash: string
  merkleRoot: string
  timestamp: string
  nonce: string
  difficulty: string
  txCount: number
}

export interface BlockchainStats {
  height?: number
  txCount?: number
  fieldCount?: number
  difficulty?: string
  leafCount?: number
  treeHeight?: number
  proofDepth?: number
}

export interface MerkleTreeNodeView {
  hash: string
  index: number
  isLeaf: boolean
}

export interface MerkleTreeLevelView {
  level: number
  nodes: MerkleTreeNodeView[]
}

export interface MerkleProofStepView {
  direction: string
  sibling: string
}

export interface MerkleProofView {
  leafIndex: number
  leafHash: string
  root: string
  verified: boolean
  steps: MerkleProofStepView[]
}

export interface BlockchainCanvasProps {
  state: SimulatorState
  mode?: 'chain' | 'block' | 'merkle' | 'fork' | 'state' | 'transaction' | 'wallet' | 'mining' | 'mempool' | 'light_client'
}

/**
 * 攻击类可视化相关类型定义。
 */
export interface BalanceCard {
  key: string
  label: string
  value: number
  color: string
  description: string
}

export interface AttackTimelineItem {
  id: string
  title: string
  description: string
  amount: string
  depth: number
}

export interface CallFrame {
  id: string
  title: string
  description: string
  caller: string
  callee: string
  amount: string
  depth: number
}

export interface StorageSlot {
  key: string
  label: string
  value: string
}

export interface AttackSummary {
  title: string
  summary: string
  depth: number
  completedSteps: number
}

export interface AttackStats {
  steps?: number
  attackDepth?: number
  drainedAmount?: number
  remainingBalance?: number
}

export type AttackMechanism =
  | 'execution'
  | 'economic'
  | 'consensus'
  | 'bridge'

export interface AttackStageSceneProps {
  moduleKey: string
  metrics: VisualizationMetricCard[]
  actors: VisualizationEntityCard[]
  sections: VisualizationStateSection[]
  timeline: AttackTimelineItem[]
  callFrames: CallFrame[]
  storage: StorageSlot[]
  balances: BalanceCard[]
  stats: AttackStats
  attack?: AttackSummary
  events: VisualizationEventItem[]
}

export interface AttackCanvasData {
  attack?: AttackSummary
  balances?: BalanceCard[]
  timeline?: AttackTimelineItem[]
  callFrames?: CallFrame[]
  storage?: StorageSlot[]
  stats?: AttackStats
  metrics?: VisualizationMetricCard[]
  actors?: VisualizationEntityCard[]
  sections?: VisualizationStateSection[]
  events?: VisualizationEventItem[]
  observationTips?: string[]
}

export interface AttackCanvasProps {
  state: SimulatorState
  moduleKey?: string
}

/**
 * EVM 执行类可视化相关类型定义。
 */
export interface EVMFrame {
  id: string
  title: string
  opcode?: string
  description: string
  depth: number
  gas?: string
}

export interface EVMExecutionStats {
  frameCount?: number
  opcodeCount?: number
  storageChanges?: number
}

export interface EVMCanvasProps {
  state: SimulatorState
  moduleKey?: string
}

/**
 * DeFi 可视化相关类型定义。
 */
export interface CurvePoint {
  x: number
  y: number
}

export interface SwapRecord {
  id: string
  type: string
  title: string
  tokenIn: string
  tokenOut: string
  amountIn: number
  amountOut: number
  priceImpact: number
}

export interface AMMPoolSnapshot {
  pair: string
  reserveA: number
  reserveB: number
  price: number
  constantProduct: number
  feeRate: number
  tvl: number
}

export interface AMMCurveSnapshot {
  points: CurvePoint[]
  current: { x: number; y: number }
}

export interface DeFiStats {
  tvl?: number
  spotPrice?: number
  slippage?: number
  eventCount?: number
}

export interface DeFiCanvasProps {
  state: SimulatorState
  protocol?: 'amm' | 'lending' | 'stablecoin' | 'governance' | 'liquidity_pool' | 'concentrated_liquidity' | 'insurance' | 'interest_model' | 'liquidation' | 'options' | 'perpetual' | 've_token' | 'yield_aggregator'
}

/**
 * 网络类可视化相关类型定义。
 */
export interface NetworkNode {
  id: string
  label: string
  type: string
  status: string
  peers: string[]
  latency: number
  partition: string
}

export interface NetworkMessage {
  id: string
  from: string
  to: string
  type: string
  progress: number
  description: string
}

export interface NetworkStats {
  totalNodes?: number
  onlineNodes?: number
  avgLatency?: number
  sessionCount?: number
  edgeCount?: number
  connectivity?: string
}

export interface NetworkCanvasProps {
  state: SimulatorState
  scenario?: 'topology' | 'discovery' | 'gossip' | 'partition' | 'kademlia' | 'block_propagation' | 'eclipse_attack' | 'sybil_attack' | 'bgp_hijack' | 'nat_traversal'
}

/**
 * 密码学与跨链类画布输入定义。
 */
export interface CryptoCanvasProps {
  state: SimulatorState
  moduleKey?: string
}

export interface CrossChainCanvasProps {
  state: SimulatorState
  moduleKey?: string
}
