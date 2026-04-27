import { Tag } from 'antd'
import {
  ClockCircleOutlined,
  ThunderboltOutlined,
  UserOutlined,
} from '@ant-design/icons'
import type {
  ConsensusCanvasProps,
  ConsensusMessage,
  ConsensusNode,
  ConsensusPhase,
  ConsensusStats,
  ConsensusTimelineItem,
} from '@/types/visualizationDomain'
import {
  getConsensusMechanism,
} from '@/domains/visualization/runtime/visualizationConsensus'
import { getConsensusContent } from '@/domains/visualization/runtime/visualizationConsensusContent'
import {
  CONSENSUS_PHASE_LABELS,
  getConsensusProgressLabel,
  getConsensusStageFlow,
  getConsensusStageState,
} from '@/domains/visualization/runtime/visualizationConsensusStage'
import ConsensusBftScene from './ConsensusBftScene'
import ConsensusCommitteeScene from './ConsensusCommitteeScene'
import ConsensusDagScene from './ConsensusDagScene'
import ConsensusLeaderScene from './ConsensusLeaderScene'
import ConsensusMiningScene from './ConsensusMiningScene'

type ConsensusCanvasData = {
  nodes?: ConsensusNode[]
  messages?: ConsensusMessage[]
  phase?: ConsensusPhase
  round?: number
  stats?: ConsensusStats
  timeline?: ConsensusTimelineItem[]
  disturbances?: { id: string; type: string; target: string; label: string; summary: string }[]
}

function renderConsensusMainStage(
  algorithm: ConsensusCanvasProps['algorithm'],
  nodes: ConsensusNode[],
  messages: ConsensusMessage[],
  currentPhase: ConsensusPhase | undefined,
  stats: ConsensusStats,
  timeline: ConsensusTimelineItem[],
) {
  const mechanism = getConsensusMechanism(algorithm)

  switch (mechanism) {
    case 'committee':
      return (
        <ConsensusCommitteeScene
          algorithm={algorithm || 'pos'}
          nodes={nodes}
          messages={messages}
          phase={currentPhase}
          stats={stats}
          timeline={timeline}
        />
      )
    case 'leader_replication':
      return (
        <ConsensusLeaderScene
          algorithm={algorithm || 'raft'}
          nodes={nodes}
          messages={messages}
          phase={currentPhase}
          stats={stats}
          timeline={timeline}
        />
      )
    case 'mining':
      return (
        <ConsensusMiningScene
          algorithm={algorithm || 'pow'}
          nodes={nodes}
          messages={messages}
          phase={currentPhase}
          stats={stats}
          timeline={timeline}
        />
      )
    case 'dag':
      return (
        <ConsensusDagScene
          algorithm={algorithm || 'dag'}
          nodes={nodes}
          messages={messages}
          phase={currentPhase}
          stats={stats}
          timeline={timeline}
        />
      )
    case 'bft':
    default:
      return (
        <ConsensusBftScene
          algorithm={algorithm || 'pbft'}
          nodes={nodes}
          messages={messages}
          phase={currentPhase}
          stats={stats}
          timeline={timeline}
        />
      )
  }
}

function renderIntegrationHint(disturbances: ConsensusCanvasData['disturbances'], phase?: ConsensusPhase) {
  const hasDisturbance = (disturbances?.length ?? 0) > 0

  return (
    <section className="rounded-xl border border-sky-200 bg-gradient-to-r from-sky-50 via-white to-slate-50 p-4">
      <div className="text-sm font-medium text-sky-800">当前联动重点</div>
      <div className="mt-2 text-sm leading-6 text-slate-700">
        {hasDisturbance
          ? `当前流程已经叠加外部影响。优先观察"${CONSENSUS_PHASE_LABELS[phase?.name || 'idle'] || phase?.name || '空闲'}"阶段是否出现票数不足、消息丢失、链头切换或提交被阻断。`
          : '当前流程处于基础协议路径。建议先看阶段轨道，再看主舞台和消息流，建立正常推进的直觉。'}
      </div>
    </section>
  )
}

// 节点状态小卡片组件
function NodeStatusMiniCard({
  node,
}: {
  node: ConsensusNode
}) {
  const statusColor = node.status === 'active'
    ? 'text-emerald-700'
    : node.status === 'offline'
      ? 'text-slate-500'
      : 'text-rose-700'

  const roleColor = {
    leader: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    follower: 'bg-sky-50 text-sky-700 border-sky-200',
    candidate: 'bg-amber-50 text-amber-700 border-amber-200',
    validator: 'bg-violet-50 text-violet-700 border-violet-200',
    byzantine: 'bg-rose-50 text-rose-700 border-rose-200',
  }[node.role] || 'bg-slate-50 text-slate-700 border-slate-200'

  return (
    <div className={`rounded border p-2 text-xs ${
      node.status === 'active'
        ? 'border-emerald-200 bg-emerald-50'
        : 'border-slate-200 bg-slate-50'
    }`}>
      <div className="flex items-center justify-between">
        <span className="font-medium text-slate-900">{node.id}</span>
        <span className={`rounded px-1.5 py-0.5 text-[10px] border ${roleColor}`}>
          {node.label}
        </span>
      </div>
      <div className={`mt-1 ${statusColor}`}>
        {node.status === 'active' ? '● 在线' : '○ 离线'}
      </div>
      <div className="mt-1 text-slate-500 truncate" title={node.summary}>
        {node.summary}
      </div>
    </div>
  )
}

