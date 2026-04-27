import { Tag } from 'antd'
import type { VisualizationDisturbancePanelProps } from '@/types/visualizationDomain'

function groupTitle(type: string) {
  return type === 'fault' ? '故障联动' : '攻击联动'
}

function groupColor(type: string) {
  return type === 'fault' ? 'gold' : 'volcano'
}

/**
 * 联动覆盖层面板。
 * 统一展示当前实验中正在生效的故障与攻击，让学生先看到“主过程受到了什么影响”，
 * 再回到主舞台理解这些影响如何改变推进结果。
 */
export default function DisturbanceOverlayPanel({
  title = '联动影响覆盖层',
  items,
}: VisualizationDisturbancePanelProps) {
  if (items.length === 0) {
    return null
  }

  const grouped = items.reduce<Record<string, typeof items>>((result, item) => {
    const key = item.type === 'fault' ? 'fault' : 'attack'
    if (!result[key]) {
      result[key] = []
    }
    result[key].push(item)
    return result
  }, {})

  return (
    <section className="rounded-2xl border border-amber-400/25 bg-gradient-to-r from-amber-500/10 via-orange-500/10 to-transparent p-4">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-amber-100">{title}</div>
          <div className="mt-1 text-xs leading-6 text-slate-300">
            当前实验正在叠加的外部影响会直接改变主过程的推进节奏、结果判断与最终状态。
            建议先观察主舞台哪里被阻断、放大或偏转，再回到这里理解原因。
          </div>
        </div>
        <Tag color="orange" className="m-0">
          生效中 {items.length}
        </Tag>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {Object.entries(grouped).map(([key, group]) => (
          <div key={key} className="rounded-xl border border-slate-700 bg-slate-950/70 p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div className="text-sm font-medium text-white">{groupTitle(key)}</div>
              <Tag color={groupColor(key)} className="m-0">
                {group.length} 项
              </Tag>
            </div>

            <div className="space-y-3">
              {group.map((item) => (
                <div key={item.id} className="rounded-xl border border-slate-700 bg-slate-900/80 p-3">
                  <div className="flex items-center justify-between gap-3">
                    <div className="text-sm font-medium text-white">{item.label}</div>
                    <Tag color={groupColor(key)} className="m-0">
                      {item.type}
                    </Tag>
                  </div>
                  <div className="mt-2 text-xs text-slate-400">作用目标：{item.target}</div>
                  <div className="mt-2 text-sm leading-6 text-slate-200">{item.summary}</div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
