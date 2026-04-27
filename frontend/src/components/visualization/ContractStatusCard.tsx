import { Tag } from 'antd'
import type { VisualizationContractCardProps } from '@/types/visualizationDomain'

export default function ContractStatusCard({
  title,
  tagColor,
  tagLabel,
  value,
  accentClass,
  description,
  tips,
}: VisualizationContractCardProps) {
  return (
    <div className="rounded-xl border border-slate-700 bg-slate-950/70 p-4">
      <div className="flex items-center justify-between">
        <div className="text-sm font-medium text-white">{title}</div>
        <Tag color={tagColor}>{tagLabel}</Tag>
      </div>
      <div className={`mt-4 text-3xl font-semibold ${accentClass}`}>
        {value.toLocaleString('zh-CN', { maximumFractionDigits: 2 })}
      </div>
      <div className="mt-2 text-xs text-slate-400">{description}</div>
      <div className="mt-4 space-y-2 text-xs text-slate-300">
        {tips.map((item) => (
          <div key={item} className="rounded bg-slate-900/80 p-2">
            {item}
          </div>
        ))}
      </div>
    </div>
  )
}
