/**
 * 空状态组件
 */
import { Empty, Button } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import type { EmptyStateProps } from '@/types/presentation'

export default function EmptyState({
  image = 'default',
  description = '暂无数据',
  showCreate = false,
  createText = '立即创建',
  onCreate,
  extra,
}: EmptyStateProps) {
  // 获取图片
  const getImage = () => {
    switch (image) {
      case 'simple':
        return Empty.PRESENTED_IMAGE_SIMPLE
      default:
        return Empty.PRESENTED_IMAGE_DEFAULT
    }
  }

  return (
    <Empty
      image={getImage()}
      description={<span className="text-text-secondary">{description}</span>}
    >
      {showCreate && onCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>
          {createText}
        </Button>
      )}
      {extra}
    </Empty>
  )
}
