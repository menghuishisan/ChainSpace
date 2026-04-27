import { Tag } from 'antd'
import {
  ApartmentOutlined,
  DatabaseOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import type {
  EVMCanvasProps,
  EVMExecutionStats,
  EVMFrame,
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
    <div key={entity.id} className="rounded-xl border border-slate-200 bg-white p-3">
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

function renderFrames(frames: EVMFrame[]) {
  if (frames.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-slate-300 p-6 text-sm text-slate-500">
        当前场景还没有返回执行轨迹。
      </div>
    )
  }

  return frames.map((frame, index) => (
    <div key={frame.id} className="relative rounded-xl border border-slate-200 bg-white p-4">
      {index < frames.length - 1 && (
        <div className="pointer-events-none absolute left-7 top-full h-6 w-px bg-cyan-400/40" />
      )}
      <div className="flex items-start gap-3">
        <div className="mt-1 flex h-8 w-8 flex-none items-center justify-center rounded-full bg-cyan-500/15 text-xs font-semibold text-cyan-200">
          {frame.depth}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-3">
            <div className="text-sm font-medium text-slate-900">{frame.title}</div>
            <span className="text-xs text-slate-500">深度 {frame.depth}</span>
          </div>
          <div className="mt-2 text-xs leading-6 text-slate-600">{frame.description}</div>
          <div className="mt-3 grid gap-2 md:grid-cols-2">
            <div className="rounded bg-slate-50 p-2 text-xs text-slate-700">
              <div className="text-slate-500">当前指令</div>
              <div className="mt-1">{frame.opcode || '--'}</div>
            </div>
            <div className="rounded bg-slate-50 p-2 text-xs text-slate-700">
              <div className="text-slate-500">Gas 消耗</div>
              <div className="mt-1">{frame.gas || '--'}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  ))
}

/**
 * EVM可视化组件
 * 展示EVM执行、调用栈、Gas分析等
 *
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - 执行过程清晰可见
 */
export default function EVMCanvas({ state, moduleKey }: EVMCanvasProps) {
  const data = (state.data || {}) as {
    title?: string
    summary?: string
    observationTips?: string[]
    metrics?: VisualizationMetricCard[]
    entities?: VisualizationEntityCard[]
    sections?: VisualizationStateSection[]
    events?: VisualizationEventItem[]
    frames?: EVMFrame[]
    stats?: EVMExecutionStats
  }
  const globalData = (state.global_data || {}) as Record<string, unknown>

  const metrics = data.metrics || []
  const entities = data.entities || []
  const sections = data.sections || []
  const events = data.events || []
  const frames = data.frames || []
  const stats = data.stats || {}
  const disturbances = buildDisturbances(globalData)
  const stageProgress = Math.min(100, Math.max(18, frames.length * 18))
  const sceneLabel = getVisualizationModuleLabel(moduleKey, data.title || 'EVM执行过程可视化')

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#eef6ff_0%,#edf7ff_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="blue" icon={<ApartmentOutlined />} className="m-0">
              主题: {sceneLabel}
            </Tag>
            <Tag color="cyan" icon={<ThunderboltOutlined />} className="m-0">
              帧: {stats.frameCount ?? frames.length}
            </Tag>
            <Tag color="green" icon={<DatabaseOutlined />} className="m-0">
              状态: {stats.storageChanges ?? sections.length}
            </Tag>
            {disturbances.length > 0 && (
              <Tag color="orange" className="m-0">联动: {disturbances.length}</Tag>
            )}
          </div>
          <div className="text-xs text-slate-500">
            步骤: {state.tick}
          </div>
        </div>

        {/* 执行进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>执行进度</span>
            <span>{stageProgress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-sky-500 transition-all duration-500"
              style={{ width: `${stageProgress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="grid flex-1 grid-cols-1 gap-3 overflow-auto bg-[linear-gradient(180deg,#f7fbff_0%,#f4f7fb_100%)] p-3 xl:grid-cols-[minmax(0,1fr)_288px]">
        <div className="min-w-0 space-y-3">
            {/* 场景说明 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="text-xs font-medium text-slate-900 mb-2">
                {data.title || 'EVM执行过程可视化'}
              </div>
              <div className="text-xs text-slate-600">
                {data.summary || '观察调用如何逐层进入、指令如何推动状态变化，以及攻击如何改写执行路径'}
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

            {/* 执行过程轨道 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-500">执行过程轨道</div>
              <div className="flex items-center gap-1 overflow-x-auto pb-1">
                {(['进入调用', '执行展开', '状态写入', '返回结果']).map((item, index) => {
                  const active = stageProgress >= (index + 1) * 25
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

            {/* 执行调用链 */}
            <section className="rounded-lg border border-slate-200 bg-white p-3">
              <div className="mb-2 text-xs font-medium text-slate-900">执行调用链</div>
              <div className="space-y-2 max-h-[250px] overflow-auto">
                {renderFrames(frames)}
              </div>
            </section>

            {/* 参与对象和状态 */}
            <section className="grid gap-3 lg:grid-cols-2">
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">关键对象</div>
                <div className="space-y-2">
                  {entities.length > 0 ? (
                    entities.slice(0, 3).map(renderEntityCard)
                  ) : (
                    <div className="text-xs text-slate-500">暂无关键对象</div>
                  )}
                </div>
              </div>
              <div className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="text-xs font-medium text-slate-900 mb-2">状态变化</div>
                <div className="space-y-2">
                  {sections.length > 0 ? (
                    sections.slice(0, 3).map((section) => (
                      <div key={section.key} className="rounded bg-slate-50 p-2 text-xs">
                        <div className="font-medium text-slate-900">{section.title}</div>
                        <div className="mt-1 space-y-1">
                          {section.items.slice(0, 3).map((item) => (
                            <div key={item.label} className="flex justify-between">
                              <span className="text-slate-500">{item.label}</span>
                              <span className="text-slate-700">{item.value}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-xs text-slate-500">暂无状态变化</div>
                  )}
                </div>
              </div>
            </section>
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          {/* 教学观察点 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <ThunderboltOutlined />
              <span>教学观察点</span>
            </div>
            <div className="space-y-2 rounded bg-slate-50 p-2 text-xs">
              {(data.observationTips || []).length > 0 ? (
                data.observationTips?.slice(0, 3).map((tip, index) => (
                  <div key={index} className="text-slate-600">
                    <span className="mr-1 text-sky-600">{index + 1}.</span>
                    {tip}
                  </div>
                ))
              ) : (
                <>
                  <div className="text-slate-600">
                    <span className="mr-1 text-sky-600">1.</span>
                    看执行链如何逐层进入
                  </div>
                  <div className="text-slate-600">
                    <span className="mr-1 text-sky-600">2.</span>
                    观察Gas消耗和写入结果
                  </div>
                  <div className="text-slate-600">
                    <span className="mr-1 text-sky-600">3.</span>
                    注意执行路径是否被重定向
                  </div>
                </>
              )}
            </div>
          </section>

          {/* 最近事件 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <DatabaseOutlined />
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

