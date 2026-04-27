import { useMemo } from 'react'
import { Tag } from 'antd'
import {
  DollarOutlined,
  LineChartOutlined,
  PercentageOutlined,
  SwapOutlined,
} from '@ant-design/icons'
import type {
  AMMCurveSnapshot,
  AMMPoolSnapshot,
  DeFiCanvasProps,
  DeFiStats,
  SwapRecord,
  VisualizationDisturbanceItem,
} from '@/types/visualizationDomain'

function formatValue(value: number, digits = 2): string {
  return value.toLocaleString('zh-CN', { maximumFractionDigits: digits })
}

function buildCurvePath(curve?: AMMCurveSnapshot): string {
  const points = curve?.points || []
  if (points.length === 0) return ''

  const xs = points.map((p) => p.x)
  const ys = points.map((p) => p.y)
  const minX = Math.min(...xs)
  const maxX = Math.max(...xs)
  const minY = Math.min(...ys)
  const maxY = Math.max(...ys)
  const scaleX = (v: number) => 48 + ((v - minX) / Math.max(maxX - minX, 1)) * 340
  const scaleY = (v: number) => 270 - ((v - minY) / Math.max(maxY - minY, 1)) * 210

  return points
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${scaleX(p.x)} ${scaleY(p.y)}`)
    .join(' ')
}

function buildCurrentDot(curve?: AMMCurveSnapshot): { x: number; y: number } | null {
  const points = curve?.points || []
  const current = curve?.current
  if (!current || points.length === 0) return null

  const xs = points.map((p) => p.x)
  const ys = points.map((p) => p.y)
  const minX = Math.min(...xs)
  const maxX = Math.max(...xs)
  const minY = Math.min(...ys)
  const maxY = Math.max(...ys)

  return {
    x: 48 + ((current.x - minX) / Math.max(maxX - minX, 1)) * 340,
    y: 270 - ((current.y - minY) / Math.max(maxY - minY, 1)) * 210,
  }
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

/**
 * DeFi可视化组件
 * 展示AMM自动做市商等DeFi机制
 * 
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - AMM曲线清晰展示
 * - 交易历史和状态同步
 */
export default function DeFiCanvas({ state, protocol = 'amm' }: DeFiCanvasProps) {
  const data = (state.data || {}) as {
    pool?: AMMPoolSnapshot
    curve?: AMMCurveSnapshot
    swaps?: SwapRecord[]
    stats?: DeFiStats
    disturbances?: VisualizationDisturbanceItem[]
  }
  const globalData = (state.global_data || {}) as Record<string, unknown>

  const pool = data.pool
  const swaps = data.swaps || []
  const stats = data.stats || {}
  const latestSwap = swaps[0]
  const disturbances = data.disturbances || buildDisturbances(globalData)
  const curvePath = useMemo(() => buildCurvePath(data.curve), [data.curve])
  const currentDot = useMemo(() => buildCurrentDot(data.curve), [data.curve])
  const flowProgress = latestSwap ? Math.min(100, 35 + (latestSwap.priceImpact || 0) * 8) : 18

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#eefbf6_0%,#eef6ff_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" icon={<SwapOutlined />} className="m-0">
              协议: {protocol.toUpperCase()}
            </Tag>
            <Tag color="cyan" icon={<DollarOutlined />} className="m-0">
              TVL: {formatValue(stats.tvl ?? 0)}
            </Tag>
            <Tag color="green" icon={<PercentageOutlined />} className="m-0">
              即时价格: {formatValue(stats.spotPrice ?? 0, 4)}
            </Tag>
            <Tag color="orange" className="m-0">
              事件数: {stats.eventCount ?? 0}
            </Tag>
          </div>
          <div className="text-xs text-slate-500">
            步骤: {state.tick}
          </div>
        </div>

        {/* 过程进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>交易进度</span>
            <span>{flowProgress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-emerald-500 transition-all duration-500"
              style={{ width: `${flowProgress}%` }}
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

            {/* 关键指标 */}
            <section className="grid gap-3 lg:grid-cols-3">
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs text-slate-500">资产A储备</div>
                <div className="mt-1 text-lg font-semibold text-sky-700">
                  {formatValue(pool?.reserveA ?? 0, 4)}
                </div>
                <div className="mt-1 text-xs text-slate-500">
                  兑换会推动池子重新分配
                </div>
              </div>
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs text-slate-500">资产B储备</div>
                <div className="mt-1 text-lg font-semibold text-orange-700">
                  {formatValue(pool?.reserveB ?? 0, 4)}
                </div>
                <div className="mt-1 text-xs text-slate-500">
                  买入时另一侧储备上升
                </div>
              </div>
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs text-slate-500">常数乘积k</div>
                <div className="mt-1 text-lg font-semibold text-emerald-700">
                  {formatValue(pool?.constantProduct ?? 0, 2)}
                </div>
                <div className="mt-1 text-xs text-slate-500">
                  x * y = k 恒定
                </div>
              </div>
            </section>

            {/* 过程轨道 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">AMM过程轨道</div>
              <div className="flex items-center gap-1 overflow-x-auto pb-1">
                {['输入进入池子', '储备重新配平', '价格沿曲线偏移', '输出与结果反馈'].map((item, index) => {
                  const active = flowProgress >= (index + 1) * 25
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

            {/* AMM曲线主可视化 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-3 text-xs font-medium text-slate-900">
                AMM曲线与当前池子位置
              </div>
              <div className="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
                {/* 曲线图 */}
                <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                  {curvePath ? (
                    <svg width="100%" height="280" viewBox="0 0 420 300" preserveAspectRatio="xMidYMid meet">
                      <defs>
                        <linearGradient id="amm-curve-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                          <stop offset="0%" stopColor="#38bdf8" />
                          <stop offset="100%" stopColor="#f97316" />
                        </linearGradient>
                      </defs>
                      {/* 坐标轴 */}
                      <line x1="40" y1="250" x2="400" y2="250" stroke="#475569" strokeWidth="2" />
                      <line x1="40" y1="250" x2="40" y2="30" stroke="#475569" strokeWidth="2" />
                      {/* 轴标签 */}
                      <text x="220" y="285" textAnchor="middle" fill="#94a3b8" fontSize="11">
                        资产A储备
                      </text>
                      <text x="15" y="140" textAnchor="middle" fill="#94a3b8" fontSize="11" transform="rotate(-90 15 140)">
                        资产B储备
                      </text>
                      {/* 曲线 */}
                      <path d={curvePath} fill="none" stroke="url(#amm-curve-gradient)" strokeWidth="3" />
                      {/* 当前点 */}
                      {currentDot && (
                        <>
                          <line 
                            x1="40" y1={currentDot.y} x2={currentDot.x} y2={currentDot.y}
                            stroke="#64748b" strokeWidth="1" strokeDasharray="4"
                          />
                          <line 
                            x1={currentDot.x} y1={currentDot.y} x2={currentDot.x} y2="250"
                            stroke="#64748b" strokeWidth="1" strokeDasharray="4"
                          />
                          <circle cx={currentDot.x} cy={currentDot.y} r="8" fill="#22c55e" stroke="#fff" strokeWidth="2" />
                          <circle cx={currentDot.x} cy={currentDot.y} r="14" fill="rgba(34,197,94,0.15)">
                            <animate attributeName="r" values="14;20;14" dur="2s" repeatCount="indefinite" />
                            <animate attributeName="opacity" values="0.3;0.05;0.3" dur="2s" repeatCount="indefinite" />
                          </circle>
                          <text x={currentDot.x + 12} y={currentDot.y - 8} fill="#0f172a" fontSize="10">
                            当前
                          </text>
                        </>
                      )}
                    </svg>
                  ) : (
                    <div className="flex h-[280px] items-center justify-center text-xs text-slate-500">
                      等待池子初始化...
                    </div>
                  )}
                </div>

                {/* 交易推演 */}
                <div className="space-y-3">
                  <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                    <div className="text-xs font-medium text-slate-900 mb-2">交易推演</div>
                    <div className="space-y-2 text-xs text-slate-600">
                      <div>1. 每次兑换沿曲线移动</div>
                      <div>2. 交易量越大，滑点越高</div>
                      <div>3. 流动性影响价格敏感度</div>
                    </div>
                    {latestSwap && (
                      <div className="mt-3 rounded bg-white p-2 text-xs">
                        <div className="font-medium text-slate-900 mb-1">最近操作</div>
                        <div className="text-slate-600">{latestSwap.title}</div>
                        <div className="text-slate-500">
                          {latestSwap.tokenIn} → {latestSwap.tokenOut}
                        </div>
                        <div className="mt-1">
                          <span className="text-sky-700">{formatValue(latestSwap.amountIn, 4)}</span>
                          <span className="text-slate-500 mx-1">→</span>
                          <span className="text-emerald-700">{formatValue(latestSwap.amountOut, 4)}</span>
                        </div>
                        <div className="mt-1 text-amber-700">
                          价格影响: {formatValue(latestSwap.priceImpact, 4)}%
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
                    <div className="text-xs font-medium text-slate-900 mb-2">池子摘要</div>
                    <div className="space-y-1 text-xs">
                      <div className="flex justify-between">
                        <span className="text-slate-500">交易对</span>
                        <span>{pool?.pair || '--'}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-slate-500">滑点</span>
                        <span>{formatValue(stats.slippage ?? 0, 4)}%</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-slate-500">价格</span>
                        <span className="text-sky-700">{formatValue(stats.spotPrice ?? 0, 4)}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </section>
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <LineChartOutlined />
              <span>最近交易</span>
              <span className="ml-auto text-slate-500">({swaps.length})</span>
            </div>
            <div className="space-y-2 max-h-[400px] overflow-auto">
              {swaps.length > 0 ? (
                swaps.slice(0, 12).map((swap, index) => (
                  <div key={swap.id || index} className={`rounded border p-2 text-xs ${
                    index === 0 
                      ? 'border-sky-200 bg-sky-50'
                      : 'border-slate-200 bg-slate-50'
                  }`}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-slate-900">{swap.title}</span>
                      <span className="text-amber-700">
                        -{formatValue(swap.priceImpact, 2)}%
                      </span>
                    </div>
                    <div className="text-slate-500">
                      {swap.tokenIn} → {swap.tokenOut}
                    </div>
                    <div className="mt-1 text-slate-600">
                      {formatValue(swap.amountIn, 4)} → {formatValue(swap.amountOut, 4)}
                    </div>
                  </div>
                ))
              ) : (
                <div className="rounded bg-slate-100 p-3 text-xs text-slate-500">
                  暂无交易记录
                </div>
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}

