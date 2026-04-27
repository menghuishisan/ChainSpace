import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

export function getTopologyActions(): VisualizationActionDefinition[] {
  return [
    {
      key: 'topology-rebuild',
      label: '重新生成拓扑',
      description: '按照当前参数重新构建网络拓扑，观察连边结构和平均度数如何变化。',
      kind: 'module_action',
      action: 'rebuild_topology',
      successMessage: '已重新生成当前网络拓扑',
      fields: [
        {
          key: 'node_count',
          label: '节点数量',
          type: 'number',
          defaultValue: 8,
        },
        {
          key: 'topology_type',
          label: '拓扑类型',
          type: 'select',
          defaultValue: 'star',
          options: [
            { label: '全连接', value: 'full_mesh' },
            { label: '环形', value: 'ring' },
            { label: '星形', value: 'star' },
            { label: '树形', value: 'tree' },
            { label: '随机图', value: 'random' },
            { label: '小世界', value: 'small_world' },
            { label: '无标度', value: 'scale_free' },
          ],
        },
      ],
    },
    {
      key: 'topology-toggle-node',
      label: '切换节点在线状态',
      description: '让指定节点离线或恢复在线，观察路径冗余和连通性变化。',
      kind: 'module_action',
      action: 'toggle_node',
      successMessage: '已更新节点在线状态',
      fields: [
        {
          key: 'target',
          label: '目标节点',
          type: 'select',
          defaultValue: 'node-0',
          options: [
            { label: 'node-0', value: 'node-0' },
          ],
        },
      ],
    },
  ]
}
