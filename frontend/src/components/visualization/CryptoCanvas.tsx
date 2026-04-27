import { Tag } from 'antd'
import {
  HistoryOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import type {
  CryptoCanvasProps,
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
      summary: `当前存在联动攻击：${String(item.type ?? '--')}，作用目标为 ${String(item.target ?? '--')}。`,
    })),
  ]
}

function renderEntityCard(entity: VisualizationEntityCard) {
  return (
    <div key={entity.id} className="rounded-lg border border-slate-200 bg-white p-3">
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="text-sm font-medium text-slate-900">{entity.title}</div>
          {entity.subtitle && <div className="text-xs text-slate-500">{entity.subtitle}</div>}
        </div>
        {entity.status && <span className="text-xs text-emerald-700">{entity.status}</span>}
      </div>
      <div className="mt-3 grid gap-2">
        {entity.details.map((detail) => (
          <div key={`${entity.id}-${detail.label}`} className="rounded bg-slate-50 p-2 text-xs text-slate-700">
            <div className="text-slate-500">{detail.label}</div>
            <div className="mt-1 break-all">{detail.value}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

/**
 * 密码学可视化组件
 * 展示哈希、签名、加密等密码学过程
 *
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - 密码学过程清晰可见
 */
export default function CryptoCanvas({ state, moduleKey }: CryptoCanvasProps) {
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
  const progress = Math.min(100, Math.max(20, entities.length * 18 + sections.length * 10))
  const sceneLabel = getVisualizationModuleLabel(moduleKey, data.title || '密码学过程演示')

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#f5f3ff_0%,#eef6ff_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" icon={<SafetyCertificateOutlined />} className="m-0">
              主题: {sceneLabel}
            </Tag>
            <Tag color="cyan" className="m-0">步骤: {state.tick}</Tag>
            <Tag color="green" className="m-0">事件: {events.length}</Tag>
          </div>
          <div className="text-xs text-slate-500">
            {data.title || '密码学过程演示'}
          </div>
        </div>

        {/* 过程进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>可视化进度</span>
            <span>{progress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-violet-500 to-cyan-500 transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="grid flex-1 grid-cols-1 gap-3 overflow-auto bg-[linear-gradient(180deg,#faf8ff_0%,#f7fbff_100%)] p-3 xl:grid-cols-[minmax(0,1fr)_288px]">
        <div className="min-w-0 space-y-3">
            {/* 场景说明 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="text-xs font-medium text-slate-900 mb-2">场景说明</div>
              <div className="text-xs text-slate-600">
                {data.summary || '观察输入如何被处理、结果如何生成，以及验证流程如何给出最终结论'}
              </div>
            </section>

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
              <div className="mb-2 text-xs font-medium text-slate-500">密码学过程轨道</div>
              <div className="flex items-center gap-1 overflow-x-auto pb-1">
                {(['输入准备', '中间处理', '生成输出', '完成验证']).map((item, index) => {
                  const active = progress >= (index + 1) * 25
                  return (
                    <div key={item} className="flex items-center">
                      <div
                        className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-xs whitespace-nowrap transition-all ${
                          active
                            ? 'border-violet-300 bg-violet-100 text-violet-800 shadow-[0_10px_25px_rgba(139,92,246,0.15)]'
                            : 'border-slate-200 bg-white text-slate-500'
                        }`}
                      >
                        {active && (
                          <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-violet-500" />
                        )}
                        <span>{item}</span>
                      </div>
                      {index < 3 && (
                        <div className={`mx-1 h-0.5 w-6 ${
                          active ? 'bg-violet-300' : 'bg-slate-300'
                        }`} />
                      )}
                    </div>
                  )
                })}
              </div>
            </section>

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

            {/* 关键对象和状态 */}
            <section className="grid gap-3 lg:grid-cols-2">
              {/* 关键对象 */}
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">关键对象</div>
                <div className="space-y-2">
                  {entities.length > 0 ? (
                    entities.map(renderEntityCard)
                  ) : (
                    <div className="rounded bg-slate-50 p-2 text-xs text-slate-500">
                      暂无密钥、摘要或证明对象
                    </div>
                  )}
                </div>
              </div>

              {/* 状态分组 */}
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">状态分组</div>
                <div className="space-y-2">
                  {sections.length > 0 ? (
                    sections.map((section) => (
                      <div key={section.key} className="rounded bg-slate-50 p-2 text-xs">
                        <div className="font-medium text-slate-900">{section.title}</div>
                        <div className="mt-1 space-y-1">
                          {section.items.slice(0, 3).map((item) => (
                            <div key={item.label} className="flex justify-between text-slate-500">
                              <span>{item.label}</span>
                              <span className="text-slate-700">{item.value}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="rounded bg-slate-50 p-2 text-xs text-slate-500">
                      暂无状态分组
                    </div>
                  )}
                </div>
              </div>
            </section>
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          {/* 教学观察点 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <SafetyCertificateOutlined />
              <span>教学观察点</span>
            </div>
            <div className="space-y-2 rounded bg-slate-50 p-2 text-xs">
              {(data.observationTips || []).length > 0 ? (
                data.observationTips?.map((tip, index) => (
                  <div key={index} className="text-slate-600">
                    <span className="mr-1 text-violet-600">{index + 1}.</span>
                    {tip}
                  </div>
                ))
              ) : (
                <>
                  <div className="text-slate-600">
                    <span className="mr-1 text-violet-600">1.</span>
                    对照输入和输出理解核心目标
                  </div>
                  <div className="text-slate-600">
                    <span className="mr-1 text-violet-600">2.</span>
                    看状态分组，确认验证是否通过
                  </div>
                  <div className="text-slate-600">
                    <span className="mr-1 text-violet-600">3.</span>
                    观察联动覆盖层是否改变结果
                  </div>
                </>
              )}
            </div>
          </section>

          {/* 最近事件 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <HistoryOutlined />
              <span>最近事件</span>
              <span className="ml-auto text-slate-500">({events.length})</span>
            </div>
            <div className="space-y-2 max-h-[300px] overflow-auto">
              {events.length > 0 ? (
                events.slice(0, 8).map((event) => (
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
                  暂无事件
                </div>
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}

