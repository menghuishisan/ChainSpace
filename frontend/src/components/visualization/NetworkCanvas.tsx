import { useMemo } from 'react'
import { Tag } from 'antd'
import {
  GlobalOutlined,
  NodeIndexOutlined,
  WifiOutlined,
} from '@ant-design/icons'
import type {
  NetworkCanvasProps,
  NetworkMessage,
  NetworkNode,
  NetworkStats,
  VisualizationDisturbanceItem,
} from '@/types/visualizationDomain'

const nodeTone: Record<string, string> = {
  bootstrap: '#8b5cf6',
  full: '#38bdf8',
  light: '#22c55e',
  miner: '#f59e0b',
  offline: '#64748b',
}

function buildNodePositions(nodes: NetworkNode[]): Record<string, { x: number; y: number; ring: number }> {
  const centerX = 330
  const centerY = 210

  if (nodes.length <= 8) {
    const radius = 142
    return nodes.reduce<Record<string, { x: number; y: number; ring: number }>>((result, node, index) => {
      const angle = (Math.PI * 2 * index) / Math.max(nodes.length, 1) - Math.PI / 2
      result[node.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
        ring: 0,
      }
      return result
    }, {})
  }

  const ringConfig =
    nodes.length <= 14
      ? [4, nodes.length - 4]
      : [5, Math.min(8, nodes.length - 5), Math.max(nodes.length - 13, 0)]
  const radii = [74, 150, 220]
  const result: Record<string, { x: number; y: number; ring: number }> = {}
  let cursor = 0

  ringConfig.forEach((count, ringIndex) => {
    if (count <= 0) {
      return
    }

    for (let index = 0; index < count; index += 1) {
      const node = nodes[cursor]
      if (!node) {
        break
      }

      const angle = (Math.PI * 2 * index) / count - Math.PI / 2
      result[node.id] = {
        x: centerX + radii[ringIndex] * Math.cos(angle),
        y: centerY + radii[ringIndex] * Math.sin(angle),
        ring: ringIndex,
      }
      cursor += 1
    }
  })

  return result
}

function buildDisturbances(globalData: Record<string, unknown>): VisualizationDisturbanceItem[] {
  const activeFaults = Array.isArray(globalData.active_faults)
    ? (globalData.active_faults as Record<string, unknown>[])
    : []
  const activeAttacks = Array.isArray(globalData.active_attacks)
    ? (globalData.active_attacks as Record<string, unknown>[])
    : []

  return [
    ...activeFaults.map((item, index) => ({
      id: String(item.id ?? `fault-${index}`),
      type: String(item.type ?? 'fault'),
      target: String(item.target ?? '--'),
      label: '故障联动',
      summary: `当前存在故障注入：${String(item.type ?? '--')}，作用目标为 ${String(item.target ?? '--')}。`,
    })),
    ...activeAttacks.map((item, index) => ({
      id: String(item.id ?? `attack-${index}`),
      type: String(item.type ?? 'attack'),
      target: String(item.target ?? '--'),
      label: '攻击联动',
      summary: `当前存在攻击叠加：${String(item.type ?? '--')}，作用目标为 ${String(item.target ?? '--')}。`,
    })),
  ]
}

function renderNetworkMessages(
  messages: NetworkMessage[],
  positions: Record<string, { x: number; y: number; ring: number }>,
) {
  return messages.slice(-8).map((message, index) => {
    const from = positions[message.from]
    const to = positions[message.to]
    if (!from || !to) {
      return null
    }

    const progress = Math.min(0.92, Math.max(0.12, ((message.progress || 100) / 100 + index * 0.11) % 1))
    const x = from.x + (to.x - from.x) * progress
    const y = from.y + (to.y - from.y) * progress

    return (
      <g key={message.id}>
        <circle cx={x} cy={y} r="9" fill="#22c55e" opacity="0.95" />
        <circle cx={x} cy={y} r="16" fill="rgba(34,197,94,0.12)">
          <animate attributeName="r" values="12;18;12" dur="1.8s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.4;0.05;0.4" dur="1.8s" repeatCount="indefinite" />
        </circle>
        <text x={x} y={y + 4} textAnchor="middle" fill="#fff" fontSize="10" fontWeight="bold">
          {message.type.slice(0, 1).toUpperCase()}
        </text>
      </g>
    )
  })
}

/**
 * P2P网络可视化组件
 * 展示网络拓扑、节点发现、消息传播等
 * 
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - 节点和消息清晰展示
 */
