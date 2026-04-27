import type { AttackStageSceneProps } from '@/types/visualizationDomain'

function laneColor(depth: number): string {
  const colors = [
    'border-sky-500/40 bg-sky-500/10',
    'border-fuchsia-500/40 bg-fuchsia-500/10',
    'border-amber-500/40 bg-amber-500/10',
    'border-emerald-500/40 bg-emerald-500/10',
  ]

  return colors[(Math.max(depth, 1) - 1) % colors.length]
}

/**
 * 合约执行与权限利用类攻击主舞台。
 * 重点展示调用链、漏洞触发位置、状态修改顺序以及修复后是否阻断关键路径。
 */
export default function AttackExecutionScene({
  callFrames,
  storage,
  balances,
  attack,
  stats,
}: AttackStageSceneProps) {
  const latestBalance = balances[0]
  const progress = Math.min(100, Math.max(8, (stats.steps ?? 0) * 12))

  return (
    <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
      <section className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <div className="text-sm font-semibold text-white">调用链与漏洞触发</div>
            <div className="mt-1 text-xs leading-6 text-slate-400">
              先看谁发起调用，再看外部交互发生在状态更新之前还是之后。对执行类攻击来说，错误顺序往往就是漏洞本身。
            </div>
          </div>
          <div className="min-w-28 rounded-full bg-slate-900/80 px-3 py-2 text-center text-xs text-cyan-300">
            最大深度 {stats.attackDepth ?? 0}
          </div>
        </div>

        <div className="mb-4 rounded-2xl border border-slate-700 bg-slate-900/80 p-4">
          <div className="flex items-center justify-between gap-3 text-xs text-slate-400">
            <span>当前攻击推进</span>
            <span>{stats.steps ?? 0} / 8+ 步</span>
          </div>
          <div className="mt-3 h-3 overflow-hidden rounded-full bg-slate-800">
            <div
              className="h-full rounded-full bg-gradient-to-r from-cyan-400 via-fuchsia-400 to-orange-400 transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>

        <div className="space-y-3">
          {callFrames.length > 0 ? (
            callFrames.map((frame, index) => (
              <div
                key={frame.id}
                className={`rounded-2xl border p-4 shadow-[0_8px_30px_rgba(15,23,42,0.18)] ${laneColor(frame.depth)}`}
                style={{ marginLeft: `${Math.max(frame.depth - 1, 0) * 18}px` }}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="text-sm font-semibold text-white">{frame.title}</div>
                  <div className="rounded-full bg-slate-900/80 px-3 py-1 text-xs text-slate-300">
                    第 {frame.depth} 层
                  </div>
                </div>

                <div className="mt-2 text-xs leading-6 text-slate-300">{frame.description}</div>

                <div className="mt-4 grid gap-3 md:grid-cols-[1fr_auto_1fr_auto]">
                  <div className="rounded-xl bg-slate-900/80 px-3 py-2 text-xs text-slate-200">
                    <div className="text-slate-400">调用方</div>
                    <div className="mt-1 break-all">{frame.caller}</div>
                  </div>
                  <div className="flex items-center justify-center text-cyan-300">→</div>
                  <div className="rounded-xl bg-slate-900/80 px-3 py-2 text-xs text-slate-200">
                    <div className="text-slate-400">被调方</div>
                    <div className="mt-1 break-all">{frame.callee}</div>
                  </div>
                  <div className="rounded-xl bg-slate-900/80 px-3 py-2 text-xs text-orange-300">
                    <div className="text-slate-400">本步影响</div>
                    <div className="mt-1">{frame.amount}</div>
                  </div>
                </div>

                {index < callFrames.length - 1 && (
                  <div className="mt-4 h-1.5 overflow-hidden rounded-full bg-slate-900/80">
                    <div className="h-full w-full animate-pulse rounded-full bg-gradient-to-r from-red-400/80 via-orange-300/80 to-transparent" />
                  </div>
                )}
              </div>
            ))
          ) : (
            <div className="rounded-xl border border-dashed border-slate-600 p-6 text-sm leading-6 text-slate-400">
              当前攻击还没有形成可展开的执行路径。执行攻击动作后，这里应该能看到逐层调用、重入深度或权限切换过程。
            </div>
          )}
        </div>
      </section>

      <section className="space-y-4">
        <div className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
          <div className="text-sm font-semibold text-white">关键状态变化</div>
          <div className="mt-3 space-y-2">
            {storage.length > 0 ? (
              storage.map((slot) => (
                <div key={slot.key} className="rounded-xl bg-slate-900/80 p-3">
                  <div className="text-xs text-slate-400">{slot.label}</div>
                  <div className="mt-1 font-mono text-sm text-cyan-300">{slot.value}</div>
                </div>
              ))
            ) : (
              <div className="rounded-xl border border-dashed border-slate-600 p-4 text-sm leading-6 text-slate-400">
                当前场景还没有输出关键状态槽位。理想情况下，这里应显示余额、权限标记或已写入的关键存储字段。
              </div>
            )}
          </div>
        </div>

        <div className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
          <div className="text-sm font-semibold text-white">攻击结论</div>
          <div className="mt-3 text-sm leading-6 text-slate-300">
            {attack?.summary || '当前还没有形成清晰的攻击结论。'}
          </div>
          <div className="mt-4 rounded-xl bg-slate-900/80 p-3 text-xs leading-6 text-slate-300">
            <div>当前已完成 {stats.steps ?? 0} 个步骤，最大嵌套深度 {stats.attackDepth ?? 0}。</div>
            {latestBalance && (
              <div className="mt-2">
                重点观察 {latestBalance.label} 是否持续下降。如果关键资产不断减少，就说明攻击路径仍然在持续生效。
              </div>
            )}
          </div>
        </div>
      </section>
    </div>
  )
}
