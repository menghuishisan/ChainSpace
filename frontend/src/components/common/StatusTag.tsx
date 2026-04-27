/**
 * 状态标签组件。
 */
import { Tag } from 'antd'
import type { StatusConfig, StatusTagProps } from '@/types/presentation'

export default function StatusTag({ status, statusMap, className }: StatusTagProps) {
  const config = statusMap[status]

  if (!config) {
    return <Tag className={className}>{status}</Tag>
  }

  return (
    <Tag color={config.color} className={className}>
      {config.text}
    </Tag>
  )
}

// 预定义状态配置。
export const CourseStatusConfig: Record<string, StatusConfig> = {
  draft: { text: '草稿', color: 'default' },
  published: { text: '已发布', color: 'success' },
  archived: { text: '已归档', color: 'warning' },
}

export const ExperimentStatusConfig: Record<string, StatusConfig> = {
  draft: { text: '草稿', color: 'default' },
  published: { text: '已发布', color: 'success' },
}

export const EnvStatusConfig: Record<string, StatusConfig> = {
  creating: { text: '创建中', color: 'processing' },
  running: { text: '运行中', color: 'success' },
  paused: { text: '已暂停', color: 'warning' },
  terminated: { text: '已终止', color: 'default' },
}

export const SubmissionStatusConfig: Record<string, StatusConfig> = {
  submitted: { text: '待批改', color: 'processing' },
  graded: { text: '已批改', color: 'success' },
}

export const ContestStatusConfig: Record<string, StatusConfig> = {
  draft: { text: '草稿', color: 'default' },
  published: { text: '已发布', color: 'processing' },
  ongoing: { text: '进行中', color: 'success' },
  ended: { text: '已结束', color: 'default' },
}

export const UserStatusConfig: Record<string, StatusConfig> = {
  active: { text: '正常', color: 'success' },
  disabled: { text: '已禁用', color: 'error' },
}