export default function NetworkCanvas({ state, scenario = 'topology' }: NetworkCanvasProps) {
  const data = (state.data || {}) as {
    nodes?: NetworkNode[]
    messages?: NetworkMessage[]
    stats?: NetworkStats
  }
  const globalData = (state.global_data || {}) as Record<string, unknown>

  const nodes = data.nodes || []
  const messages = data.messages || []
  const stats = data.stats || {}
  const disturbances = buildDisturbances(globalData)
  const positions = useMemo(() => buildNodePositions(nodes), [nodes])

  // 计算消息传播进度
  const msgProgress = Math.min(100, messages.length * 12)

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#eefbf6_0%,#eef6ff_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" icon={<GlobalOutlined />} className="m-0">
              在线: {stats.onlineNodes ?? 0}/{stats.totalNodes ?? 0}
            </Tag>
            <Tag color="cyan" icon={<NodeIndexOutlined />} className="m-0">
              {scenario === 'discovery' ? '会话' : '连边'}: {scenario === 'discovery' ? (stats.sessionCount ?? 0) : (stats.edgeCount ?? 0)}
            </Tag>
            <Tag color="green" icon={<WifiOutlined />} className="m-0">
              延迟: {(stats.avgLatency ?? 0).toFixed(0)}ms
            </Tag>
            {disturbances.length > 0 && (
              <Tag color="orange" className="m-0">
                联动: {disturbances.length}
              </Tag>
            )}
          </div>
          <div className="text-xs text-slate-500">
            步骤: {state.tick}
          </div>
        </div>

        {/* 过程进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>消息传播</span>
            <span>{msgProgress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-emerald-500 to-cyan-500 transition-all duration-500"
              style={{ width: `${msgProgress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="grid flex-1 grid-cols-1 gap-3 overflow-auto bg-[linear-gradient(180deg,#f6fffb_0%,#f7fbff_100%)] p-3 xl:grid-cols-[minmax(0,1fr)_288px]">
        <div className="min-w-0 space-y-3">
            {/* 联动影响提示 */}
            {disturbances.length > 0 && (
              <section className="rounded-lg border border-amber-200 bg-amber-50 p-3">
                <div className="mb-2 text-xs font-medium text-amber-800">
                  当前存在联动影响
                </div>
                <div className="grid gap-2 md:grid-cols-2">
                  {disturbances.map((item) => (
                    <div key={item.id} className="rounded border border-amber-100 bg-white p-2 text-xs">
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-slate-900">{item.label}</span>
                        <span className="text-amber-700">{item.type}</span>
                      </div>
                      <div className="mt-1 text-slate-600">{item.summary}</div>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {/* 过程轨道 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">
                {scenario === 'discovery' ? '节点发现轨道' : '消息传播轨道'}
              </div>
              <div className="flex items-center gap-1 overflow-x-auto pb-1">
                {(['形成拓扑', '建立邻居', '扩散消息', '反馈结果']).map((item, index) => {
                  const active = msgProgress >= (index + 1) * 25
                  return (
                    <div key={item} className="flex items-center">
                      <div
                        className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-xs whitespace-nowrap transition-all ${
                          active
                            ? 'border-sky-300 bg-sky-100 text-sky-800 shadow-[0_10px_25px_rgba(56,189,248,0.15)]'
                            : 'border-slate-200 bg-white text-slate-500'
                        }`}
                      >
                        {active && (
                          <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-sky-500" />
                        )}
                        <span>{item}</span>
                      </div>
                      {index < 3 && (
                        <div className={`mx-1 h-0.5 w-6 ${
                          active ? 'bg-sky-300' : 'bg-slate-300'
                        }`} />
                      )}
                    </div>
                  )
                })}
              </div>
            </section>

            {/* 网络拓扑主可视化 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-3 flex items-center justify-between">
                <div className="text-xs font-medium text-slate-900">
                  {scenario === 'discovery' ? '节点发现与邻居扩散' : 'P2P拓扑与消息传播'}
                </div>
                <div className="rounded bg-sky-50 px-2 py-1 text-xs text-sky-700">
                  消息: {messages.length}
                </div>
              </div>
              
              <div className="rounded-lg border border-slate-200 bg-slate-50 overflow-hidden">
                {nodes.length > 0 ? (
                  <svg width="100%" height="340" viewBox="0 0 660 340" preserveAspectRatio="xMidYMid meet">
                    {/* 绘制连边 */}
                    {nodes.flatMap((node) =>
                      node.peers.map((peer) => {
                        const from = positions[node.id]
                        const to = positions[peer]
                        if (!from || !to || node.id > peer) {
                          return []
                        }
                        return [
                          <line
                            key={`${node.id}-${peer}`}
                            x1={from.x}
                            y1={from.y}
                            x2={to.x}
                            y2={to.y}
                            stroke="#334155"
                            strokeWidth={from.ring === to.ring ? 1.5 : 2}
                          />,
                        ]
                      }),
                    )}

                    {/* 绘制消息 */}
                    {renderNetworkMessages(messages, positions)}

                    {/* 绘制节点 */}
                    {nodes.map((node) => {
                      const position = positions[node.id]
                      const tone =
                        node.status === 'offline'
                          ? nodeTone.offline
                          : (nodeTone[node.type] || '#38bdf8')
                      const radius = nodes.length > 14 ? 18 : 24

                      return (
                        <g key={node.id} transform={`translate(${position.x}, ${position.y})`}>
                          {/* 外圈光晕 */}
                          <circle 
                            r={radius + 8} 
                            fill="transparent" 
                            stroke={`${tone}30`} 
                            strokeWidth="2"
                            opacity={node.status === 'offline' ? 0.3 : 0.8}
                          />
                          {/* 节点主体 */}
                          <circle
                            r={radius}
                            fill={tone}
                            stroke="#fff"
                            strokeWidth="2"
                            opacity={node.status === 'offline' ? 0.5 : 1}
                          />
                          {/* 节点标签 */}
                          <text textAnchor="middle" dy="4" fill="#fff" fontSize={nodes.length > 14 ? '9' : '11'} fontWeight="bold">
                            {node.id.replace('node-', 'N')}
                          </text>
                          {/* 节点名 */}
                          <text textAnchor="middle" dy={radius + 16} fill="#cbd5e1" fontSize="10">
                            {node.label}
                          </text>
                          {/* 在线状态指示 */}
                          {node.status !== 'offline' && (
                            <circle 
                              cx={radius * 0.7} 
                              cy={-radius * 0.7} 
                              r="4" 
                              fill="#22c55e"
                              stroke="#fff"
                              strokeWidth="1"
                            />
                          )}
                        </g>
                      )
                    })}
                  </svg>
                ) : (
                  <div className="flex h-[340px] items-center justify-center text-xs text-slate-500">
                    等待节点数据初始化...
                  </div>
                )}
              </div>
            </section>

            {/* 消息详情 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-900">
                当前消息流
              </div>
              <div className="space-y-2 max-h-[150px] overflow-auto">
                {messages.length > 0 ? (
                  messages.slice(-5).reverse().map((message) => (
                    <div key={message.id} className="rounded border border-slate-200 bg-slate-50 p-2 text-xs">
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-slate-900">{message.type}</span>
                        <span className="text-slate-500">
                          {message.from} → {message.to}
                        </span>
                      </div>
                      <div className="mt-1 h-1 overflow-hidden rounded-full bg-slate-200">
                        <div
                          className="h-full rounded-full bg-gradient-to-r from-emerald-400 to-cyan-400"
                          style={{ width: `${Math.max(10, message.progress ?? 100)}%` }}
                        />
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="rounded bg-slate-50 p-2 text-xs text-slate-500">
                    暂无消息传播
                  </div>
                )}
              </div>
            </section>

            {/* 观察提示 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-900">观察提示</div>
              <div className="grid gap-2 md:grid-cols-2">
                <div className="rounded bg-slate-50 p-2 text-xs text-slate-600">
                  <div className="mb-1 text-sky-600">节点角色</div>
                  <div>Bootstrap: 入口节点</div>
                  <div>Full Node: 保存更多邻居</div>
                  <div>Offline: 暂时离线</div>
                </div>
                <div className="rounded bg-slate-50 p-2 text-xs text-slate-600">
                  <div className="mb-1 text-sky-600">连通状态</div>
                  <div>在线节点: {stats.onlineNodes ?? 0}</div>
                  <div>连边数: {stats.edgeCount ?? 0}</div>
                  <div>延迟: {(stats.avgLatency ?? 0).toFixed(0)}ms</div>
                </div>
              </div>
            </section>
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <NodeIndexOutlined />
              <span>节点状态</span>
              <span className="ml-auto text-slate-500">({nodes.length})</span>
            </div>
            <div className="space-y-2 max-h-[500px] overflow-auto">
              {nodes.length > 0 ? (
                nodes.map((node) => (
                  <div key={node.id} className={`rounded border p-2 text-xs ${
                    node.status === 'offline'
                      ? 'border-slate-200 bg-slate-50 opacity-60'
                      : 'border-emerald-200 bg-emerald-50'
                  }`}>
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-slate-900">{node.label}</span>
                      <span className={node.status === 'offline' ? 'text-slate-500' : 'text-emerald-700'}>
                        {node.status === 'offline' ? '离线' : '在线'}
                      </span>
                    </div>
                    <div className="mt-1 text-slate-500">
                      角色: {node.type}
                    </div>
                    <div className="mt-1 text-slate-500">
                      邻居: {node.peers.length}
                    </div>
                  </div>
                ))
              ) : (
                <div className="rounded bg-slate-100 p-2 text-xs text-slate-500">
                  暂无节点数据
                </div>
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}

