/**
 * 搜索筛选组件
 * 用于列表页的搜索和筛选
 */
import { Input, Select, Button, Space, Card } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { SearchFilterProps } from '@/types/presentation'

export default function SearchFilter({
  filters,
  values,
  onChange,
  onSearch,
  onReset,
  bordered = true,
  extra,
}: SearchFilterProps) {
  // 处理单个字段变化
  const handleChange = (key: string, value: unknown) => {
    onChange({ ...values, [key]: value })
  }

  // 处理键盘回车
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      onSearch()
    }
  }

  // 处理重置
  const handleReset = () => {
    const emptyValues: Record<string, unknown> = {}
    filters.forEach(f => { emptyValues[f.key] = '' })
    onChange(emptyValues)
    onReset?.()
  }

  const content = (
    <div className="flex flex-wrap items-center gap-4">
      {filters.map((filter) => (
        <div key={filter.key} className="flex items-center">
          {filter.type === 'input' ? (
            <Input
              placeholder={filter.placeholder || '搜索'}
              value={values[filter.key] as string}
              onChange={(e) => handleChange(filter.key, e.target.value)}
              onKeyDown={handleKeyDown}
              style={{ width: filter.width || 180 }}
              prefix={<SearchOutlined className="text-text-secondary" />}
              allowClear
            />
          ) : (
            <Select
              placeholder={filter.placeholder || '全部'}
              value={values[filter.key] || undefined}
              onChange={(value) => handleChange(filter.key, value)}
              options={filter.options}
              style={{ width: filter.width || 140 }}
              allowClear
            />
          )}
        </div>
      ))}

      <Space>
        <Button type="primary" icon={<SearchOutlined />} onClick={onSearch}>
          搜索
        </Button>
        <Button icon={<ReloadOutlined />} onClick={handleReset}>
          重置
        </Button>
        {extra}
      </Space>
    </div>
  )

  if (bordered) {
    return (
      <Card className="mb-4" styles={{ body: { padding: '16px 24px' } }}>
        {content}
      </Card>
    )
  }

  return <div className="mb-4">{content}</div>
}
