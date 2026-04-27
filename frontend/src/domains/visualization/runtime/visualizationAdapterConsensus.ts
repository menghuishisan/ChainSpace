import type {
  ConsensusMessage,
  ConsensusNode,
  ConsensusPhase,
  ConsensusStats,
  ConsensusTimelineItem,
  SimulatorEvent,
  SimulatorState,
  VisualizationRecord,
  VisualizationRuntimeSpec,
} from '@/types/visualizationDomain'
import {
  asNumber,
  asRecord,
  asString,
  buildDisturbances,
  getEvents,
  getGlobalData,
  getNodes,
} from './visualizationAdapterCommon'
import { getVisualizationEventLabel } from './visualizationEventLabels'
import { getVisualizationEventSummary } from './visualizationEventFormatter'

const consensusEventTypes = new Set([
  'client_request',
  'command_submitted',
  'pre_prepare',
  'prepare',
  'commit',
  'committed',
  'request_vote',
  'vote_response',
  'leader_elected',
  'append_entries',
  'view_change',
  'election_started',
  'block_mined',
  'difficulty_increased',
  'difficulty_decreased',
  '51_attack_started',
  'selfish_mining_enabled',
  'selfish_mining_hidden',
  'block_proposed',
  'block_finalized',
  'epoch_transition',
  'validator_slashed',
  'vote_cast',
  'delegates_elected',
  'block_produced',
  'block_missed',
  'propose',
  'vote_received',
  'prepare_qc_formed',
  'precommit_qc_formed',
  'commit_qc_formed',
  'block_committed',
  'leader_rotated',
  'prevote',
  'precommit',
  'new_round',
  'vertex_created',
  'vertex_confirmed',
  'node_selected',
  'election_complete',
  'block_created',
  'fork_created',
  'chain_reorg',
])

function getConsensusPhaseName(
  algorithm: VisualizationRuntimeSpec['algorithm'],
  latestEvent?: SimulatorEvent,
): string {
  const eventType = latestEvent?.type || 'idle'

  if (algorithm === 'raft') {
    if (eventType === 'leader_elected') return 'leader-elected'
    if (['election_started', 'request_vote', 'vote_response'].includes(eventType)) return 'election'
    if (['append_entries', 'command_submitted'].includes(eventType)) return 'replication'
    return 'idle'
  }

  if (eventType === 'client_request') return 'request'
  if (eventType === 'pre_prepare') return 'pre-prepare'
  if (eventType === 'view_change') return 'view-change'
  if (eventType === 'committed') return 'reply'

  if (algorithm === 'pow') {
    if (eventType === 'block_mined') return 'mining'
    if (['difficulty_increased', 'difficulty_decreased'].includes(eventType)) return 'difficulty-adjust'
    if (['51_attack_started', 'selfish_mining_enabled', 'selfish_mining_hidden'].includes(eventType)) return 'fork-race'
    return 'idle'
  }

  if (algorithm === 'pos') {
    if (eventType === 'block_proposed') return 'proposal'
    if (eventType === 'block_finalized') return 'finalize'
    if (eventType === 'epoch_transition') return 'epoch'
    if (eventType === 'validator_slashed') return 'slashing'
    return 'idle'
  }

  if (algorithm === 'dpos') {
    if (eventType === 'vote_cast') return 'delegate-vote'
    if (eventType === 'delegates_elected') return 'delegate-election'
    if (['block_produced', 'block_missed'].includes(eventType)) return 'block-production'
    return 'idle'
  }

  if (algorithm === 'hotstuff') {
    if (eventType === 'propose') return 'proposal'
    if (eventType === 'prepare_qc_formed') return 'prepare-qc'
    if (eventType === 'precommit_qc_formed') return 'precommit-qc'
    if (eventType === 'commit_qc_formed') return 'commit-qc'
    if (eventType === 'block_committed') return 'commit'
    if (eventType === 'leader_rotated') return 'leader-rotate'
    return 'idle'
  }

  if (algorithm === 'tendermint') {
    if (eventType === 'propose') return 'proposal'
    if (eventType === 'prevote') return 'prevote'
    if (eventType === 'precommit') return 'precommit'
    if (eventType === 'commit') return 'commit'
    if (eventType === 'new_round') return 'new-round'
    return 'idle'
  }

  if (algorithm === 'dag') {
    if (eventType === 'vertex_created') return 'vertex-create'
    if (eventType === 'vertex_confirmed') return 'vertex-confirm'
    return 'idle'
  }

  if (algorithm === 'vrf') {
    if (eventType === 'node_selected') return 'selection'
    if (eventType === 'election_complete') return 'round-complete'
    return 'idle'
  }

  if (algorithm === 'fork_choice') {
    if (eventType === 'block_created') return 'block-proposed'
    if (eventType === 'fork_created') return 'fork-created'
    if (eventType === 'chain_reorg') return 'reorg'
    return 'idle'
  }

  return eventType || 'idle'
}

