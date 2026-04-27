import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

export function getBlockchainActions(mode: 'chain' | 'merkle' = 'chain'): VisualizationActionDefinition[] {
  if (mode === 'merkle') {
    return [
      {
        key: 'merkle-add-leaf',
        label: '添加叶子',
        description: '向 Merkle 树加入一条新数据，观察叶子、父节点和根哈希如何联动变化。',
        kind: 'module_action',
        action: 'add_leaf',
        successMessage: '已向 Merkle 树添加新叶子',
        fields: [
          {
            key: 'data',
            label: '叶子数据',
            type: 'text',
            defaultValue: 'transaction:alice->bob:10',
          },
        ],
      },
      {
        key: 'merkle-generate-proof',
        label: '生成证明',
        description: '为指定叶子生成证明路径，观察兄弟节点哈希如何组成验证路径。',
        kind: 'module_action',
        action: 'generate_proof',
        successMessage: '已生成证明路径',
        fields: [
          {
            key: 'leaf_index',
            label: '叶子索引',
            type: 'number',
            defaultValue: 0,
          },
        ],
      },
      {
        key: 'merkle-verify-proof',
        label: '验证证明',
        description: '使用最近一次生成的证明执行验证，观察验证结果与根哈希是否一致。',
        kind: 'module_action',
        action: 'verify_proof',
        successMessage: '已完成证明验证',
      },
      {
        key: 'merkle-reset',
        label: '重置树',
        description: '清空当前树结构与证明记录，回到初始状态重新演示。',
        kind: 'module_action',
        action: 'reset_tree',
        successMessage: '已重置当前 Merkle 树场景',
      },
    ]
  }

  return [
    {
      key: 'blockchain-add-tx',
      label: '添加交易',
      description: '向当前区块加入一笔交易，观察 Merkle Root 和区块哈希变化。',
      kind: 'module_action',
      action: 'add_transaction',
      successMessage: '已向当前区块添加交易',
    },
    {
      key: 'blockchain-mine',
      label: '模拟挖矿',
      description: '尝试找到满足难度的 Nonce，观察区块完成状态和链高度变化。',
      kind: 'module_action',
      action: 'mine_block',
      successMessage: '已完成一次区块挖矿演示',
    },
    {
      key: 'blockchain-update-field',
      label: '修改区块字段',
      description: '手动修改区块头字段，观察哈希如何立刻发生变化。',
      kind: 'module_action',
      action: 'set_field',
      successMessage: '已更新当前区块字段',
      fields: [
        {
          key: 'field',
          label: '字段名',
          type: 'select',
          defaultValue: 'nonce',
          options: [
            { label: 'version', value: 'version' },
            { label: 'prev_block_hash', value: 'prev_block_hash' },
            { label: 'timestamp', value: 'timestamp' },
            { label: 'difficulty', value: 'difficulty' },
            { label: 'nonce', value: 'nonce' },
          ],
        },
        {
          key: 'value',
          label: '字段值',
          type: 'text',
          defaultValue: '1',
        },
      ],
    },
  ]
}
