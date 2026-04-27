import type {
  ConsensusCanvasProps,
  ConsensusPhase,
  ConsensusPhaseGuide,
  ConsensusStageFlowItem,
} from '@/types/visualizationDomain'

export const CONSENSUS_PHASE_LABELS: Record<string, string> = {
  idle: '空闲',
  running: '运行中',
  request: '请求',
  'pre-prepare': '预准备',
  prepare: '准备',
  commit: '提交',
  reply: '响应客户端',
  election: '选主投票',
  replication: '日志复制',
  'leader-elected': '确认领导者',
  'view-change': '视图切换',
  mining: '挖矿出块',
  'difficulty-adjust': '难度调整',
  'fork-race': '分叉竞争',
  proposal: '发起提案',
  finalize: '最终确认',
  epoch: 'Epoch 切换',
  slashing: '惩罚处理',
  'delegate-vote': '委托投票',
  'delegate-election': '选出委托者',
  'block-production': '轮值出块',
  'prepare-qc': 'Prepare QC',
  'precommit-qc': 'Precommit QC',
  'commit-qc': 'Commit QC',
  'leader-rotate': '轮换领导者',
  prevote: 'Prevote',
  precommit: 'Precommit',
  'new-round': '新一轮',
  'vertex-create': '创建顶点',
  'vertex-confirm': '确认顶点',
  selection: 'VRF 抽签',
  'round-complete': '本轮完成',
  'block-proposed': '提出新区块',
  'fork-created': '形成分叉',
  reorg: '链重组',
}

