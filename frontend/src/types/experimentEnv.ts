import type { ReactNode } from 'react'
import type { ExperimentToolKey } from './experimentBlueprint'

/**
 * 实验环境页左侧工具栏页签定义。
 * 这里统一描述工具键、图标和标题，避免页面组件内再声明结构类型。
 */
export interface ExperimentEnvTabDefinition {
  key: ExperimentToolKey
  icon: ReactNode
  title: string
}