/**
 * 共识类统一外壳。
 * 统一展示阶段轨道、主舞台、联动覆盖层、结果摘要、节点状态和事件时间线。
 *
 * 优化要点：
 * - 清晰的视觉层次：顶部信息栏 → 阶段轨道 → 主舞台 → 辅助面板
 * - 左右分栏布局：左侧主可视化区域，右侧状态面板
 * - 动画和状态指示清晰可见
 */
export default function ConsensusCanvas({ state, algorithm = 'pbft' }: ConsensusCanvasProps) {
  const data = (state.data || {}) as ConsensusCanvasData

  const nodes = data.nodes || []
  const messages = data.messages || []
  const currentPhase = data.phase
  const stats = data.stats || {}
  const timeline = data.timeline || []
  const disturbances = data.disturbances || []
  const progressLabel = getConsensusProgressLabel(algorithm, currentPhase?.name)
  const content = getConsensusContent(algorithm)

  // 计算当前阶段进度百分比
  const phaseProgress = currentPhase?.progress ?? 0

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#eef5ff_0%,#e7eef9_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          {/* 左侧：核心指标 */}
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" className="m-0">算法: {algorithm.toUpperCase()}</Tag>
            <Tag color="cyan" className="m-0">
              视图/轮次: {stats.view ?? data.round ?? 0}
            </Tag>
            {algorithm === 'raft' && (
              <Tag color="purple" className="m-0">任期: {stats.term ?? 0}</Tag>
            )}
            <Tag color={phaseProgress >= 100 ? 'success' : 'processing'} className="m-0">
              阶段: {CONSENSUS_PHASE_LABELS[currentPhase?.name || 'idle'] || currentPhase?.name || '空闲'}
            </Tag>
            {stats.leaderId && (
              <Tag color="lime" className="m-0">推进者: {stats.leaderId}</Tag>
            )}
          </div>

          {/* 右侧：统计信息 */}
          <div className="flex flex-wrap items-center gap-4 text-xs">
            <span className="text-slate-600">
              请求: <span className="text-slate-900">{stats.requestCount ?? 0}</span>
            </span>
            <span className="text-emerald-700">
              成功: {stats.successCount ?? stats.committedCount ?? 0}
            </span>
            <span className="text-amber-700">
              异常: {stats.failureCount ?? 0}
            </span>
            <span className="text-sky-700">
              延迟: {(stats.avgLatency ?? 0).toFixed(1)} 步
            </span>
          </div>
        </div>

        {/* 阶段进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>共识进度</span>
            <span>{phaseProgress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-emerald-500 transition-all duration-500"
              style={{ width: `${phaseProgress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="grid flex-1 grid-cols-1 gap-3 overflow-auto bg-[linear-gradient(180deg,#f7fbff_0%,#f3f6fa_100%)] p-3 xl:grid-cols-[minmax(0,1fr)_288px]">
        <div className="min-w-0 space-y-3">
            {/* 协议阶段轨道 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">协议阶段轨道</div>
              <div className="flex items-center gap-1 overflow-x-auto pb-1">
                {getConsensusStageFlow(algorithm).map((stage, index) => {
                  const { active, completed } = getConsensusStageState(stage.key, currentPhase, algorithm)
                  const isLast = index === getConsensusStageFlow(algorithm).length - 1

                  return (
                    <div key={stage.key} className="flex items-center">
                      <div
                        className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-xs whitespace-nowrap transition-all ${
                          active
                            ? 'border-sky-300 bg-sky-100 text-sky-800 shadow-[0_10px_25px_rgba(56,189,248,0.15)]'
                            : completed
                              ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
                              : 'border-slate-200 bg-white text-slate-500'
                        }`}
                      >
                        {active && (
                          <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-sky-500" />
                        )}
                        {completed && (
                          <span className="text-emerald-600">✓</span>
                        )}
                        <span>{CONSENSUS_PHASE_LABELS[stage.key] || stage.title}</span>
                      </div>
                      {!isLast && (
                        <div className={`mx-1 h-0.5 w-6 ${
                          completed ? 'bg-emerald-300' : 'bg-slate-300'
                        }`} />
                      )}
                    </div>
                  )
                })}
              </div>
            </section>

            {/* 联动提示 */}
            {renderIntegrationHint(disturbances, currentPhase)}

            {/* 主舞台可视化 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 flex items-center justify-between">
                <div className="text-xs font-medium text-slate-900">
                  {content.title}
                  <span className="ml-2 text-slate-500">
                    步骤 {state.tick}
                    {typeof stats.sequence === 'number' ? ` · 序号 ${stats.sequence}` : ''}
                  </span>
                </div>
              </div>

              <div className="relative overflow-hidden rounded-lg border border-slate-200 bg-slate-50"
                   style={{ height: Math.max(280, Math.min(400, nodes.length * 60 + 120)) }}>
                {renderConsensusMainStage(algorithm, nodes, messages, currentPhase, stats, timeline)}
              </div>
            </section>

            {/* 阶段说明和结果 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">当前阶段说明</div>
              <div className="rounded bg-slate-50 p-2 text-xs text-slate-600">
                {currentPhase?.explanation || stats.latestEventLabel || '当前暂无阶段说明。'}
                <div className="mt-1 text-slate-500">{progressLabel}</div>
              </div>

              {/* 结果卡片 */}
              <div className="mt-2 grid grid-cols-4 gap-2">
                <div className="rounded bg-slate-50 p-2 text-center">
                  <div className="text-lg font-semibold text-sky-700">
                    {currentPhase?.votes ?? 0}/{currentPhase?.required ?? nodes.length}
                  </div>
                  <div className="text-xs text-slate-500">当前票数</div>
                </div>
                <div className="rounded bg-slate-50 p-2 text-center">
                  <div className="text-lg font-semibold text-emerald-700">
                    {stats.committedCount ?? 0}
                  </div>
                  <div className="text-xs text-slate-500">已提交</div>
                </div>
                <div className="rounded bg-slate-50 p-2 text-center">
                  <div className="text-lg font-semibold text-amber-700">
                    {stats.faultTolerance ?? 0}
                  </div>
                  <div className="text-xs text-slate-500">容错阈值</div>
                </div>
                <div className="rounded bg-slate-50 p-2 text-center">
                  <div className="text-lg font-semibold text-violet-700">
                    {(stats.avgLatency ?? 0).toFixed(1)}
                  </div>
                  <div className="text-xs text-slate-500">平均延迟</div>
                </div>
              </div>
            </section>

            {/* 观察重点 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">观察重点</div>
              <div className="grid gap-2 md:grid-cols-3">
                {content.observations.map((item, index) => (
                  <div key={`obs-${index}`} className="rounded bg-slate-50 p-2 text-xs text-slate-600">
                    <span className="mr-1 text-sky-600">{index + 1}.</span>
                    {item}
                  </div>
                ))}
              </div>
            </section>
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          {/* 节点状态 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <UserOutlined />
              <span>节点状态</span>
              <span className="ml-auto text-slate-500">({nodes.length})</span>
            </div>
            <div className="space-y-2 max-h-[200px] overflow-auto">
              {nodes.slice(0, 8).map((node) => (
                <NodeStatusMiniCard key={node.id} node={node} />
              ))}
            </div>
          </section>

          {/* 联动影响 */}
          {disturbances.length > 0 && (
            <section className="rounded-lg border border-amber-200 bg-amber-50 p-3">
              <div className="mb-2 text-xs font-medium text-amber-800">联动影响</div>
              <div className="space-y-2">
                {disturbances.map((item) => (
                  <div key={item.id} className="rounded border border-amber-100 bg-white p-2 text-xs">
                    <div className="font-medium text-slate-900">{item.label}</div>
                    <div className="mt-1 text-slate-600">{item.summary}</div>
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* 最近事件 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <ClockCircleOutlined />
              <span>最近事件</span>
              <span className="ml-auto text-slate-500">({timeline.length})</span>
            </div>
            <div className="space-y-2 max-h-[180px] overflow-auto">
              {timeline.length > 0 ? (
                timeline.slice(-5).reverse().map((item, index) => (
                  <div key={`${item.id}-${index}`} className="rounded border border-slate-200 bg-slate-50 p-2 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-slate-900">{item.title}</span>
                      <span className="text-slate-500">步骤 {item.tick}</span>
                    </div>
                    {item.summary && (
                      <div className="mt-1 text-slate-500 line-clamp-2">{item.summary}</div>
                    )}
                  </div>
                ))
              ) : (
                <div className="rounded bg-slate-100 p-2 text-xs text-slate-500">暂无事件</div>
              )}
            </div>
          </section>

          {/* 实验结果摘要 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <ThunderboltOutlined />
              <span>结果摘要</span>
            </div>
            <div className="rounded bg-slate-50 p-2 text-xs">
              <div className="grid grid-cols-2 gap-2">
                <div className="text-center">
                  <div className="text-lg font-semibold text-emerald-700">
                    {stats.successCount ?? 0}
                  </div>
                  <div className="text-slate-500">成功</div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-semibold text-amber-700">
                    {stats.failureCount ?? 0}
                  </div>
                  <div className="text-slate-500">异常</div>
                </div>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}

