import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

/**
 * AMM 可视化动作定义。
 * 这些动作围绕学生最需要观察的三条主链展开：
 * 兑换、添加流动性、移除流动性。
 */
export function getAmmActions(): VisualizationActionDefinition[] {
  return [
    {
      key: 'amm-swap',
      label: '模拟交易',
      description: '执行一笔兑换，观察储备、价格、滑点和曲线位置如何变化。',
      kind: 'module_action',
      action: 'swap',
      successMessage: '已执行一笔 AMM 兑换。',
      fields: [
        {
          key: 'token_in',
          label: '输入资产',
          type: 'select',
          defaultValue: 'ETH',
          options: [
            { label: 'ETH', value: 'ETH' },
            { label: 'USDC', value: 'USDC' },
          ],
        },
        {
          key: 'amount_in',
          label: '输入数量',
          type: 'number',
          defaultValue: 10000,
          min: 1,
          max: 1000000,
        },
      ],
    },
    {
      key: 'amm-add-liquidity',
      label: '添加流动性',
      description: '向资金池同时注入两侧资产，观察 TVL、储备规模和曲线范围如何变化。',
      kind: 'module_action',
      action: 'add_liquidity',
      successMessage: '已添加流动性。',
      fields: [
        {
          key: 'amount_a',
          label: 'ETH 数量',
          type: 'number',
          defaultValue: 20000,
          min: 1,
          max: 1000000,
        },
        {
          key: 'amount_b',
          label: 'USDC 数量',
          type: 'number',
          defaultValue: 20000,
          min: 1,
          max: 1000000,
        },
      ],
    },
    {
      key: 'amm-remove-liquidity',
      label: '移除流动性',
      description: '赎回 LP 份额对应的资产，观察池子规模、储备和价格区间如何回落。',
      kind: 'module_action',
      action: 'remove_liquidity',
      successMessage: '已移除流动性。',
      fields: [
        {
          key: 'lp_shares',
          label: 'LP 份额',
          type: 'number',
          defaultValue: 10000,
          min: 1,
          max: 1000000,
        },
      ],
    },
  ]
}
