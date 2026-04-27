/**
 * 页面头部组件
 * 包含标题、返回按钮、操作按钮等
 */
import { useNavigate } from 'react-router-dom'
import { Button, Space } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import type { PageHeaderProps } from '@/types/presentation'

export default function PageHeader({
  title,
  subtitle,
  showBack = false,
  backPath,
  extra,
  tags,
  children,
}: PageHeaderProps) {
  const navigate = useNavigate()

  // 处理返回
  const handleBack = () => {
    if (backPath) {
      navigate(backPath)
    } else {
      navigate(-1)
    }
  }

  return (
    <div className="mb-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0 flex-1">
          <div className="flex items-start">
          {showBack && (
            <Button
              type="text"
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
              className="mr-2 mt-1 shrink-0"
            />
          )}
            <div className="min-w-0">
              <div className="flex flex-col gap-2 md:flex-row md:items-center">
                <h1 className="m-0 text-title-lg text-text-primary">{title}</h1>
                {tags ? <div className="flex flex-wrap gap-2 md:ml-1">{tags}</div> : null}
              </div>
              {subtitle && (
                <p className="mt-1 mb-0 text-sm text-text-secondary">{subtitle}</p>
              )}
            </div>
          </div>
        </div>

        {extra ? <Space wrap className="justify-start lg:justify-end">{extra}</Space> : null}
      </div>

      {children && <div className="mt-4">{children}</div>}
    </div>
  )
}
