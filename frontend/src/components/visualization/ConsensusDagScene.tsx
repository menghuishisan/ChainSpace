import type { ConsensusStageSceneProps } from '@/types/visualizationDomain'

type VertexPoint = {
  id: string
  x: number
  y: number
  confirmed: boolean
}

function buildVertices(confirmedCount: number): VertexPoint[] {
  return [
    { id: 'V1', x: 172, y: 250, confirmed: confirmedCount >= 1 },
    { id: 'V2', x: 286, y: 164, confirmed: confirmedCount >= 2 },
    { id: 'V3', x: 286, y: 316, confirmed: confirmedCount >= 3 },
    { id: 'V4', x: 418, y: 118, confirmed: confirmedCount >= 4 },
    { id: 'V5', x: 418, y: 242, confirmed: confirmedCount >= 5 },
    { id: 'V6', x: 548, y: 188, confirmed: confirmedCount >= 6 },
  ]
}

export default function ConsensusDagScene({
  phase,
  stats,
  timeline,
}: ConsensusStageSceneProps) {
  const createActive = phase?.name === 'vertex-create'
  const confirmActive = phase?.name === 'vertex-confirm'
  const vertices = buildVertices(stats.committedCount ?? 0)
  const latestSummary = timeline[timeline.length - 1]?.summary || '等待新的顶点扩展事件'

  return (
    <svg width="100%" height="100%" viewBox="0 0 760 430" preserveAspectRatio="xMidYMid meet">
      <defs>
        <marker id="dag-arrow" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
          <path d="M0,0 L0,8 L8,4 z" fill="#38bdf8" />
        </marker>
      </defs>

      <rect x="20" y="20" width="720" height="390" rx="28" fill="#020617" stroke="#1e293b" />

      <text x="36" y="46" fill="#f8fafc" fontSize="18" fontWeight="bold">
        DAG 顶点扩展与确认主舞台
      </text>
      <text x="36" y="68" fill="#94a3b8" fontSize="12">
        看新顶点如何并行接入图结构，以及哪些顶点因为后续引用足够多而进入确认状态。
      </text>

      <line x1="172" y1="250" x2="286" y2="164" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />
      <line x1="172" y1="250" x2="286" y2="316" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />
      <line x1="286" y1="164" x2="418" y2="118" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />
      <line x1="286" y1="164" x2="418" y2="242" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />
      <line x1="286" y1="316" x2="418" y2="242" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />
      <line x1="418" y1="118" x2="548" y2="188" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />
      <line x1="418" y1="242" x2="548" y2="188" stroke="#334155" strokeWidth="3.5" markerEnd="url(#dag-arrow)" />

      {vertices.map((vertex) => (
        <g key={vertex.id} transform={`translate(${vertex.x}, ${vertex.y})`}>
          <circle
            r={createActive && !vertex.confirmed ? 24 : 20}
            fill={vertex.confirmed ? '#22c55e' : '#0f172a'}
            stroke={vertex.confirmed ? '#86efac' : '#38bdf8'}
            strokeWidth="2.5"
          />
          <circle r={28} fill="none" stroke={vertex.confirmed ? '#22c55e' : '#38bdf8'} strokeOpacity="0.22" strokeWidth="1.5" />
          <text textAnchor="middle" y="-2" fill="#f8fafc" fontSize="12" fontWeight="bold">
            {vertex.id}
          </text>
          <text textAnchor="middle" y="16" fill={vertex.confirmed ? '#bbf7d0' : '#93c5fd'} fontSize="10">
            {vertex.confirmed ? '已确认' : '等待确认'}
          </text>
        </g>
      ))}

      {createActive && (
        <circle r={6} fill="#38bdf8">
          <animateMotion dur="1.5s" repeatCount="indefinite" path="M 286 164 L 418 242" />
        </circle>
      )}
      {confirmActive && (
        <circle r={7} fill="#22c55e">
          <animateMotion dur="1.5s" repeatCount="indefinite" path="M 418 242 L 548 188" />
        </circle>
      )}

      <g transform="translate(128, 104)">
        <rect x="-80" y="-36" width="160" height="72" rx="22" fill="#0f172a" stroke="#38bdf8" strokeWidth="2" />
        <text textAnchor="middle" y="-4" fill="#f8fafc" fontSize="15" fontWeight="bold">
          图结构持续扩展
        </text>
        <text textAnchor="middle" y="18" fill="#93c5fd" fontSize="12">
          并行生成，而不是单链排队
        </text>
      </g>

      <g transform="translate(590, 104)">
        <rect x="-88" y="-36" width="176" height="72" rx="22" fill="#10232b" stroke="#34d399" strokeWidth="2" />
        <text textAnchor="middle" y="-4" fill="#f8fafc" fontSize="15" fontWeight="bold">
          确认结果
        </text>
        <text textAnchor="middle" y="18" fill="#86efac" fontSize="12">
          已确认 {stats.committedCount ?? 0}
        </text>
      </g>

      <g transform="translate(386, 336)">
        <rect x="-224" y="-26" width="448" height="52" rx="18" fill="#0f172a" stroke="#334155" />
        <text x="-202" y="4" fill="#cbd5e1" fontSize="11">
          {latestSummary}
        </text>
      </g>
    </svg>
  )
}
