import type { ConsensusNode, ConsensusStageSceneProps } from '@/types/visualizationDomain'

function buildCommitteeNodes(nodes: ConsensusNode[]) {
  return nodes.filter((node) => node.role === 'validator' || node.role === 'leader' || node.role === 'candidate')
}

function sceneTitle(algorithm?: string) {
  if (algorithm === 'dpos') {
    return '投票与轮值出块主舞台'
  }
  if (algorithm === 'vrf') {
    return '随机抽签与入选结果主舞台'
  }
  return '委员会选择与推进主舞台'
}

function sceneSummary(algorithm?: string) {
  if (algorithm === 'dpos') {
    return '先看投票与质押如何筛出活跃委托者，再看当前轮值出块者如何把结果推进到链上。'
  }
  if (algorithm === 'vrf') {
    return '先看随机性如何筛选本轮参与者，再看谁被选中推进本轮结果。'
  }
  return '先看候选池如何组成当前委员会，再看本轮关键操作者如何推动结果进入链上。'
}

export default function ConsensusCommitteeScene({
  algorithm,
  nodes,
  phase,
  stats,
  timeline,
}: ConsensusStageSceneProps) {
  const committeeNodes = buildCommitteeNodes(nodes).slice(0, 8)
  const latestOperator = timeline[timeline.length - 1]?.source || stats.leaderId || '--'
  const selectionActive =
    phase?.name === 'selection' || phase?.name === 'delegate-vote' || phase?.name === 'delegate-election'
  const productionActive =
    phase?.name === 'block-production' || phase?.name === 'proposal' || phase?.name === 'finalize'

  return (
    <svg width="100%" height="100%" viewBox="0 0 760 430" preserveAspectRatio="xMidYMid meet">
      <defs>
        <marker id="committee-arrow" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
          <path d="M0,0 L0,8 L8,4 z" fill="#38bdf8" />
        </marker>
      </defs>

      <rect x="20" y="20" width="720" height="390" rx="28" fill="#020617" stroke="#1e293b" />

      <text x="36" y="46" fill="#f8fafc" fontSize="18" fontWeight="bold">
        {sceneTitle(algorithm)}
      </text>
      <text x="36" y="68" fill="#94a3b8" fontSize="12">
        {sceneSummary(algorithm)}
      </text>

      <g transform="translate(126, 198)">
        <rect x="-76" y="-58" width="152" height="116" rx="28" fill="#0f172a" stroke="#38bdf8" strokeWidth="2.5" />
        <text textAnchor="middle" y="-18" fill="#f8fafc" fontSize="16" fontWeight="bold">
          {algorithm === 'vrf' ? '随机种子池' : algorithm === 'dpos' ? '投票 / 质押池' : '候选验证者'}
        </text>
        <text textAnchor="middle" y="16" fill="#93c5fd" fontSize="13">
          {selectionActive ? '正在筛选本轮参与者' : '等待下一轮选择'}
        </text>
        <text textAnchor="middle" y="38" fill="#bae6fd" fontSize="11">
          当前阈值 {phase?.required ?? committeeNodes.length}
        </text>
      </g>

      <g transform="translate(374, 110)">
        <rect x="-96" y="-50" width="192" height="100" rx="26" fill="#10232b" stroke="#34d399" strokeWidth="2.5" />
        <text textAnchor="middle" y="-14" fill="#f8fafc" fontSize="16" fontWeight="bold">
          {algorithm === 'dpos' ? '活跃委托者集合' : algorithm === 'vrf' ? '本轮候选集合' : '当前委员会'}
        </text>
        <text textAnchor="middle" y="14" fill="#86efac" fontSize="14">
          {phase?.votes ?? 0} / {phase?.required ?? 0}
        </text>
        <text textAnchor="middle" y="36" fill="#bbf7d0" fontSize="11">
          当前阶段 {phase?.name || 'idle'}
        </text>
      </g>

      {committeeNodes.map((node, index) => {
        const column = index % 4
        const row = Math.floor(index / 4)
        const x = 270 + column * 92
        const y = 190 + row * 82
        const highlight = node.id === latestOperator || node.id === stats.leaderId

        return (
          <g key={node.id} transform={`translate(${x}, ${y})`}>
            <circle r={24} fill="#0f172a" stroke={highlight ? '#22c55e' : '#64748b'} strokeWidth="2.5" />
            <circle r={32} fill="none" stroke={highlight ? '#22c55e' : '#64748b'} strokeOpacity="0.25" strokeWidth="1.5" />
            <text textAnchor="middle" y="-2" fill="#f8fafc" fontSize="11" fontWeight="bold">
              {node.id.replace(/^(validator|delegate)-/, '')}
            </text>
            <text textAnchor="middle" y="14" fill={highlight ? '#86efac' : '#94a3b8'} fontSize="9">
              {highlight ? '本轮关键' : '候选'}
            </text>
          </g>
        )
      })}

      <g transform="translate(598, 114)">
        <rect x="-90" y="-48" width="180" height="96" rx="26" fill="#1e1b4b" stroke="#a78bfa" strokeWidth="2.5" />
        <text textAnchor="middle" y="-12" fill="#f8fafc" fontSize="16" fontWeight="bold">
          {algorithm === 'dpos' ? '当前出块者' : algorithm === 'vrf' ? '当前入选者' : '当前提议者'}
        </text>
        <text textAnchor="middle" y="14" fill="#c4b5fd" fontSize="14">
          {latestOperator}
        </text>
        <text textAnchor="middle" y="34" fill="#ddd6fe" fontSize="11">
          当前高度 {stats.sequence ?? stats.committedCount ?? 0}
        </text>
      </g>

      <line x1="202" y1="198" x2="278" y2="126" stroke="#f59e0b" strokeWidth="4" strokeDasharray="10 5" markerEnd="url(#committee-arrow)" />
      <line x1="470" y1="110" x2="508" y2="110" stroke="#22c55e" strokeWidth="4" strokeDasharray="10 5" markerEnd="url(#committee-arrow)" />
      <line x1="598" y1="162" x2="598" y2="246" stroke="#38bdf8" strokeWidth="4" strokeDasharray="10 5" markerEnd="url(#committee-arrow)" />

      {selectionActive && (
        <circle r={6} fill="#f59e0b">
          <animateMotion dur="1.6s" repeatCount="indefinite" path="M 202 198 L 278 126" />
        </circle>
      )}
      {productionActive && (
        <>
          <circle r={6} fill="#22c55e">
            <animateMotion dur="1.6s" repeatCount="indefinite" path="M 470 110 L 508 110" />
          </circle>
          <circle r={6} fill="#38bdf8">
            <animateMotion dur="1.6s" repeatCount="indefinite" path="M 598 162 L 598 246" />
          </circle>
        </>
      )}
    </svg>
  )
}
