import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

/**
 * 跨链桥演示动作定义。
 * 覆盖锁定-铸造、流动性桥和重置场景三条实验主链。
 */
export function getBridgeActions(): VisualizationActionDefinition[] {
  return [
    {
      key: 'bridge-lock-mint',
      label: '锁定-铸造流程',
      description: '执行一条完整的锁定-铸造跨链路径，观察确认、签名和目标链执行。',
      kind: 'module_action',
      action: 'simulate_lock_mint',
      successMessage: '已完成一次锁定-铸造跨链流程。',
      fields: [
        {
          key: 'user',
          label: '用户标识',
          type: 'text',
          defaultValue: 'student',
        },
        {
          key: 'amount',
          label: '转移数量',
          type: 'number',
          defaultValue: 1,
          min: 1,
          max: 100,
        },
      ],
    },
    {
      key: 'bridge-liquidity',
      label: '流动性桥流程',
      description: '执行一次流动性桥跨链，观察即时放款和后续再平衡过程。',
      kind: 'module_action',
      action: 'simulate_liquidity_bridge',
      successMessage: '已完成一次流动性桥跨链流程。',
      fields: [
        {
          key: 'user',
          label: '用户标识',
          type: 'text',
          defaultValue: 'student',
        },
        {
          key: 'amount',
          label: '转移数量',
          type: 'number',
          defaultValue: 1,
          min: 1,
          max: 100,
        },
      ],
    },
    {
      key: 'bridge-reset',
      label: '重置桥场景',
      description: '恢复到当前桥类型与安全模型下的初始状态，便于重新演示完整闭环。',
      kind: 'module_action',
      action: 'reset_bridge',
      successMessage: '已重置跨链桥场景。',
    },
  ]
}
