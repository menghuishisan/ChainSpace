import type { ConsensusNode, ConsensusStageSceneProps } from '@/types/visualizationDomain'

type StageKey = 'request' | 'proposal' | 'vote' | 'commit' | 'reply'

type ReplicaNode = {
  id: string
  x: number
  y: number
  faulty: boolean
}

function normalizeText(value?: string, fallback = '--') {
  return value && value.trim().length > 0 ? value : fallback
}

function currentStage(phaseName?: string): StageKey | null {
  const phase = phaseName || 'idle'
  if (phase === 'request') return 'request'
  if (phase === 'pre-prepare' || phase === 'proposal') return 'proposal'
  if (phase === 'prepare' || phase === 'prevote' || phase === 'prepare-qc') return 'vote'
  if (phase === 'commit' || phase === 'precommit' || phase === 'precommit-qc' || phase === 'commit-qc') return 'commit'
  if (phase === 'reply' || phase === 'new-round') return 'reply'
  return null
}

function stageLabel(stage: StageKey) {
  if (stage === 'request') return '请求'
  if (stage === 'proposal') return '提案'
  if (stage === 'vote') return '投票'
  if (stage === 'commit') return '提交'
  return '返回'
}

function buildStageDots(phaseName?: string) {
  const current = currentStage(phaseName)
  const order: StageKey[] = ['request', 'proposal', 'vote', 'commit', 'reply']
  return order.map((key, index) => {
    const activeIndex = current ? order.indexOf(current) : -1
    return {
      key,
      active: current === key,
      done: activeIndex > index,
      label: stageLabel(key),
    }
  })
}

function buildReplicaNodes(nodes: ConsensusNode[]): ReplicaNode[] {
  const replicas = nodes.filter((node) => node.role !== 'leader')
  const count = Math.max(replicas.length, 1)
  const columns = Math.min(count, 3)
  const rows = Math.ceil(count / columns)
  const startX = columns === 1 ? 565 : columns === 2 ? 525 : 470
  const startY = rows > 1 ? 122 : 170
  const gapX = columns === 1 ? 0 : 88
  const gapY = 92

  return replicas.map((node, index) => {
    const column = index % columns
    const row = Math.floor(index / columns)
    return {
      id: node.id,
      x: startX + column * gapX,
      y: startY + row * gapY,
      faulty: node.role === 'byzantine' || node.status === 'faulty',
    }
  })
}

function flowColor(stage: StageKey) {
  if (stage === 'request') return '#38bdf8'
  if (stage === 'proposal') return '#22c55e'
  if (stage === 'vote') return '#f59e0b'
  if (stage === 'commit') return '#a78bfa'
  return '#f472b6'
}

