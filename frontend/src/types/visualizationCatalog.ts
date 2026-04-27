import type { VisualizationModuleOption } from './experimentBlueprint'
import type { VisualizationActionDefinition } from './visualizationAction'
import type { VisualizationRuntimeSpec } from './visualization'

/**
 * 前端可视化目录条目。
 * 统一描述教师端可选模块、运行时渲染配置以及对应动作入口。
 */
export interface VisualizationCatalogEntry {
  option: VisualizationModuleOption
  runtime: Omit<VisualizationRuntimeSpec, 'moduleKey'>
  getActions?: () => VisualizationActionDefinition[]
}
