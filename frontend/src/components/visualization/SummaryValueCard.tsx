import type { VisualizationSummaryCardProps } from '@/types/visualizationDomain'

export default function SummaryValueCard({
  title,
  value,
  hint,
  valueClassName,
}: VisualizationSummaryCardProps) {
  return (
    <div className="rounded-xl border border-slate-700 bg-slate-950/70 p-4">
      <div className="text-sm font-medium text-white">{title}</div>
      <div className={`mt-3 text-3xl font-semibold ${valueClassName}`}>{value}</div>
      <div className="mt-2 text-xs text-slate-400">{hint}</div>
    </div>
  )
}
