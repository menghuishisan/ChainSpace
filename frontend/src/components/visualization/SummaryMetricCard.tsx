import type { VisualizationSummaryMetricProps } from '@/types/visualizationDomain'

export default function SummaryMetricCard({
  label,
  value,
  accentClass,
}: VisualizationSummaryMetricProps) {
  return (
    <div className="rounded-xl border border-slate-700 bg-slate-950/70 p-4">
      <div className="text-sm font-medium text-white">{label}</div>
      <div className={`mt-3 text-2xl font-semibold ${accentClass}`}>{value}</div>
    </div>
  )
}