export function buildConsensusVisualizationData(
  state: SimulatorState,
  runtime: VisualizationRuntimeSpec,
): VisualizationRecord {
  const globalData = getGlobalData(state)
  const events = getEvents(state).filter((event) => consensusEventTypes.has(event.type))
  const rawNodes = getNodes(state)
  const algorithm = runtime.algorithm || 'pbft'
  const currentStep = asString(globalData.step)

  const nodes = rawNodes.map<ConsensusNode>((node, index) => {
    const nodeData = asRecord(node.data)
    const roleValue = asString(nodeData.role, node.status)
    const role: ConsensusNode['role'] =
      node.is_byzantine
        ? 'byzantine'
        : nodeData.is_primary
            || roleValue === 'leader'
            || String(globalData.leader) === node.id
            || String(globalData.proposer) === node.id
            || String(globalData.current_producer) === node.id
            || String(globalData.current_actor) === node.id
          ? 'leader'
          : roleValue === 'candidate'
            ? 'candidate'
            : roleValue === 'follower'
              ? 'follower'
              : 'validator'

    const roleLabel = role === 'leader'
      ? '主导节点'
      : role === 'candidate'
        ? '候选节点'
        : role === 'follower'
          ? '跟随节点'
          : role === 'byzantine'
            ? '异常节点'
            : '参与节点'

    return {
      id: node.id,
      role,
      status: node.is_byzantine ? 'faulty' : (node.status === 'offline' ? 'offline' : 'active'),
      label: roleLabel,
      summary: `${node.id} / 第 ${index + 1} 个参与节点`,
      view: asNumber(nodeData.view),
      term: asNumber(nodeData.term),
      commitIndex: asNumber(nodeData.committed_seq ?? nodeData.commit_index),
      lastLogIndex: asNumber(nodeData.last_log_index ?? nodeData.log_length ?? nodeData.sequence),
      prepareVotes: asNumber(nodeData.prepare_count),
      commitVotes: asNumber(nodeData.commit_count),
      votedFor: asString(nodeData.voted_for),
      roleDescription: role === 'leader' ? '当前负责推动本轮协议前进。' : role === 'byzantine' ? '当前被注入异常行为。' : '参与当前轮次的同步、投票或确认。',
      stateLabel: asString(node.status),
    }
  })

  const latestEvent = events[events.length - 1]
  const phaseName = algorithm === 'tendermint' && currentStep
    ? getConsensusPhaseName(algorithm, { ...latestEvent, type: currentStep } as SimulatorEvent)
    : getConsensusPhaseName(algorithm, latestEvent)

  const threshold = asNumber(
    globalData.threshold ?? globalData.target_count ?? (asNumber(globalData.fault_tolerance) * 2 + 1),
    Math.max(1, Math.floor(nodes.length / 2) + 1),
  )

  const phaseVotes = Math.max(
    0,
    ...rawNodes.map((node) => Math.max(
      asNumber(asRecord(node.data).prepare_count),
      asNumber(asRecord(node.data).commit_count),
      asNumber(asRecord(node.data).votes_granted),
    )),
  )

  const phase: ConsensusPhase = {
    name: phaseName,
    progress: {
      idle: 0,
      request: 16,
      'pre-prepare': 32,
      prepare: 54,
      commit: 78,
      reply: 100,
      election: 36,
      replication: 86,
      proposal: 28,
      finalize: 100,
      'delegate-vote': 26,
      'delegate-election': 52,
      'block-production': 86,
      mining: 42,
      'fork-race': 65,
      'vertex-create': 36,
      'vertex-confirm': 82,
      selection: 34,
      'round-complete': 100,
      'block-proposed': 38,
      reorg: 88,
    }[phaseName] ?? 0,
    votes: phaseVotes,
    required: ['reply', 'finalize', 'round-complete'].includes(phaseName) ? 1 : threshold,
    explanation: latestEvent ? getVisualizationEventLabel(latestEvent.type) : '模拟器已启动，正在等待下一次自动请求或手动操作。',
  }

  const stats: ConsensusStats = {
    requestCount: asNumber(globalData.request_count),
    committedCount: asNumber(globalData.committed_count),
    successCount: events.filter((event) => event.type === 'committed' || event.type === 'leader_elected').length,
    failureCount: events.filter((event) => ['fault_injected', 'node_status_changed'].includes(event.type)).length,
    avgLatency: 0,
    latestEvent: latestEvent?.type || 'idle',
    latestEventLabel: getVisualizationEventLabel(latestEvent?.type || 'idle'),
    faultTolerance: asNumber(globalData.fault_tolerance),
    view: asNumber(globalData.view),
    term: asNumber(globalData.term),
    leaderId: nodes.find((node) => node.role === 'leader')?.id || asString(globalData.current_actor),
    sequence: asNumber(globalData.sequence ?? globalData.result_height ?? globalData.chain_height),
    activeFaultCount: asNumber(globalData.fault_count),
    activeAttackCount: asNumber(globalData.attack_overlay_count),
  }

  const messages = events.slice(-8).map<ConsensusMessage>((event) => ({
    id: event.id || `${event.type}-${event.tick}`,
    from: asString(event.source, 'system'),
    to: asString(event.target, 'network'),
    type: event.type,
    phase: getConsensusPhaseName(algorithm, event),
    status: 'delivered',
    description: getVisualizationEventLabel(event.type),
  }))

  const timeline = events.slice(-6).map<ConsensusTimelineItem>((event) => ({
    id: event.id || `${event.type}-${event.tick}`,
    title: getVisualizationEventLabel(event.type),
    tick: event.tick,
    source: asString(event.source, '-'),
    target: asString(event.target, '-'),
    summary: getVisualizationEventSummary(event),
    phase: getConsensusPhaseName(algorithm, event),
  }))

  return {
    nodes,
    messages,
    phase,
    round: asNumber(globalData.view ?? globalData.term),
    stats,
    timeline,
    disturbances: buildDisturbances(globalData),
  }
}
