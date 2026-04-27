import { Tag } from 'antd'
import { AlertOutlined, HistoryOutlined, WalletOutlined } from '@ant-design/icons'
import type {
  AttackCanvasData,
  AttackCanvasProps,
  AttackMechanism,
  AttackStageSceneProps,
} from '@/types/visualizationDomain'
import { getAttackMechanism, getAttackMechanismLabel } from '@/domains/visualization/runtime/visualizationAttack'
import { getVisualizationModuleLabel } from '@/domains/visualization/runtime/visualizationMeta'
import AttackBridgeScene from './AttackBridgeScene'
import AttackConsensusScene from './AttackConsensusScene'
import AttackEconomicScene from './AttackEconomicScene'
import AttackExecutionScene from './AttackExecutionScene'

const ATTACK_FLOW: Record<AttackMechanism, string[]> = {
  execution: ['入口识别', '危险调用', '状态偏移', '结果放大'],
  economic: ['制造窗口', '扭曲价格', '放大利润', '错误结算'],
  consensus: ['制造分叉', '扩大优势', '替换结果', '完成获利'],
  bridge: ['源链请求', '验证绕过', '目标执行', '资产异常'],
}

function buildDisturbances(globalData: Record<string, unknown>) {
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

function renderStageScene(mechanism: AttackMechanism, props: AttackStageSceneProps) {
  switch (mechanism) {
    case 'economic':
      return <AttackEconomicScene {...props} />
    case 'consensus':
      return <AttackConsensusScene {...props} />
    case 'bridge':
      return <AttackBridgeScene {...props} />
    case 'execution':
    default:
      return <AttackExecutionScene {...props} />
  }
}

function renderAttackFlow(mechanism: AttackMechanism, stepCount: number) {
  const flow = ATTACK_FLOW[mechanism]

  return (
    <section className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold text-slate-900">攻击过程轨道</div>
        <div className="text-xs text-slate-500">已推进 {stepCount} 步</div>
      </div>
      <div className="grid gap-3 md:grid-cols-4">
        {flow.map((item, index) => {
          const active = stepCount >= index + 1
          return (
            <div
              key={item}
              className={`rounded-xl border p-3 text-sm ${
                active
                  ? 'border-rose-200 bg-rose-50 text-rose-800'
                  : 'border-slate-200 bg-white text-slate-600'
              }`}
            >
              <div className="flex items-center justify-between gap-3">
                <span className="font-medium">{item}</span>
                <span className="text-xs">{index + 1}</span>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

function renderMechanismHint(mechanism: AttackMechanism) {
  const hints: Record<AttackMechanism, string> = {
    execution: '优先观察调用链哪里偏离了安全路径，以及哪一个状态写入成为攻击的放大点。',
    economic: '优先观察价格、池子或仓位在何处被扭曲，以及收益如何被逐步累积。',
    consensus: '优先观察诚实链和攻击链的竞争过程，判断主导权在哪一步发生了转移。',
    bridge: '优先观察跨链生命周期里哪一步验证被绕过，导致目标链出现错误执行。',
  }

  return (
    <section className="rounded-2xl border border-rose-200 bg-gradient-to-r from-rose-50 via-orange-50 to-white p-4">
      <div className="text-sm font-semibold text-rose-800">当前联动重点</div>
      <div className="mt-2 text-sm leading-6 text-slate-700">{hints[mechanism]}</div>
    </section>
  )
}

/**
 * 攻击可视化组件
 * 展示合约攻击、经济攻击、共识攻击等
 * 
 * 优化要点：
 * - 清晰的视觉层次
 * - 左右分栏布局
 * - 攻击过程清晰可见
 */
export default function AttackCanvas({ state, moduleKey }: AttackCanvasProps) {
  const data = (state.data || {}) as AttackCanvasData
  const globalData = (state.global_data || {}) as Record<string, unknown>
  const mechanism = getAttackMechanism({
    moduleKey: moduleKey || 'attacks/reentrancy',
    renderer: 'attack',
    simulatorId: moduleKey?.split('/')[1] || 'reentrancy',
  })
  const sceneLabel = getVisualizationModuleLabel(moduleKey, getAttackMechanismLabel(mechanism))

  const sceneProps: AttackStageSceneProps = {
    moduleKey: moduleKey || 'attacks/reentrancy',
    metrics: data.metrics || [],
    actors: data.actors || [],
    sections: data.sections || [],
    timeline: data.timeline || [],
    callFrames: data.callFrames || [],
    storage: data.storage || [],
    balances: data.balances || [],
    stats: data.stats || {},
    attack: data.attack,
    events: data.events || [],
  }

  const stats = sceneProps.stats
  const disturbances = buildDisturbances(globalData)

  // 攻击步骤进度
  const attackProgress = Math.min(100, (stats.steps ?? 0) * 20)

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== 顶部信息栏 ===== */}
      <div className="shrink-0 border-b border-slate-200 bg-[linear-gradient(135deg,#fff1f2_0%,#fff7ed_100%)] p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Tag color="red" icon={<AlertOutlined />} className="m-0">
              攻击深度: {stats.attackDepth ?? 0}
            </Tag>
            <Tag color="orange" icon={<HistoryOutlined />} className="m-0">
              步骤: {stats.steps ?? 0}
            </Tag>
            <Tag color="blue" icon={<WalletOutlined />} className="m-0">
              影响: {(stats.drainedAmount ?? 0).toLocaleString('zh-CN', { maximumFractionDigits: 2 })}
            </Tag>
            <Tag color="purple" className="m-0">
              {getAttackMechanismLabel(mechanism)}
            </Tag>
          </div>
          <div className="text-xs text-slate-500">
            {sceneLabel} · 步骤 {state.tick}
          </div>
        </div>

        {/* 攻击进度条 */}
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>攻击进度</span>
            <span>{attackProgress.toFixed(0)}%</span>
          </div>
          <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-slate-200">
            <div
              className="h-full rounded-full bg-gradient-to-r from-red-500 to-orange-500 transition-all duration-500"
              style={{ width: `${attackProgress}%` }}
            />
          </div>
        </div>
      </div>

      {/* ===== 主可视化区域 ===== */}
      <div className="grid flex-1 grid-cols-1 gap-3 overflow-auto bg-[linear-gradient(180deg,#fff8f8_0%,#f8fafc_100%)] p-3 xl:grid-cols-[minmax(0,1fr)_288px]">
        <div className="min-w-0 space-y-3">
            {/* 关键指标 */}
            {sceneProps.metrics && sceneProps.metrics.length > 0 && (
              <section className="grid gap-3 lg:grid-cols-4">
                {sceneProps.metrics.slice(0, 4).map((metric) => (
                  <div key={metric.key} className="rounded-lg border border-slate-200 bg-white p-3">
                    <div className="text-xs text-slate-500">{metric.label}</div>
                    <div className="mt-1 text-lg font-semibold text-rose-700">
                      {metric.value}
                    </div>
                    {metric.hint && (
                      <div className="mt-1 text-xs text-slate-500">{metric.hint}</div>
                    )}
                  </div>
                ))}
              </section>
            )}

            {/* 攻击过程轨道 */}
            {renderAttackFlow(mechanism, stats.steps ?? 0)}

            {/* 机制提示 */}
            {renderMechanismHint(mechanism)}

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

            {/* 参与方 */}
            {sceneProps.actors && sceneProps.actors.length > 0 && (
              <section className="grid gap-3 lg:grid-cols-2">
                {sceneProps.actors.slice(0, 4).map((actor) => (
                  <div key={actor.id} className="rounded-lg border border-slate-200 bg-white p-3">
                    <div className="flex items-center justify-between mb-2">
                      <div>
                        <div className="text-xs font-medium text-slate-900">{actor.title}</div>
                        {actor.subtitle && (
                          <div className="text-xs text-slate-500">{actor.subtitle}</div>
                        )}
                      </div>
                      {actor.status && (
                        <span className="text-xs text-amber-700">{actor.status}</span>
                      )}
                    </div>
                    <div className="grid gap-1">
                      {actor.details.slice(0, 3).map((detail) => (
                        <div key={`${actor.id}-${detail.label}`} className="rounded bg-slate-50 p-1.5 text-xs">
                          <span className="text-slate-500">{detail.label}: </span>
                          <span className="text-slate-700">{detail.value}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </section>
            )}

            {/* 攻击场景主舞台 */}
            {renderStageScene(mechanism, sceneProps)}

            {/* 观察重点 */}
            {data.observationTips && data.observationTips.length > 0 && (
              <section className="rounded-lg border border-slate-200 bg-white p-3">
                <div className="mb-2 text-xs font-medium text-slate-500">观察重点</div>
                <div className="grid gap-2 md:grid-cols-3">
                  {data.observationTips.slice(0, 3).map((tip, index) => (
                    <div key={index} className="rounded bg-slate-50 p-2 text-xs text-slate-600">
                      <span className="mr-1 text-rose-600">{index + 1}.</span>
                      {tip}
                    </div>
                  ))}
                </div>
              </section>
            )}
        </div>

        <aside className="space-y-3 xl:sticky xl:top-0 xl:self-start">
          {/* 攻击结果摘要 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <AlertOutlined />
              <span>攻击结果</span>
            </div>
            <div className="rounded bg-slate-50 p-2 text-xs">
              <div className="mb-1">
                <span className="text-slate-500">标题: </span>
                <span className="text-slate-900">{data.attack?.title || '未开始'}</span>
              </div>
              <div className="mb-1">
                <span className="text-slate-500">深度: </span>
                <span className="text-rose-700">{data.attack?.depth ?? stats.attackDepth ?? 0}</span>
              </div>
              <div className="mb-1">
                <span className="text-slate-500">步骤: </span>
                <span className="text-amber-700">{data.attack?.completedSteps ?? stats.steps ?? 0}</span>
              </div>
              <div className="mt-2 text-slate-600">
                {data.attack?.summary || '等待攻击动作触发后更新'}
              </div>
            </div>
          </section>

          {/* 最近事件 */}
          <section className="rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium text-slate-900">
              <HistoryOutlined />
              <span>最近事件</span>
              <span className="ml-auto text-slate-500">({sceneProps.events.length})</span>
            </div>
            <div className="space-y-2 max-h-[400px] overflow-auto">
              {sceneProps.events.length > 0 ? (
                sceneProps.events.slice(0, 10).map((event) => (
                  <div key={event.id} className="rounded border border-slate-200 bg-slate-50 p-2 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-slate-900">{event.title}</span>
                      <span className="text-slate-500">步骤 {event.tick}</span>
                    </div>
                    <div className="mt-1 text-slate-500 line-clamp-2">{event.summary}</div>
                  </div>
                ))
              ) : (
                <div className="rounded bg-slate-100 p-2 text-xs text-slate-500">
                  先执行攻击动作，观察过程
                </div>
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}

