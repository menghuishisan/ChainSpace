import type { AttackStageSceneProps } from '@/types/visualizationDomain'

/**
 * 跨链桥与消息验证类攻击主舞台。
 * 重点展示源链、桥验证层、目标链之间哪一步被伪造、绕过或重复利用。
 */
export default function AttackBridgeScene({
  timeline,
  sections,
  attack,
  events,
}: AttackStageSceneProps) {
  const lifecycleSection = sections.find((section) => /验证|桥|签名|消息|链/.test(section.title)) || sections[0]
  const displayedSteps = timeline.length > 0 ? timeline : events.map((event) => ({
    id: event.id,
    title: event.title,
    description: event.summary,
    amount: '',
    depth: 0,
  }))

  return (
    <div className="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
      <section className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
        <div className="mb-4">
          <div className="text-sm font-semibold text-white">跨链消息生命周期</div>
          <div className="mt-1 text-xs leading-6 text-slate-400">
            先看消息从源链进入桥验证层，再看签名、证明或挑战期在哪一步被绕过，最后判断目标链为什么会错误执行。
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-[1fr_auto_1fr_auto_1fr]">
          <div className="rounded-2xl border border-sky-500/30 bg-sky-500/10 p-4">
            <div className="text-sm font-semibold text-white">源链</div>
            <div className="mt-3 text-xs leading-6 text-slate-300">锁定、存款、原始消息</div>
          </div>

          <div className="flex items-center justify-center text-cyan-300">→</div>

          <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 p-4">
            <div className="text-sm font-semibold text-white">桥验证层</div>
            <div className="mt-3 text-xs leading-6 text-slate-300">签名、证明、挑战期、验证者阈值</div>
          </div>

          <div className="flex items-center justify-center text-cyan-300">→</div>

          <div className="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-4">
            <div className="text-sm font-semibold text-white">目标链</div>
            <div className="mt-3 text-xs leading-6 text-slate-300">铸造、解锁、执行结果</div>
          </div>
        </div>

        <div className="mt-4 rounded-2xl border border-slate-700 bg-slate-900/80 p-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <div className="text-sm font-semibold text-white">攻击推进轨道</div>
            <div className="text-xs text-slate-400">已记录 {displayedSteps.length} 步</div>
          </div>
          <div className="grid gap-3 md:grid-cols-4">
            {['发起请求', '伪造验证', '穿透桥层', '目标异常'].map((item, index) => {
              const active = displayedSteps.length >= index + 1
              return (
                <div
                  key={item}
                  className={`rounded-xl border p-3 text-sm ${
                    active
                      ? 'border-red-400/40 bg-red-500/10 text-red-100'
                      : 'border-slate-700 bg-slate-950/80 text-slate-300'
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
        </div>

        {lifecycleSection && (
          <div className="mt-4 rounded-2xl border border-slate-700 bg-slate-900/80 p-4">
            <div className="text-sm font-semibold text-white">{lifecycleSection.title}</div>
            <div className="mt-3 grid gap-3 md:grid-cols-2">
              {lifecycleSection.items.slice(0, 8).map((item) => (
                <div key={`${lifecycleSection.key}-${item.label}`} className="rounded-xl bg-slate-950/80 p-3">
                  <div className="text-xs text-slate-400">{item.label}</div>
                  <div className="mt-1 text-sm text-white">{item.value}</div>
                </div>
              ))}
            </div>
          </div>
        )}
      </section>

      <section className="space-y-4">
        <div className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
          <div className="text-sm font-semibold text-white">攻击步骤</div>
          <div className="mt-3 space-y-3">
            {displayedSteps.length > 0 ? (
              displayedSteps.slice(0, 8).map((item, index) => (
                <div key={item.id} className="rounded-xl border border-slate-700 bg-slate-900/80 p-3">
                  <div className="text-sm font-medium text-white">
                    {index + 1}. {item.title}
                  </div>
                  <div className="mt-2 text-xs leading-6 text-slate-300">{item.description}</div>
                  <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-slate-800">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-red-400/80 to-orange-300/80"
                      style={{ width: `${Math.min(100, 25 + index * 18)}%` }}
                    />
                  </div>
                </div>
              ))
            ) : (
              <div className="rounded-xl border border-dashed border-slate-600 p-6 text-sm leading-6 text-slate-400">
                当前攻击还没有形成明确的桥攻击步骤。
              </div>
            )}
          </div>
        </div>

        <div className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4 text-sm leading-6 text-slate-300">
          {attack?.summary || '当前还没有形成明确的跨链攻击结论。'}
        </div>
      </section>
    </div>
  )
}
