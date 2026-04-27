import type { AttackStageSceneProps } from '@/types/visualizationDomain'

/**
 * 经济操纵类攻击主舞台。
 * 重点展示参与方、资金流、价格或清算阈值变化，以及攻击收益如何逐步形成。
 */
export default function AttackEconomicScene({
  balances,
  timeline,
  attack,
  stats,
}: AttackStageSceneProps) {
  const left = balances[0]
  const right = balances[1]
  const total = Math.max((left?.value || 0) + (right?.value || 0), 1)
  const victimWidth = `${Math.max(10, ((left?.value || 0) / total) * 100)}%`
  const attackerWidth = `${Math.max(10, ((right?.value || 0) / total) * 100)}%`

  return (
    <div className="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
      <section className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
        <div className="mb-4">
          <div className="text-sm font-semibold text-white">资金路径与价格冲击</div>
          <div className="mt-1 text-xs leading-6 text-slate-400">
            先判断资金从哪里流出、流向哪里，再看价格、清算或收益率在攻击过程中如何被扭曲。
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-[1fr_auto_1fr]">
          <div className="rounded-2xl border border-sky-500/30 bg-sky-500/10 p-4">
            <div className="text-sm font-semibold text-white">{left?.label || '受害方 / 目标协议'}</div>
            <div className="mt-3 text-3xl font-semibold text-sky-300">
              {left ? left.value.toLocaleString('zh-CN', { maximumFractionDigits: 2 }) : '--'}
            </div>
            <div className="mt-3 h-2 overflow-hidden rounded-full bg-slate-900/80">
              <div className="h-full rounded-full bg-sky-400/80 transition-all duration-500" style={{ width: victimWidth }} />
            </div>
            <div className="mt-3 text-xs leading-6 text-slate-300">
              {left?.description || '这里展示被操纵的一侧，例如池子储备、受害仓位或协议余额。'}
            </div>
          </div>

          <div className="flex items-center justify-center">
            <div className="w-28 rounded-full border border-slate-700 bg-slate-900/80 px-4 py-3 text-center text-xs text-slate-200">
              <div className="text-orange-300">攻击路径</div>
              <div className="mt-1">资金推进</div>
              <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-slate-800">
                <div className="h-full w-full animate-pulse rounded-full bg-gradient-to-r from-orange-400/80 via-yellow-300/80 to-transparent" />
              </div>
            </div>
          </div>

          <div className="rounded-2xl border border-orange-500/30 bg-orange-500/10 p-4">
            <div className="text-sm font-semibold text-white">{right?.label || '攻击者收益'}</div>
            <div className="mt-3 text-3xl font-semibold text-orange-300">
              {right ? right.value.toLocaleString('zh-CN', { maximumFractionDigits: 2 }) : '--'}
            </div>
            <div className="mt-3 h-2 overflow-hidden rounded-full bg-slate-900/80">
              <div className="h-full rounded-full bg-orange-400/80 transition-all duration-500" style={{ width: attackerWidth }} />
            </div>
            <div className="mt-3 text-xs leading-6 text-slate-300">
              {right?.description || '这里展示攻击者在这条路径中逐步获得的利润、可提取价值或额外头寸。'}
            </div>
          </div>
        </div>

        <div className="mt-4 grid gap-3 md:grid-cols-3">
          <div className="rounded-xl bg-slate-900/80 p-3">
            <div className="text-xs text-slate-400">累计步骤</div>
            <div className="mt-1 text-xl font-semibold text-white">{stats.steps ?? 0}</div>
          </div>
          <div className="rounded-xl bg-slate-900/80 p-3">
            <div className="text-xs text-slate-400">累计获利</div>
            <div className="mt-1 text-xl font-semibold text-orange-300">
              {(stats.drainedAmount ?? 0).toLocaleString('zh-CN', { maximumFractionDigits: 2 })}
            </div>
          </div>
          <div className="rounded-xl bg-slate-900/80 p-3">
            <div className="text-xs text-slate-400">剩余安全边际</div>
            <div className="mt-1 text-xl font-semibold text-cyan-300">
              {(stats.remainingBalance ?? 0).toLocaleString('zh-CN', { maximumFractionDigits: 2 })}
            </div>
          </div>
        </div>
      </section>

      <section className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
        <div className="text-sm font-semibold text-white">经济攻击时间线</div>
        <div className="mt-3 space-y-3">
          {timeline.length > 0 ? (
            timeline.map((item, index) => (
              <div key={item.id} className="rounded-xl border border-slate-700 bg-slate-900/80 p-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="text-sm font-medium text-white">
                    {index + 1}. {item.title}
                  </div>
                  <div className="rounded-full bg-orange-500/15 px-3 py-1 text-xs text-orange-300">
                    影响 {item.amount}
                  </div>
                </div>
                <div className="mt-2 text-xs leading-6 text-slate-300">{item.description}</div>
                <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-slate-800">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-orange-400/80 to-amber-300/80"
                    style={{ width: `${Math.min(100, 25 + index * 18)}%` }}
                  />
                </div>
              </div>
            ))
          ) : (
            <div className="rounded-xl border border-dashed border-slate-600 p-6 text-sm leading-6 text-slate-400">
              当前攻击还没有形成清晰的经济时间线。执行攻击动作后，这里应该能看到价格冲击、清算推进或套利获利过程。
            </div>
          )}
        </div>

        <div className="mt-4 rounded-xl bg-slate-900/80 p-3 text-sm leading-6 text-slate-300">
          {attack?.summary || '当前还没有形成明确的经济攻击结论。'}
        </div>
      </section>
    </div>
  )
}