const CONSENSUS_PHASE_GUIDES: Record<string, ConsensusPhaseGuide> = {
  idle: {
    meaning: '系统处于待命状态，还没有进入当前这一轮协议主流程。',
    condition: '等待自动请求，或由学生手动发起一次操作。',
    next: '关键参与者收到请求后，会进入对应算法的第一阶段。',
  },
  request: {
    meaning: '客户端已经把请求提交给主节点，准备开始本轮共识。',
    condition: '主节点成功生成请求摘要并准备广播。',
    next: '主节点广播预准备消息，进入预准备阶段。',
  },
  'pre-prepare': {
    meaning: '主节点正在把本轮提案和摘要广播给所有副本。',
    condition: '副本节点收到相同提案并准备确认。',
    next: '副本逐步投出 Prepare 票，进入准备阶段。',
  },
  prepare: {
    meaning: '副本节点正在确认提案一致性，并逐步累积投票。',
    condition: 'Prepare 票达到法定阈值。',
    next: '节点开始广播 Commit，进入提交阶段。',
  },
  commit: {
    meaning: '节点正在确认这一请求可以正式提交。',
    condition: 'Commit 票达到法定阈值。',
    next: '系统向客户端返回结果，进入响应客户端阶段。',
  },
  reply: {
    meaning: '本轮请求已经完成提交，系统正在把结果返回给客户端。',
    condition: '客户端收到响应，本轮流程结束。',
    next: '如果继续自动运行，会开始下一轮请求。',
  },
  election: {
    meaning: '节点正在发起选主投票，尝试选出新的领导者。',
    condition: '候选者获得多数票。',
    next: '确认领导者后，进入日志复制阶段。',
  },
  replication: {
    meaning: '领导者正在复制日志并推进提交索引。',
    condition: '多数节点追上当前日志。',
    next: '日志提交后，继续处理下一轮命令。',
  },
  'leader-elected': {
    meaning: '新的领导者已经被确认，集群恢复稳定。',
    condition: '大多数节点承认当前领导者。',
    next: '随后会进入日志复制或新的客户端请求处理。',
  },
  'view-change': {
    meaning: '系统正在切换视图，尝试让新的主节点接管。',
    condition: '足够多节点同意视图切换。',
    next: '新主节点接管后，重新进入协议主流程。',
  },
  mining: {
    meaning: '矿工正在持续尝试找到满足难度要求的新块。',
    condition: '率先找到有效区块头并成功广播。',
    next: '新区块会被全网接受，并继续竞争规范链头。',
  },
  'difficulty-adjust': {
    meaning: '系统正在根据历史出块速度重新计算全网难度。',
    condition: '新的难度参数被确定并同步到网络。',
    next: '后续新区块会在新的难度下继续竞争产生。',
  },
  'fork-race': {
    meaning: '当前存在多条竞争链，系统正在比较哪条链更优。',
    condition: '其中一条链取得更高权重或更长累计工作量。',
    next: '主链确定后，网络回到单链继续出块。',
  },
  proposal: {
    meaning: '当前提议者或领导者正在向全网广播提案。',
    condition: '其他节点收到提案并准备进入投票或确认。',
    next: '系统会进入该算法对应的投票阶段。',
  },
  finalize: {
    meaning: '系统正在把当前区块推进到最终确认状态。',
    condition: '满足最终确认规则，状态不可逆。',
    next: '协议会进入下一槽位或下一轮提案。',
  },
  epoch: {
    meaning: '系统正在跨越 Epoch 边界并刷新全局状态。',
    condition: '新一轮 Epoch 的验证者与参数更新完成。',
    next: '新的提议者集合会继续参与后续流程。',
  },
  slashing: {
    meaning: '协议正在对违规验证者应用惩罚。',
    condition: '违规行为被确认并写入惩罚结果。',
    next: '其余节点继续维持正常共识。',
  },
  'delegate-vote': {
    meaning: '持币人正在把票投给目标委托节点。',
    condition: '票数统计完成，可以确认活跃委托集合。',
    next: '系统会选出委托者，并进入轮值出块。',
  },
  'delegate-election': {
    meaning: '系统正在根据票数确认当前轮值委托节点。',
    condition: '活跃委托者集合和顺序完全确定。',
    next: '委托节点开始轮流出块。',
  },
  'block-production': {
    meaning: '当前轮到的节点正在负责产生新区块。',
    condition: '区块成功生成并被网络接受。',
    next: '系统会切换到下一位轮值节点继续出块。',
  },
  'prepare-qc': {
    meaning: '系统正在形成 Prepare QC，证明第一阶段票数达标。',
    condition: '领导者收齐阈值投票并形成 QC。',
    next: '协议继续进入 Precommit QC 阶段。',
  },
  'precommit-qc': {
    meaning: '系统正在形成第二阶段法定证书。',
    condition: 'Precommit 票达到阈值。',
    next: '协议继续进入 Commit QC 阶段。',
  },
  'commit-qc': {
    meaning: '系统正在形成最终提交证书。',
    condition: 'Commit 票达到阈值。',
    next: '区块正式提交并准备轮换领导者。',
  },
  'leader-rotate': {
    meaning: '系统正在把下一轮职责切换给新的领导者。',
    condition: '新领导者成功接管当前视图。',
    next: '新的领导者会发出下一轮提案。',
  },
  prevote: {
    meaning: '验证者正在进行 Tendermint 的第一轮投票。',
    condition: 'Prevote 达到法定阈值。',
    next: '协议进入 Precommit 阶段。',
  },
  precommit: {
    meaning: '验证者正在进行第二轮提交前投票。',
    condition: 'Precommit 达到法定阈值。',
    next: '区块进入 Commit 并写入当前高度。',
  },
  'new-round': {
    meaning: '系统已经结束上一轮，正在开启新一轮共识。',
    condition: '新的提议者准备就绪。',
    next: '重新回到提案阶段开始下一轮。',
  },
  'vertex-create': {
    meaning: '节点正在向 DAG 中创建新顶点，并连接前序引用。',
    condition: '顶点成功扩散并被其他节点看到。',
    next: '随着更多引用出现，该顶点会进入确认态。',
  },
  'vertex-confirm': {
    meaning: '当前顶点已经被足够多的后续引用确认。',
    condition: '确认规则阈值被满足。',
    next: '新的顶点会继续扩展这张图。',
  },
  selection: {
    meaning: 'VRF 正在根据随机种子和权重抽选本轮候选者。',
    condition: '所有候选结果生成完毕。',
    next: '系统宣布本轮结果并刷新下一轮种子。',
  },
  'round-complete': {
    meaning: '本轮随机选举已经完成，系统准备进入下一轮。',
    condition: '本轮结果被确认并写入状态。',
    next: '新的随机种子会驱动下一轮抽签。',
  },
  'block-proposed': {
    meaning: '网络刚刚产生新的候选区块，等待链选择规则判断它的去向。',
    condition: '新区块进入多个候选链头之一。',
    next: '如果链头竞争加剧，就会形成分叉。',
  },
  'fork-created': {
    meaning: '当前已经出现多条竞争链头，系统进入分叉观察期。',
    condition: '某一条分支逐渐取得更高优先级。',
    next: '系统会根据 fork-choice 规则进行链重组。',
  },
  reorg: {
    meaning: '系统正在把规范链头切换到更优分支。',
    condition: '新的规范链头被确认并被网络接受。',
    next: '网络会在新的主链基础上继续出块。',
  },
}

