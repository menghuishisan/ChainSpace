import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

/**
 * EVM 调用跟踪动作定义。
 * 覆盖典型调用链场景与重置流程。
 */
export function getCallTraceActions(): VisualizationActionDefinition[] {
  return [
    {
      key: 'call-trace-simulate',
      label: '生成调用链',
      description: '按选定场景生成一条完整调用跟踪，观察调用层级、Gas 消耗和调用类型。',
      kind: 'module_action',
      action: 'simulate_trace',
      successMessage: '已生成一条完整调用跟踪。',
      fields: [
        {
          key: 'scenario',
          label: '调用场景',
          type: 'select',
          defaultValue: 'defi_swap',
          options: [
            { label: '简单转账', value: 'simple_transfer' },
            { label: 'ERC20 转账', value: 'token_transfer' },
            { label: 'DeFi 兑换', value: 'defi_swap' },
            { label: '嵌套调用', value: 'nested_calls' },
          ],
        },
      ],
    },
    {
      key: 'call-trace-reset',
      label: '重置跟踪',
      description: '清空当前调用轨迹和统计结果，便于重新选择场景演示。',
      kind: 'module_action',
      action: 'reset_trace',
      successMessage: '已重置调用跟踪场景。',
    },
  ]
}
