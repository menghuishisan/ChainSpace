import { Tag } from 'antd'
import { GatewayOutlined } from '@ant-design/icons'
import type {
  CrossChainCanvasProps,
  VisualizationDisturbanceItem,
  VisualizationEntityCard,
  VisualizationEventItem,
  VisualizationMetricCard,
  VisualizationStateSection,
} from '@/types/visualizationDomain'
import { getVisualizationModuleLabel } from '@/domains/visualization/runtime/visualizationMeta'

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

function renderLifecycle(events: VisualizationEventItem[]) {
  const stages = ['源链发起', '中继证明', '桥验证', '目标执行']
  const progress = Math.min(stages.length, Math.max(0, events.length))

  return (
    <section className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold text-slate-900">跨链生命周期</div>
        <div className="text-xs text-slate-500">已推进 {progress} / {stages.length} 阶段</div>
      </div>
      <div className="grid gap-3 md:grid-cols-4">
        {stages.map((stage, index) => {
          const active = progress >= index + 1
          return (
            <div
              key={stage}
              className={`rounded-xl border p-3 text-sm ${
                active
                  ? 'border-sky-200 bg-sky-50 text-sky-800'
                  : 'border-slate-200 bg-white text-slate-600'
              }`}
            >
              <div className="flex items-center justify-between gap-3">
                <span className="font-medium">{stage}</span>
                <span className="text-xs">{index + 1}</span>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

function renderEntityCard(entity: VisualizationEntityCard) {
  return (
    <div key={entity.id} className="rounded-2xl border border-slate-200 bg-white p-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-slate-900">{entity.title}</div>
          {entity.subtitle && <div className="text-xs text-slate-500">{entity.subtitle}</div>}
        </div>
        {entity.status && <span className="text-xs text-emerald-700">{entity.status}</span>}
      </div>

      <div className="mt-3 grid gap-2">
        {entity.details.map((detail) => (
          <div key={`${entity.id}-${detail.label}`} className="rounded-xl bg-slate-50 p-3 text-xs text-slate-700">
            <div className="text-slate-500">{detail.label}</div>
            <div className="mt-1 break-all">{detail.value}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

function renderIntegrationHint(disturbanceCount: number) {
  return (
    <section className="rounded-2xl border border-sky-200 bg-gradient-to-r from-sky-50 via-white to-slate-50 p-4">
      <div className="text-sm font-semibold text-sky-800">跨链联动重点</div>
      <div className="mt-2 text-sm leading-6 text-slate-700">
        {disturbanceCount > 0
          ? '当前桥接流程已经叠加攻击或故障影响。优先观察“桥验证”和“目标执行”两个阶段，判断哪一步被绕过、伪造或延迟。'
          : '当前桥接流程处于基础闭环。建议先看消息如何穿过桥验证层，再结合事件时间线理解跨链请求是否真正完成。'}
      </div>
    </section>
  )
}

/**
 * 跨链可视化组件
 * 展示跨链桥、消息传递等跨链机制
 * 
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - 跨链生命周期清晰可见
 */
export default function CrossChainCanvas({ state, moduleKey }: CrossChainCanvasProps) {
  const data = (state.data || {}) as {
    title?: string
    summary?: string
    observationTips?: string[]
    metrics?: VisualizationMetricCard[]
    entities?: VisualizationEntityCard[]
    sections?: VisualizationStateSection[]
    events?: VisualizationEventItem[]
  }
  const globalData = (state.global_data || {}) as Record<string, unknown>

  const metrics = data.metrics || []
  const entities = data.entities || []
  const sections = data.sections || []
  const events = data.events || []
  const disturbances = buildDisturbances(globalData)
  const sceneLabel = getVisualizationModuleLabel(moduleKey, data.title || '跨链演示')

  // 跨链进度
  const progress = Math.min(100, Math.max(25, events.length * 20))

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#eef6ff_0%,#f5f3ff_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" icon={<GatewayOutlined />} className="m-0">
              主题: {sceneLabel}
            </Tag>
            <Tag color="cyan" className="m-0">步骤: {state.tick}</Tag>
            <Tag color="green" className="m-0">事件: {events.length}</Tag>
            {disturbances.length > 0 && (
              <Tag color="orange" className="m-0">联动: {disturbances.length}</Tag>
            )}
          </div>
          <div className="text-xs text-slate-500">
            {data.title || '跨链演示'}
          </div>
        </div>

        {/* 跨链进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>跨链进度</span>
            <span>{progress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-violet-500 transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="grid flex-1 grid-cols-1 gap-3 overflow-auto bg-[linear-gradient(180deg,#f7fbff_0%,#faf7ff_100%)] p-3 xl:grid-cols-[minmax(0,1fr)_288px]">
        <div className="min-w-0 space-y-3">
            {/* 场景说明 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="text-xs font-medium text-slate-900 mb-2">场景说明</div>
              <div className="text-xs text-slate-600">
                {data.summary || '展示跨链消息、证明、桥验证与目标链执行之间的关键链路和状态变化'}
              </div>
            </section>

            {/* 跨链生命周期轨道 */}
            {renderLifecycle(events)}

            {/* 联动提示 */}
            {renderIntegrationHint(disturbances.length)}

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
            {metrics.length > 0 && (
              <section className="grid gap-3 lg:grid-cols-4">
                {metrics.slice(0, 4).map((metric) => (
                  <div key={metric.key} className="rounded-lg border border-slate-200 bg-white p-3">
                    <div className="text-xs text-slate-500">{metric.label}</div>
                    <div className="mt-1 text-lg font-semibold text-sky-700">
                      {metric.value}
                    </div>
                    {metric.hint && (
                      <div className="mt-1 text-xs text-slate-500">{metric.hint}</div>
                    )}
                  </div>
                ))}
              </section>
            )}

            {/* 双链与桥接主舞台 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-3 text-xs font-medium text-slate-900">双链与桥接</div>
              <div className="grid gap-4 lg:grid-cols-[1fr_auto_1fr]">
                {/* 源链 */}
                <div className="space-y-2">
                  <div className="text-center text-xs text-sky-700">源链</div>
                  {entities.slice(0, 2).map(renderEntityCard)}
                  {entities.length === 0 && (
                    <div className="rounded border border-dashed border-slate-300 p-4 text-xs text-slate-500">
                      暂无源链数据
                    </div>
                  )}
                </div>

                {/* 桥接层 */}
                <div className="flex items-center justify-center">
                  <div className="w-28 rounded-full border border-sky-200 bg-sky-50 px-3 py-3 text-center">
                    <div className="text-xs font-medium text-sky-700">桥接层</div>
                    <div className="mt-1 text-xs text-slate-500">消息穿越</div>
                    <div className="mt-2 h-1 overflow-hidden rounded-full bg-slate-700">
                      <div className="h-full animate-pulse rounded-full bg-gradient-to-r from-cyan-400 to-violet-400" />
                    </div>
                  </div>
                </div>

                {/* 目标链 */}
                <div className="space-y-2">
                  <div className="text-center text-xs text-violet-700">目标链</div>
                  {entities.slice(2, 4).map(renderEntityCard)}
                  {entities.length <= 2 && (
                    <div className="rounded border border-dashed border-slate-300 p-4 text-xs text-slate-500">
                      暂无目标链数据
                    </div>
                  )}
                </div>
              </div>
            </section>

            {/* 状态和观察点 */}
            <section className="grid gap-3 lg:grid-cols-2">
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">状态分组</div>
                <div className="space-y-2">
                  {sections.length > 0 ? (
                    sections.slice(0, 3).map((section) => (
                      <div key={section.key} className="rounded bg-slate-50 p-2 text-xs">
                        <div className="font-medium text-slate-900">{section.title}</div>
                        <div className="mt-1 space-y-1">
                          {section.items.slice(0, 2).map((item) => (
                            <div key={item.label} className="flex justify-between text-slate-500">
                              <span>{item.label}</span>
                              <span className="text-slate-700">{item.value}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-xs text-slate-500">暂无状态分组</div>
                  )}
                </div>
              </div>

              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">教学观察点</div>
                <div className="space-y-2 text-xs text-slate-600">
                  {(data.observationTips || []).length > 0 ? (
                    data.observationTips?.slice(0, 3).map((tip, index) => (
                      <div key={index} className="rounded bg-slate-50 p-2">
                        <span className="mr-1 text-sky-600">{index + 1}.</span>
                        {tip}
                      </div>
                    ))
                  ) : (
                    <>
                      <div className="rounded bg-slate-50 p-2">
                        <span className="mr-1 text-sky-600">1.</span>
                        理解消息如何跨链传递
                      </div>
                      <div className="rounded bg-slate-50 p-2">
                        <span className="mr-1 text-sky-600">2.</span>
                        关注锁定、证明、签名等关键阶段
                      </div>
                      <div className="rounded bg-slate-50 p-2">
                        <span className="mr-1 text-sky-600">3.</span>
                        判断跨链请求是否完成闭环
                      </div>
                    </>
                  )}
                </div>
              </div>
            </section>
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <GatewayOutlined />
              <span>最近事件</span>
              <span className="ml-auto text-slate-500">({events.length})</span>
            </div>
            <div className="space-y-2 max-h-[500px] overflow-auto">
              {events.length > 0 ? (
                events.slice(0, 12).map((event) => (
                  <div key={event.id} className="rounded border border-slate-200 bg-slate-50 p-2 text-xs">
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-slate-900">{event.title}</span>
                      <span className="text-slate-500">步骤 {event.tick}</span>
                    </div>
                    <div className="text-slate-500 line-clamp-2">{event.summary}</div>
                  </div>
                ))
              ) : (
                <div className="rounded bg-slate-100 p-2 text-xs text-slate-500">
                  暂无跨链事件
                </div>
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}

