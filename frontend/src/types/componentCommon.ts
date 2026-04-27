import type { ReactNode } from 'react'

export interface EmptyStateProps {
  image?: 'default' | 'simple' | 'search'
  description?: ReactNode
  showCreate?: boolean
  createText?: string
  onCreate?: () => void
  extra?: ReactNode
}

export interface PageHeaderProps {
  title: string
  subtitle?: string
  showBack?: boolean
  backPath?: string
  extra?: ReactNode
  tags?: ReactNode
  children?: ReactNode
}

export interface SearchFilterItem {
  key: string
  label?: string
  type: 'input' | 'select'
  placeholder?: string
  options?: Array<{ label: string; value: string | number }>
  width?: number
}

export interface SearchFilterProps {
  filters: SearchFilterItem[]
  values: Record<string, unknown>
  onChange: (values: Record<string, unknown>) => void
  onSearch: () => void
  onReset?: () => void
  bordered?: boolean
  extra?: ReactNode
}

export interface StatusConfig {
  text: string
  color: string
}

export interface StatusTagProps {
  status: string
  statusMap: Record<string, StatusConfig>
  className?: string
}
