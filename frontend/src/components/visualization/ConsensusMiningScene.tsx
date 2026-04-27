import type { ConsensusStageSceneProps } from '@/types/visualizationDomain'

function buildChainBlocks(isFork: boolean) {
  const canonical = [
    { x: 214, y: 128, label: 'B1' },
    { x: 314, y: 128, label: 'B2' },
    { x: 414, y: 128, label: 'B3' },
    { x: 514, y: 128, label: 'Head' },
  ]

  const fork = isFork
    ? [
        { x: 314, y: 248, label: 'F2' },
        { x: 414, y: 248, label: 'F3' },
        { x: 514, y: 248, label: 'Fork' },
      ]
    : []

  return { canonical, fork }
}

function sceneTitle(algorithm?: string) {
  return algorithm === 'pow' ? '出块竞争与主链形成主舞台' : '分叉选择与链重组主舞台'
}

export default function ConsensusMiningScene({
  algorithm,
  phase,
  stats,
  timeline,
}: ConsensusStageSceneProps) {
  const isFork = phase?.name === 'fork-created' || phase?.name === 'fork-race'
  const isReorg = phase?.name === 'reorg'
  const chains = buildChainBlocks(isFork || isReorg)
  const latestEvent = timeline[timeline.length - 1]?.summary || '等待新的出块竞争事件'

  return (
    <svg width="100%" height="100%" viewBox="0 0 760 430" preserveAspectRatio="xMidYMid meet">
      <defs>
        <marker id="mining-arrow" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
          <path d="M0,0 L0,8 L8,4 z" fill="#38bdf8" />
        </marker>
      </defs>

      <rect x="20" y="20" width="720" height="390" rx="28" fill="#020617" stroke="#1e293b" />

      <text x="36" y="46" fill="#f8fafc" fontSize="18" fontWeight="bold">
        {sceneTitle(algorithm)}
      </text>
      <text x="36" y="68" fill="#94a3b8" fontSize="12">
        先看新区块从哪里长出来，再看竞争分支何时形成，以及规范链头何时发生切换。
      </text>

      <g transform="translate(114, 300)">
        <rect x="-68" y="-42" width="136" height="84" rx="24" fill="#0f172a" stroke="#f59e0b" strokeWidth="2.5" />
        <text textAnchor="middle" y="-6" fill="#f8fafc" fontSize="16" fontWeight="bold">
          {algorithm === 'pow' ? '矿工集合' : '候选出块者'}
        </text>
        <text textAnchor="middle" y="18" fill="#fdba74" fontSize="12">
          竞争新区块
        </text>
      </g>

      <line x1="182" y1="288" x2="214" y2="180" stroke="#f59e0b" strokeWidth="4" strokeDasharray="10 5" markerEnd="url(#mining-arrow)" />
      <circle r={6} fill="#f59e0b">
        <animateMotion dur="1.6s" repeatCount="indefinite" path="M 182 288 L 214 180" />
      </circle>

      <text x="182" y="112" fill="#94a3b8" fontSize="12">
        规范链
      </text>
      {chains.canonical.map((block, index) => (
        <g key={block.label} transform={`translate(${block.x}, ${block.y})`}>
          {index > 0 && (
            <line x1="-88" y1="0" x2="-34" y2="0" stroke="#22c55e" strokeWidth="4" strokeLinecap="round" />
          )}
          <rect x="-34" y="-24" width="68" height="48" rx="14" fill="#052e16" stroke="#22c55e" strokeWidth="2.5" />
          <text textAnchor="middle" y="-2" fill="#f8fafc" fontSize="13" fontWeight="bold">
            {block.label}
          </text>
          <text textAnchor="middle" y="16" fill="#86efac" fontSize="10">
            高度 {index + 1}
          </text>
        </g>
      ))}

      {(isFork || isReorg) && (
        <>
          <text x="282" y="232" fill="#c4b5fd" fontSize="12">
            竞争分支
          </text>
          <line x1="314" y1="152" x2="314" y2="224" stroke="#a78bfa" strokeWidth="3" strokeDasharray="8 4" />
          {chains.fork.map((block, index) => (
            <g key={block.label} transform={`translate(${block.x}, ${block.y})`}>
              {index > 0 && (
                <line x1="-88" y1="0" x2="-34" y2="0" stroke="#a78bfa" strokeWidth="4" strokeLinecap="round" />
              )}
              <rect x="-34" y="-24" width="68" height="48" rx="14" fill="#1e1b4b" stroke="#a78bfa" strokeWidth="2.5" />
              <text textAnchor="middle" y="-2" fill="#f8fafc" fontSize="13" fontWeight="bold">
                {block.label}
              </text>
              <text textAnchor="middle" y="16" fill="#ddd6fe" fontSize="10">
                分叉
              </text>
            </g>
          ))}
        </>
      )}

      <g transform="translate(590, 114)">
        <rect x="-88" y="-44" width="176" height="88" rx="26" fill="#082f49" stroke="#38bdf8" strokeWidth="2.5" />
        <text textAnchor="middle" y="-10" fill="#f8fafc" fontSize="16" fontWeight="bold">
          当前结果
        </text>
        <text textAnchor="middle" y="16" fill="#7dd3fc" fontSize="14">
          链高 {stats.committedCount ?? stats.sequence ?? 0}
        </text>
      </g>

      <g transform="translate(590, 266)">
        <rect x="-96" y="-50" width="192" height="100" rx="26" fill="#1e293b" stroke={isReorg ? '#ef4444' : '#a78bfa'} strokeWidth="2.5" />
        <text textAnchor="middle" y="-16" fill="#f8fafc" fontSize="16" fontWeight="bold">
          分叉观察
        </text>
        <text textAnchor="middle" y="10" fill={isReorg ? '#fca5a5' : '#ddd6fe'} fontSize="12">
          {isReorg ? '规范链正在切换到更优分支' : isFork ? '当前存在分叉竞争' : '主链目前保持单一路径'}
        </text>
        <text textAnchor="middle" y="34" fill="#94a3b8" fontSize="10">
          {latestEvent}
        </text>
      </g>
    </svg>
  )
}
