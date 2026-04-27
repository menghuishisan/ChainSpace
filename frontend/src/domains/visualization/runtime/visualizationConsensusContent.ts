import type { ConsensusAlgorithm } from '@/types/visualizationDomain'

type ConsensusContent = {
  title: string
  observations: string[]
  explainers: string[]
}

const DEFAULT_CONTENT: ConsensusContent = {
  title: 'PBFT 三阶段协议可视化',
  observations: [
    '1. 先看主节点何时发出预准备消息，再看 Prepare 与 Commit 票数是否逐步达到阈值。',
    '2. 节点卡片里的 Prepare / Commit 计数，可以帮助判断这一轮为什么能提交，或为什么会被故障打断。',
    '3. 当你注入异常节点或触发视图切换时，重点观察主节点变化、日志位置和时间线是否同步改变。',
  ],
  explainers: [
    'Request：客户端把请求发送给主节点，开始这一轮共识。',
    'Pre-prepare：主节点向所有副本广播同一份提案与摘要。',
    'Prepare：副本确认提案内容一致，并逐步累积票数。',
    'Commit：副本继续累积提交票，达到阈值后允许正式提交。',
    'Reply：节点完成提交后，把结果返回给客户端。',
  ],
}

const CONSENSUS_CONTENT: Partial<Record<ConsensusAlgorithm, ConsensusContent>> = {
  raft: {
    title: 'Raft 选主与日志复制可视化',
    observations: [
      '1. 先看谁发起选举，再看谁获得多数票，最后观察日志是否由新的领导者继续复制。',
      '2. 当领导者失效或离线时，心跳会停止，跟随者会因为超时进入下一轮选举。',
      '3. 节点卡片里的日志位置和提交索引，能够帮助判断复制是否真正达到多数确认。',
    ],
    explainers: [
      'Election：候选节点向其他节点请求投票，只有拿到多数票才能成为领导者。',
      'Leader Elected：新领导者确认后开始发送心跳，并承担日志复制职责。',
      'Replication：领导者向跟随者发送日志条目，达到多数副本后推进提交索引。',
    ],
  },
  pow: {
    title: 'PoW 挖矿与主链竞争可视化',
    observations: [
      '1. 先看哪个矿工率先出块，再看新区块如何推动链高和难度变化。',
      '2. 当出现分叉竞争时，重点观察哪条链最终成为规范链头。',
      '3. 如果叠加 51% 攻击或自私挖矿，要观察主链切换是否被异常影响。',
    ],
    explainers: [
      'Mining：矿工竞争计算满足难度要求的新区块。',
      'Difficulty Adjust：系统根据历史出块速度调节难度，稳定平均出块时间。',
      'Fork Race：当多条竞争链并存时，观察哪条链最终成为主链。',
    ],
  },
  pos: {
    title: 'PoS 提案与最终确认可视化',
    observations: [
      '1. 先看谁是当前提议者，再看区块是否顺利进入最终确认。',
      '2. 观察 Epoch 切换时提议者、已确认高度和活跃验证者数量如何变化。',
      '3. 如果发生 Slash，重点看惩罚后流程是否仍能稳定推进。',
    ],
    explainers: [
      'Proposal：提议者广播新区块，等待验证者确认。',
      'Finalize：验证者完成确认后，区块进入最终确认态。',
      'Epoch：跨越 Epoch 时，会刷新验证者集合与全局状态。',
    ],
  },
  dpos: {
    title: 'DPoS 投票与轮值出块可视化',
    observations: [
      '1. 先看投票如何决定活跃委托者，再看出块顺序是否按轮值推进。',
      '2. 当轮值节点错过出块时，重点观察系统如何切换到下一位出块者。',
      '3. 节点卡片里的日志位置和提交高度，可以帮助判断是否真的完成出块。',
    ],
    explainers: [
      '委托投票：持币人通过投票决定哪些候选者进入活跃委托集合。',
      '选出委托者：系统统计票数，确认当前轮值顺序和活跃委托者名单。',
      '轮值出块：轮到的委托节点负责产生新区块并广播给网络。',
    ],
  },
  hotstuff: {
    title: 'HotStuff QC 链路可视化',
    observations: [
      '1. 先看提案，再看 Prepare QC、Precommit QC、Commit QC 是否依次形成。',
      '2. 每一个 QC 都代表票数达到阈值，是理解 HotStuff 的关键。',
      '3. 领导者轮换时，重点观察谁接管了下一轮视图。',
    ],
    explainers: [
      'Proposal：领导者先发出新区块提案。',
      'Prepare / Precommit / Commit QC：每个 QC 都代表一个阶段票数达标。',
      'Commit：完成最终提交后，系统准备轮换下一位领导者。',
    ],
  },
  tendermint: {
    title: 'Tendermint 多轮投票可视化',
    observations: [
      '1. 先看提案，再看 Prevote 和 Precommit 是否逐步达到阈值。',
      '2. 如果进入新一轮，要判断是正常进入下一高度，还是因为上一轮没达成共识。',
      '3. 时间线能帮助你判断当前是卡在提案、Prevote 还是 Precommit。',
    ],
    explainers: [
      'Proposal：提议者广播区块提案。',
      'Prevote / Precommit：验证者分两轮完成投票。',
      'Commit / New Round：提交后进入下一高度或新一轮。',
    ],
  },
  dag: {
    title: 'DAG 顶点扩展与确认可视化',
    observations: [
      '1. 重点看新顶点如何接入图结构，而不是像传统区块链那样接成单链。',
      '2. 观察哪些顶点已经被足够多的后续引用并进入确认态。',
      '3. Tips 数量的变化可以帮助理解当前 DAG 的并行程度。',
    ],
    explainers: [
      'Vertex Create：节点不断向 DAG 中创建并扩散顶点。',
      'Vertex Confirm：顶点被足够多后续引用后进入确认态。',
    ],
  },
  vrf: {
    title: 'VRF 随机选举可视化',
    observations: [
      '1. 先看 VRF 如何根据随机种子和权重选出本轮候选者。',
      '2. 观察目标数量、已选数量和总质押，理解它不是平均随机，而是带权选择。',
      '3. 每轮结束后，重点看新的随机种子会如何影响下一轮选择。',
    ],
    explainers: [
      'Selection：VRF 根据随机种子和权重选出本轮候选者。',
      'Round Complete：本轮抽签完成后，系统进入下一轮。',
    ],
  },
  fork_choice: {
    title: 'Fork Choice 主链选择可视化',
    observations: [
      '1. 先看新区块如何成为候选链头，再看是否形成分叉。',
      '2. 分叉出现后，重点观察链重组何时发生、规范链头如何切换。',
      '3. 规范链头、分叉数量和重组次数，是理解 fork-choice 的关键指标。',
    ],
    explainers: [
      'Block Proposed：系统产生新的候选区块。',
      'Fork Created：多个链头并存时形成分叉。',
      'Reorg：系统按照规则切换新的规范链头。',
    ],
  },
}

export function getConsensusContent(algorithm?: ConsensusAlgorithm): ConsensusContent {
  if (!algorithm) {
    return DEFAULT_CONTENT
  }

  return CONSENSUS_CONTENT[algorithm] || DEFAULT_CONTENT
}
