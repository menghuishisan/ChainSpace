import type { VisualizationActionDefinition } from '@/types/visualizationDomain'

/**
 * 哈希算法可视化动作定义。
 * 用于覆盖单次计算、算法对比、雪崩效应和完整性校验四条教学主链。
 */
export function getHashActions(): VisualizationActionDefinition[] {
  return [
    {
      key: 'hash-compute',
      label: '计算哈希',
      description: '对输入文本执行一次哈希计算，观察摘要输出如何生成。',
      kind: 'module_action',
      action: 'compute_hash',
      successMessage: '已完成一次哈希计算。',
      fields: [
        {
          key: 'algorithm',
          label: '算法',
          type: 'select',
          defaultValue: 'sha256',
          options: [
            { label: 'MD5', value: 'md5' },
            { label: 'SHA-1', value: 'sha1' },
            { label: 'SHA-256', value: 'sha256' },
            { label: 'SHA-512', value: 'sha512' },
          ],
        },
        {
          key: 'input',
          label: '输入文本',
          type: 'text',
          defaultValue: 'ChainSpace',
        },
      ],
    },
    {
      key: 'hash-compare',
      label: '算法对比',
      description: '对同一输入使用多种哈希算法，比较摘要长度和结果差异。',
      kind: 'module_action',
      action: 'compare_algorithms',
      successMessage: '已完成多算法对比。',
      fields: [
        {
          key: 'input',
          label: '输入文本',
          type: 'text',
          defaultValue: 'ChainSpace',
        },
      ],
    },
    {
      key: 'hash-avalanche',
      label: '雪崩效应',
      description: '只改变很小的输入片段，观察输出摘要如何发生大范围变化。',
      kind: 'module_action',
      action: 'demo_avalanche',
      successMessage: '已演示雪崩效应。',
      fields: [
        {
          key: 'algorithm',
          label: '算法',
          type: 'select',
          defaultValue: 'sha256',
          options: [
            { label: 'MD5', value: 'md5' },
            { label: 'SHA-1', value: 'sha1' },
            { label: 'SHA-256', value: 'sha256' },
            { label: 'SHA-512', value: 'sha512' },
          ],
        },
        {
          key: 'input',
          label: '输入文本',
          type: 'text',
          defaultValue: 'ChainSpace',
        },
      ],
    },
    {
      key: 'hash-verify',
      label: '完整性校验',
      description: '使用期望摘要验证输入是否被篡改，观察校验通过和失败的差异。',
      kind: 'module_action',
      action: 'verify_integrity',
      successMessage: '已完成完整性校验。',
      fields: [
        {
          key: 'algorithm',
          label: '算法',
          type: 'select',
          defaultValue: 'sha256',
          options: [
            { label: 'MD5', value: 'md5' },
            { label: 'SHA-1', value: 'sha1' },
            { label: 'SHA-256', value: 'sha256' },
            { label: 'SHA-512', value: 'sha512' },
          ],
        },
        {
          key: 'input',
          label: '输入文本',
          type: 'text',
          defaultValue: 'ChainSpace',
        },
      ],
    },
    {
      key: 'hash-reset',
      label: '清空历史',
      description: '清空当前哈希历史、对比结果和雪崩效应记录，便于重新演示。',
      kind: 'module_action',
      action: 'clear_history',
      successMessage: '已清空哈希历史。',
    },
  ]
}
