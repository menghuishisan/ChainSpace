import type { AttackStageSceneProps } from '@/types/visualizationDomain'

/**
 * 共识与链安全攻击主舞台。
 * 重点展示诚实链、攻击链、投票或算力偏移，以及主导权如何逐步发生转移。
 */
export default function AttackConsensusScene({
  sections,
  timeline,
  stats,
  attack,
}: AttackStageSceneProps) {
  const chainSection = sections.find((section) =>
    /链|分叉|投票|哈希|奖励/.test(section.title),
  ) || sections[0]
  const honestLength = Math.max(3, Math.min((stats.steps ?? 0) + 1, 6))
  const attackLength = Math.max(2, Math.min((stats.attackDepth ?? 0) + 1, 6))

  return (
    <div className="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
      <section className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
        <div className="mb-4">
          <div className="text-sm font-semibold text-white">链竞争与主导权转移</div>
          <div className="mt-1 text-xs leading-6 text-slate-400">
            先看诚实链和攻击链如何分叉，再看攻击者是否通过私有出块、投票操纵或奖励诱导取得主导权。
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-2">
          <div className="rounded-2xl border border-sky-500/30 bg-sky-500/10 p-4">
            <div className="text-sm font-semibold text-white">公开链 / 诚实路径</div>
            <div className="mt-4 space-y-2">
              {Array.from({ length: honestLength }, (_, index) => (
                <div key={`honest-${index}`} className="rounded-xl bg-slate-900/80 p-3 text-xs text-slate-200">
                  诚实区块 {index + 1}
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-2xl border border-red-500/30 bg-red-500/10 p-4">
            <div className="text-sm font-semibold text-white">攻击链 / 作恶路径</div>
            <div className="mt-4 space-y-2">
              {Array.from({ length: attackLength }, (_, index) => (
                <div key={`attack-${index}`} className="rounded-xl bg-slate-900/80 p-3 text-xs text-slate-200">
                  攻击区块 {index + 1}
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-4 rounded-2xl border border-slate-700 bg-slate-900/80 p-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <div className="text-sm font-semibold text-white">攻击推进轨道</div>
            <div className="text-xs text-slate-400">攻击深度 {stats.attackDepth ?? 0}</div>
          </div>
          <div className="grid gap-3 md:grid-cols-4">
            {['制造分叉', '扩大优势', '替换结果', '完成获利'].map((item, index) => {
              const active = (stats.steps ?? 0) >= index + 1
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

        {chainSection && (
          <div className="mt-4 rounded-2xl border border-slate-700 bg-slate-900/80 p-4">
            <div className="text-sm font-semibold text-white">{chainSection.title}</div>
            <div className="mt-3 grid gap-3 md:grid-cols-2">
              {chainSection.items.slice(0, 6).map((item) => (
                <div key={`${chainSection.key}-${item.label}`} className="rounded-xl bg-slate-950/80 p-3">
                  <div className="text-xs text-slate-400">{item.label}</div>
                  <div className="mt-1 text-sm text-white">{item.value}</div>
                </div>
              ))}
            </div>
          </div>
        )}
      </section>

      <section className="rounded-2xl border border-slate-700 bg-slate-950/70 p-4">
        <div className="text-sm font-semibold text-white">攻击推进时间线</div>
        <div className="mt-3 space-y-3">
          {timeline.length > 0 ? (
            timeline.map((item, index) => (
              <div key={item.id} className="rounded-xl border border-slate-700 bg-slate-900/80 p-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="text-sm font-medium text-white">
                    {index + 1}. {item.title}
                  </div>
                  <div className="rounded-full bg-red-500/15 px-3 py-1 text-xs text-red-300">
                    阶段影响 {item.amount}
                  </div>
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
              当前攻击还没有形成链竞争时间线。执行攻击动作后，这里应该能看到分叉形成、私有链推进或投票扭曲过程。
            </div>
          )}
        </div>

        <div className="mt-4 rounded-xl bg-slate-900/80 p-3 text-sm leading-6 text-slate-300">
          {attack?.summary || '当前还没有形成明确的链安全攻击结论。'}
        </div>
      </section>
    </div>
  )
}
