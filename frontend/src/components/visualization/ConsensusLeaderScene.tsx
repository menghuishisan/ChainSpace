import type { ConsensusNode, ConsensusStageSceneProps } from '@/types/visualizationDomain'

function buildFollowerNodes(nodes: ConsensusNode[]) {
  return nodes.filter((node) => node.role !== 'leader' && node.role !== 'candidate')
}

function buildFollowerProgress(node: ConsensusNode, latestLogHeight: number) {
  return Math.min(((node.commitIndex ?? 0) / Math.max(latestLogHeight || 1, 1)) * 100, 100)
}

export default function ConsensusLeaderScene({
  nodes,
  phase,
  stats,
  timeline,
}: ConsensusStageSceneProps) {
  const leader = nodes.find((node) => node.role === 'leader')
  const candidate = nodes.find((node) => node.role === 'candidate') || leader || nodes[0]
  const followers = buildFollowerNodes(nodes).slice(0, 5)
  const activeElection = phase?.name === 'election' || phase?.name === 'leader-elected'
  const activeReplication = phase?.name === 'replication'
  const recentLogHeight = stats.committedCount ?? stats.sequence ?? 0
  const latestEvent = timeline[timeline.length - 1]?.summary || '当前还没有新的选举或复制事件。'

  return (
    <svg width="100%" height="100%" viewBox="0 0 760 430" preserveAspectRatio="xMidYMid meet">
      <defs>
        <marker id="leader-arrow" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
          <path d="M0,0 L0,8 L8,4 z" fill="#38bdf8" />
        </marker>
      </defs>

      <rect x="20" y="20" width="720" height="390" rx="28" fill="#020617" stroke="#1e293b" />

      <text x="36" y="46" fill="#f8fafc" fontSize="18" fontWeight="bold">
        Leader 选举与日志复制主舞台
      </text>
      <text x="36" y="68" fill="#94a3b8" fontSize="12">
        先看谁赢得选举，再看日志如何从 Leader 复制到多数节点，并最终推进提交索引。
      </text>

      <g transform="translate(126, 196)">
        <rect x="-66" y="-50" width="132" height="100" rx="26" fill="#111827" stroke="#f59e0b" strokeWidth="2.5" />
        <text textAnchor="middle" y="-12" fill="#f8fafc" fontSize="16" fontWeight="bold">
          候选者
        </text>
        <text textAnchor="middle" y="12" fill="#fdba74" fontSize="14">
          {candidate?.id || '--'}
        </text>
        <text textAnchor="middle" y="34" fill="#fde68a" fontSize="11">
          争取多数投票
        </text>
      </g>

      <g transform="translate(352, 116)">
        <rect x="-84" y="-52" width="168" height="104" rx="26" fill="#052e16" stroke="#22c55e" strokeWidth="2.5" />
        <text textAnchor="middle" y="-14" fill="#f8fafc" fontSize="16" fontWeight="bold">
          当前 Leader
        </text>
        <text textAnchor="middle" y="12" fill="#86efac" fontSize="14">
          {leader?.id || stats.leaderId || '--'}
        </text>
        <text textAnchor="middle" y="36" fill="#bbf7d0" fontSize="11">
          负责心跳广播与日志复制
        </text>
      </g>

      <g transform="translate(352, 300)">
        <rect x="-100" y="-48" width="200" height="96" rx="26" fill="#082f49" stroke="#38bdf8" strokeWidth="2.5" />
        <text textAnchor="middle" y="-12" fill="#f8fafc" fontSize="16" fontWeight="bold">
          提交进度
        </text>
        <text textAnchor="middle" y="14" fill="#7dd3fc" fontSize="14">
          已提交索引 {recentLogHeight}
        </text>
        <text textAnchor="middle" y="36" fill="#bae6fd" fontSize="11">
          当前阶段 {phase?.name || 'idle'}
        </text>
      </g>

      <line
        x1="194"
        y1="196"
        x2="278"
        y2="136"
        stroke="#f59e0b"
        strokeWidth={activeElection ? 4.5 : 2.5}
        strokeOpacity={activeElection ? 1 : 0.35}
        strokeDasharray="10 5"
        markerEnd="url(#leader-arrow)"
      />
      {activeElection && (
        <circle r={6} fill="#f59e0b">
          <animateMotion dur="1.5s" repeatCount="indefinite" path="M 194 196 L 278 136" />
        </circle>
      )}

      {followers.map((node, index) => {
        const x = 586
        const y = 90 + index * 62
        const progress = buildFollowerProgress(node, recentLogHeight)

        return (
          <g key={node.id} transform={`translate(${x}, ${y})`}>
            <rect x="-84" y="-28" width="168" height="56" rx="20" fill="#1e293b" stroke="#64748b" strokeWidth="2" />
            <text textAnchor="middle" y="-6" fill="#f8fafc" fontSize="13" fontWeight="bold">
              {node.id}
            </text>
            <rect x="-54" y="8" width="108" height="8" rx="4" fill="#334155" />
            <rect x="-54" y="8" width={(108 * progress) / 100} height="8" rx="4" fill="#22c55e" />
            <text textAnchor="middle" y="28" fill="#94a3b8" fontSize="10">
              复制进度 {node.commitIndex ?? 0}
            </text>

            <line
              x1="-162"
              y1={20 - y}
              x2="-88"
              y2={0}
              stroke="#38bdf8"
              strokeWidth={activeReplication ? 4 : 2.5}
              strokeOpacity={activeReplication ? 1 : 0.35}
              strokeDasharray="10 5"
              markerEnd="url(#leader-arrow)"
            />

            {activeReplication && (
              <circle r={5} fill="#38bdf8">
                <animateMotion
                  dur="1.5s"
                  repeatCount="indefinite"
                  path={`M ${x - 162} ${y + 20} L ${x - 88} ${y}`}
                />
              </circle>
            )}
          </g>
        )
      })}

      <g transform="translate(36, 338)">
        <rect x="0" y="0" width="300" height="54" rx="18" fill="#0f172a" stroke="#334155" />
        <text x="16" y="21" fill="#f8fafc" fontSize="13" fontWeight="bold">
          最近协议推进
        </text>
        <text x="16" y="40" fill="#cbd5e1" fontSize="11">
          {latestEvent}
        </text>
      </g>
    </svg>
  )
}