const BFT_STAGE_FLOW: ConsensusStageFlowItem[] = [
  { key: 'request', title: '客户端请求', actor: 'Client -> Primary', goal: '把本轮请求交给主节点，开始生成提案。' },
  { key: 'pre-prepare', title: '主节点提案', actor: 'Primary -> Replicas', goal: '广播同一份摘要与序号，让所有副本进入同一轮。' },
  { key: 'prepare', title: '副本确认', actor: 'Replicas <-> Replicas', goal: '逐步累积 Prepare 票，确认提案内容一致。' },
  { key: 'commit', title: '提交确认', actor: 'Replicas <-> Replicas', goal: '继续累积 Commit 票，满足法定阈值后允许提交。' },
  { key: 'reply', title: '返回结果', actor: 'Replica -> Client', goal: '完成提交并把本轮结果返回给客户端。' },
]

const LEADER_STAGE_FLOW: ConsensusStageFlowItem[] = [
  { key: 'election', title: '发起选举', actor: 'Candidate -> Followers', goal: '候选节点向其他节点请求投票，争取多数支持。' },
  { key: 'leader-elected', title: '确认领导者', actor: 'Majority -> Leader', goal: '多数节点承认新的领导者，集群恢复稳定。' },
  { key: 'replication', title: '复制日志', actor: 'Leader -> Followers', goal: '复制日志并推进提交索引，直到多数节点追上。' },
]

const MINING_STAGE_FLOW_BY_ALGORITHM: Record<'pow' | 'fork_choice', ConsensusStageFlowItem[]> = {
  pow: [
    { key: 'mining', title: '挖矿出块', actor: 'Miner -> Network', goal: '矿工持续尝试随机数，直到挖出新区块并广播。' },
    { key: 'difficulty-adjust', title: '难度调整', actor: 'Protocol', goal: '根据历史出块速度调整难度，稳定平均区块时间。' },
    { key: 'fork-race', title: '分叉竞争', actor: 'Miners <-> Chains', goal: '当出现隐藏块或双链头时，观察哪条链最终胜出。' },
  ],
  fork_choice: [
    { key: 'block-proposed', title: '提出新区块', actor: 'Miner -> Tips', goal: '系统产生新的候选区块并挂到链头附近。' },
    { key: 'fork-created', title: '形成分叉', actor: 'Tips -> Tips', goal: '多个候选区块同时存在，分叉开始显现。' },
    { key: 'reorg', title: '选择主链', actor: 'Fork Choice -> Chain', goal: '按照规则选出新的规范链头并完成重组。' },
  ],
}

