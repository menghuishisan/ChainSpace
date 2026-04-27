import type { KeepAliveTabPanelProps } from '@/types/presentation'

/**
 * 保持挂载的页签面板。
 * 通过显示/隐藏切换页签，避免实验环境、日志或可视化面板被重复卸载。
 */
export default function KeepAliveTabPanel({ active, children }: KeepAliveTabPanelProps) {
  return <div className={active ? 'h-full' : 'hidden'}>{children}</div>
}