export default function ConsensusBftScene({
  algorithm,
  nodes,
  phase,
  stats,
  timeline,
}: ConsensusStageSceneProps) {
  const leader = nodes.find((node) => node.role === 'leader') || nodes[0]
  const stage = currentStage(phase?.name)
  const stageDots = buildStageDots(phase?.name)
  const replicas = buildReplicaNodes(nodes)
  const faultCount = replicas.filter((replica) => replica.faulty).length
  const voteRatio = phase?.required ? Math.min(((phase?.votes || 0) / phase.required) * 100, 100) : 0
  const latest = timeline[timeline.length - 1]

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex flex-wrap items-center gap-2">
        {stageDots.map((item) => (
          <div
            key={item.key}
            className={`inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs ${
              item.active
                ? 'border-cyan-300 bg-cyan-400/15 text-cyan-100'
                : item.done
                  ? 'border-emerald-300/40 bg-emerald-400/10 text-emerald-100'
                  : 'border-slate-700 bg-slate-900/80 text-slate-400'
            }`}
          >
            <span
              className={`inline-flex h-2.5 w-2.5 rounded-full ${
                item.active ? 'animate-pulse bg-cyan-200' : item.done ? 'bg-emerald-300' : 'bg-slate-600'
              }`}
            />
            {item.label}
          </div>
        ))}
      </div>

      <div className="relative flex-1 overflow-hidden rounded-3xl border border-slate-800 bg-[radial-gradient(circle_at_top,#0f172a_0%,#071120_55%,#020617_100%)]">
        <svg width="100%" height="100%" viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <defs>
            {(['request', 'proposal', 'vote', 'commit', 'reply'] as StageKey[]).map((item) => (
              <marker
                key={`arrow-${item}`}
                id={`arrow-${item}`}
                markerWidth="8"
                markerHeight="8"
                refX="6"
                refY="4"
                orient="auto"
              >
                <path d="M0,0 L0,8 L8,4 z" fill={flowColor(item)} />
              </marker>
            ))}
          </defs>

          <rect x="18" y="18" width="724" height="344" rx="24" fill="rgba(2,6,23,0.44)" stroke="#1e293b" />

          <text x="34" y="42" fill="#f8fafc" fontSize="16" fontWeight="bold">
            {algorithm === 'hotstuff'
              ? 'BFT 证书推进主舞台'
              : algorithm === 'tendermint'
                ? 'BFT 多轮投票主舞台'
                : 'BFT 提案与投票主舞台'}
          </text>
          <text x="34" y="62" fill="#94a3b8" fontSize="11">
            当前阶段：{normalizeText(phase?.name, 'idle')} · 推进者：{normalizeText(stats.leaderId || leader?.id, '--')}
          </text>
          <text x="706" y="62" textAnchor="end" fill="#94a3b8" fontSize="11">
            副本 {replicas.length} · 故障 {faultCount}
          </text>

          <line
            x1="160"
            y1="180"
            x2="270"
            y2="122"
            stroke={flowColor('request')}
            strokeWidth={stage === 'request' ? 4 : 2.5}
            strokeOpacity={stage === 'request' ? 1 : 0.45}
            markerEnd="url(#arrow-request)"
          />
          <line
            x1="372"
            y1="122"
            x2="470"
            y2="122"
            stroke={flowColor('proposal')}
            strokeWidth={stage === 'proposal' ? 4 : 2.5}
            strokeOpacity={stage === 'proposal' ? 1 : 0.45}
            markerEnd="url(#arrow-proposal)"
          />
          <line
            x1="534"
            y1="230"
            x2="382"
            y2="266"
            stroke={flowColor('vote')}
            strokeWidth={stage === 'vote' ? 4 : 2.5}
            strokeOpacity={stage === 'vote' ? 1 : 0.45}
            markerEnd="url(#arrow-vote)"
          />
          <line
            x1="382"
            y1="266"
            x2="588"
            y2="266"
            stroke={flowColor('commit')}
            strokeWidth={stage === 'commit' ? 4 : 2.5}
            strokeOpacity={stage === 'commit' ? 1 : 0.45}
            markerEnd="url(#arrow-commit)"
          />
          <line
            x1="330"
            y1="304"
            x2="162"
            y2="204"
            stroke={flowColor('reply')}
            strokeWidth={stage === 'reply' ? 4 : 2.5}
            strokeOpacity={stage === 'reply' ? 1 : 0.45}
            strokeDasharray="8 6"
            markerEnd="url(#arrow-reply)"
          />

          {stage && (
            <circle r="6" fill={flowColor(stage)}>
              <animateMotion
                dur="1.45s"
                repeatCount="indefinite"
                path={
                  stage === 'request'
                    ? 'M 160 180 L 270 122'
                    : stage === 'proposal'
                      ? 'M 372 122 L 470 122'
                      : stage === 'vote'
                        ? 'M 534 230 L 382 266'
                        : stage === 'commit'
                          ? 'M 382 266 L 588 266'
                          : 'M 330 304 L 162 204'
                }
              />
            </circle>
          )}

          <g transform="translate(120,180)">
            <rect x="-44" y="-34" width="88" height="68" rx="22" fill="#0f172a" stroke="#38bdf8" strokeWidth="2.5" />
            <text x="0" y="-4" textAnchor="middle" fill="#f8fafc" fontSize="15" fontWeight="bold">
              客户端
            </text>
            <text x="0" y="16" textAnchor="middle" fill="#93c5fd" fontSize="10">
              请求 / 返回
            </text>
          </g>

          <g transform="translate(322,122)">
            <rect x="-56" y="-40" width="112" height="80" rx="24" fill="#052e16" stroke="#22c55e" strokeWidth="2.5" />
            <text x="0" y="-10" textAnchor="middle" fill="#f8fafc" fontSize="16" fontWeight="bold">
              {algorithm === 'pbft' ? '主节点' : '提议者'}
            </text>
            <text x="0" y="15" textAnchor="middle" fill="#86efac" fontSize="13">
              {normalizeText(leader?.id || stats.leaderId, '--')}
            </text>
          </g>

          {replicas.map((replica) => (
            <g key={replica.id} transform={`translate(${replica.x},${replica.y})`}>
              <circle
                r="30"
                fill={replica.faulty ? 'rgba(127,29,29,0.85)' : 'rgba(12,74,110,0.85)'}
                stroke={replica.faulty ? '#f87171' : '#38bdf8'}
                strokeWidth="2.5"
              />
              <circle
                r="38"
                fill="none"
                stroke={replica.faulty ? 'rgba(248,113,113,0.22)' : 'rgba(56,189,248,0.2)'}
                strokeWidth="2"
              />
              <text x="0" y="-4" textAnchor="middle" fill="#fff" fontSize="13" fontWeight="bold">
                {replica.id}
              </text>
              <text x="0" y="16" textAnchor="middle" fill={replica.faulty ? '#fecaca' : '#bae6fd'} fontSize="10">
                {replica.faulty ? '异常' : '副本'}
              </text>
            </g>
          ))}

          <g transform="translate(324,304)">
            <rect x="-92" y="-28" width="184" height="56" rx="20" fill="#082f49" stroke="#38bdf8" strokeWidth="2.5" />
            <text x="0" y="-4" textAnchor="middle" fill="#f8fafc" fontSize="14" fontWeight="bold">
              链上结果
            </text>
            <text x="0" y="16" textAnchor="middle" fill="#7dd3fc" fontSize="12">
              已提交 {stats.committedCount ?? 0} 轮
            </text>
          </g>

          <g transform="translate(592,266)">
            <rect x="-90" y="-42" width="180" height="84" rx="22" fill="#1e1b4b" stroke="#a78bfa" strokeWidth="2.5" />
            <text x="0" y="-14" textAnchor="middle" fill="#f8fafc" fontSize="14" fontWeight="bold">
              阈值
            </text>
            <rect x="-58" y="2" width="116" height="10" rx="5" fill="#312e81" />
            <rect x="-58" y="2" width={(116 * voteRatio) / 100} height="10" rx="5" fill="#c4b5fd" />
            <text x="0" y="30" textAnchor="middle" fill="#ddd6fe" fontSize="12">
              {phase?.votes ?? 0} / {phase?.required ?? 0}
            </text>
          </g>

          <g transform="translate(42,282)">
            <rect x="0" y="0" width="156" height="56" rx="18" fill="#0f172a" stroke="#334155" />
            <text x="14" y="20" fill="#f8fafc" fontSize="12" fontWeight="bold">
              最近推进
            </text>
            <text x="14" y="39" fill="#94a3b8" fontSize="10">
              {latest ? `${latest.title} · 步骤 ${latest.tick}` : '等待新的协议事件'}
            </text>
          </g>
        </svg>
      </div>
    </div>
  )
}