const COMMITTEE_STAGE_FLOW_BY_ALGORITHM: Record<'pos' | 'dpos' | 'hotstuff' | 'tendermint' | 'vrf', ConsensusStageFlowItem[]> = {
  pos: [
    { key: 'proposal', title: '提出新区块', actor: 'Proposer -> Validators', goal: '当前槽位的提议者向验证者广播候选区块。' },
    { key: 'finalize', title: '最终确认', actor: 'Validators -> Chain', goal: '完成投票与确认，推动区块进入 finalized 状态。' },
    { key: 'epoch', title: 'Epoch 切换', actor: 'Protocol', goal: '在 Epoch 边界更新验证者和全局状态。' },
  ],
  dpos: [
    { key: 'delegate-vote', title: '委托投票', actor: 'Voters -> Delegates', goal: '持币人把权重投给目标委托节点。' },
    { key: 'delegate-election', title: '选出委托者', actor: 'Protocol', goal: '统计票数并确认当前活跃委托节点集合。' },
    { key: 'block-production', title: '轮值出块', actor: 'Delegate -> Network', goal: '轮到的委托节点负责产生新区块。' },
  ],
  hotstuff: [
    { key: 'proposal', title: '发起提案', actor: 'Leader -> Replicas', goal: '领导者广播新区块提案。' },
    { key: 'prepare-qc', title: '形成 Prepare QC', actor: 'Replicas -> Leader', goal: '收集足够投票，形成第一份法定证书。' },
    { key: 'precommit-qc', title: '形成 Precommit QC', actor: 'Replicas -> Leader', goal: '继续收集第二阶段票，确认链路稳定。' },
    { key: 'commit-qc', title: '形成 Commit QC', actor: 'Replicas -> Leader', goal: '收齐最后一轮证书并准备正式提交。' },
  ],
  tendermint: [
    { key: 'proposal', title: '发起提案', actor: 'Proposer -> Validators', goal: '提议者广播候选区块。' },
    { key: 'prevote', title: 'Prevote', actor: 'Validators <-> Validators', goal: '第一轮投票确认提案是否成立。' },
    { key: 'precommit', title: 'Precommit', actor: 'Validators <-> Validators', goal: '第二轮投票确认是否可以提交。' },
    { key: 'new-round', title: '新一轮', actor: 'Protocol', goal: '若当前轮未完成，就推进到下一轮提案。' },
  ],
  vrf: [
    { key: 'selection', title: 'VRF 抽签', actor: 'Seed -> Candidates', goal: '根据随机种子和权重选出本轮候选者。' },
    { key: 'round-complete', title: '本轮完成', actor: 'Protocol', goal: '确认抽签结果并刷新下一轮随机性。' },
  ],
}

const DAG_STAGE_FLOW: ConsensusStageFlowItem[] = [
  { key: 'vertex-create', title: '创建顶点', actor: 'Nodes -> DAG', goal: '节点并行创建新顶点并连接前序引用。' },
  { key: 'vertex-confirm', title: '确认顶点', actor: 'References -> Vertex', goal: '随着更多后续引用出现，顶点逐步进入确认态。' },
]

export function getConsensusPhaseGuide(phaseName?: string, _algorithm?: ConsensusCanvasProps['algorithm']) {
  if (!phaseName) {
    return CONSENSUS_PHASE_GUIDES.idle
  }

  return CONSENSUS_PHASE_GUIDES[phaseName] || CONSENSUS_PHASE_GUIDES.idle
}

export function getConsensusStageFlow(algorithm?: ConsensusCanvasProps['algorithm']): ConsensusStageFlowItem[] {
  switch (algorithm) {
    case 'raft':
      return LEADER_STAGE_FLOW
    case 'pow':
      return MINING_STAGE_FLOW_BY_ALGORITHM.pow
    case 'fork_choice':
      return MINING_STAGE_FLOW_BY_ALGORITHM.fork_choice
    case 'pos':
      return COMMITTEE_STAGE_FLOW_BY_ALGORITHM.pos
    case 'dpos':
      return COMMITTEE_STAGE_FLOW_BY_ALGORITHM.dpos
    case 'hotstuff':
      return COMMITTEE_STAGE_FLOW_BY_ALGORITHM.hotstuff
    case 'tendermint':
      return COMMITTEE_STAGE_FLOW_BY_ALGORITHM.tendermint
    case 'vrf':
      return COMMITTEE_STAGE_FLOW_BY_ALGORITHM.vrf
    case 'dag':
      return DAG_STAGE_FLOW
    case 'pbft':
    default:
      return BFT_STAGE_FLOW
  }
}

export function getConsensusStageState(
  stageKey: string,
  currentPhase?: ConsensusPhase,
  algorithm?: ConsensusCanvasProps['algorithm'],
) {
  const flow = getConsensusStageFlow(algorithm)
  const currentIndex = flow.findIndex((stage) => stage.key === (currentPhase?.name || 'idle'))
  const targetIndex = flow.findIndex((stage) => stage.key === stageKey)

  return {
    active: targetIndex === currentIndex,
    completed: currentIndex > -1 && targetIndex > -1 && targetIndex < currentIndex,
  }
}

export function getConsensusProgressLabel(
  algorithm?: ConsensusCanvasProps['algorithm'],
  phaseName?: string,
) {
  const label = CONSENSUS_PHASE_LABELS[phaseName || 'idle'] || phaseName || '空闲'
  return `${algorithm?.toUpperCase() || 'CONSENSUS'} · ${label}`
}
