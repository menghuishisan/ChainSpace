import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

export function getDiscoveryActions(): VisualizationActionDefinition[] {
  return [
    {
      key: 'discovery-join',
      label: '模拟新节点加入',
      description: '触发一次完整的加入流程，观察新节点如何连接 Bootstrap 并逐步发现更多邻居。',
      kind: 'module_action',
      action: 'simulate_join',
      successMessage: '已开始一次新的节点发现流程',
    },
    {
      key: 'discovery-reset',
      label: '重置发现过程',
      description: '按给定参数重新初始化网络规模与 Bootstrap 节点，清空上一轮发现记录。',
      kind: 'module_action',
      action: 'reset_network',
      successMessage: '已重置当前节点发现场景',
      fields: [
        {
          key: 'network_size',
          label: '网络规模',
          type: 'number',
          defaultValue: 12,
        },
        {
          key: 'bootstrap_count',
          label: 'Bootstrap节点数',
          type: 'number',
          defaultValue: 2,
        },
      ],
    },
  ]
}
