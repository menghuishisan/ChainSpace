import type {
  DockerImage,
  Experiment,
  ExperimentBlueprint,
  ExperimentCheckpointBlueprint,
  ExperimentCollaborationBlueprint,
  ExperimentContentBlueprintAsset,
  ExperimentEditorFormState,
  ExperimentNodeBlueprint,
  ExperimentServiceBlueprint,
  ExperimentToolKey,
  ExperimentType,
} from '@/types'
import {
  DEFAULT_VISUALIZATION_MODULE_KEY,
  SERVICE_TEMPLATES,
  createDefaultBlueprint,
  createExperimentEditorFormState,
  getSelectedVisualizationModule,
  getImageLabel,
  normalizeExperimentBlueprintDraft,
  WORKSPACE_TOOL_OPTIONS,
} from '@/domains/experiment/blueprint'

export function buildExperimentEditorFormState(
  type: ExperimentType = 'code_dev',
  defaults?: Partial<ExperimentEditorFormState>,
): ExperimentEditorFormState {
  return {
    ...createExperimentEditorFormState(type),
    ...defaults,
  }
}

export function mapExperimentToEditorFormState(experiment: Experiment): ExperimentEditorFormState {
  return {
    course_id: experiment.course_id,
    chapter_id: experiment.chapter_id,
    title: experiment.title,
    description: experiment.description || '',
    estimated_time: experiment.estimated_time || 60,
    type: experiment.type,
    max_score: experiment.max_score,
    blueprint: normalizeExperimentBlueprintDraft(experiment.blueprint, experiment.type),
  }
}

export function normalizeEditorBlueprint(formData: ExperimentEditorFormState): ExperimentBlueprint {
  return normalizeExperimentBlueprintDraft(formData.blueprint, formData.type)
}

export function changeEditorExperimentType(
  formData: ExperimentEditorFormState,
  type: ExperimentType,
): ExperimentEditorFormState {
  return {
    ...formData,
    type,
    blueprint: createDefaultBlueprint(type),
  }
}

export function updateEditorBlueprint(
  formData: ExperimentEditorFormState,
  updater: (blueprint: ExperimentBlueprint) => ExperimentBlueprint,
): ExperimentEditorFormState {
  return {
    ...formData,
    blueprint: normalizeExperimentBlueprintDraft(updater(formData.blueprint), formData.type),
  }
}

export function setEditorWorkspaceImage(
  formData: ExperimentEditorFormState,
  image: string,
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    workspace: {
      ...blueprint.workspace,
      image,
    },
  }))
}

export function setEditorWorkspaceResource(
  formData: ExperimentEditorFormState,
  key: 'cpu' | 'memory' | 'storage',
  value: string,
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    workspace: {
      ...blueprint.workspace,
      resources: {
        ...blueprint.workspace.resources,
        [key]: value,
      },
    },
  }))
}

export function setEditorInteractionTools(
  formData: ExperimentEditorFormState,
  tools: string[],
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    workspace: {
      ...blueprint.workspace,
      interaction_tools: tools,
    },
  }))
}

export function setEditorTopologyTemplate(
  formData: ExperimentEditorFormState,
  template: string,
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    topology: {
      ...(blueprint.topology || {}),
      template,
    },
  }))
}

export function setEditorServices(
  formData: ExperimentEditorFormState,
  keys: string[],
): ExperimentEditorFormState {
  const services: ExperimentServiceBlueprint[] = SERVICE_TEMPLATES
    .filter((item) => keys.includes(item.key))
    .map((item) => ({ ...item }))

  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    services,
  }))
}

export function setEditorVisualizationModule(
  formData: ExperimentEditorFormState,
  moduleKey: string = DEFAULT_VISUALIZATION_MODULE_KEY,
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    tools: (blueprint.tools || []).map((tool) => (
      tool.key === 'visualization'
        ? { ...tool, kind: moduleKey }
        : tool
    )),
  }))
}

export function setEditorWorkspaceInitScripts(
  formData: ExperimentEditorFormState,
  scripts: string[],
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    workspace: {
      ...blueprint.workspace,
      init_scripts: scripts,
    },
  }))
}

export function setEditorContentInitScripts(
  formData: ExperimentEditorFormState,
  scripts: string[],
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    content: {
      ...(blueprint.content || {}),
      assets: blueprint.content?.assets || [],
      init_scripts: scripts,
    },
  }))
}

export function setEditorContentAssets(
  formData: ExperimentEditorFormState,
  assets: ExperimentContentBlueprintAsset[],
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    content: {
      ...(blueprint.content || {}),
      assets,
      init_scripts: blueprint.content?.init_scripts || [],
    },
  }))
}

export function setEditorNodes(
  formData: ExperimentEditorFormState,
  nodes: ExperimentNodeBlueprint[],
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    nodes,
  }))
}

export function setEditorGradingStrategy(
  formData: ExperimentEditorFormState,
  strategy: string,
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    grading: {
      ...(blueprint.grading || {}),
      strategy,
      checkpoints: blueprint.grading?.checkpoints || [],
    },
  }))
}

export function setEditorCheckpoints(
  formData: ExperimentEditorFormState,
  checkpoints: ExperimentCheckpointBlueprint[],
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    grading: {
      ...(blueprint.grading || {}),
      strategy: blueprint.grading?.strategy || 'checkpoint',
      checkpoints,
    },
  }))
}

export function setEditorCollaborationConfig(
  formData: ExperimentEditorFormState,
  collaboration: ExperimentCollaborationBlueprint,
): ExperimentEditorFormState {
  return updateEditorBlueprint(formData, (blueprint) => ({
    ...blueprint,
    collaboration,
  }))
}

export function getEditorSelectedServiceKeys(formData: ExperimentEditorFormState): string[] {
  return (formData.blueprint.services || []).map((service) => service.key)
}

export function getEditorWorkspaceToolValues(formData: ExperimentEditorFormState): ExperimentToolKey[] {
  return (formData.blueprint.workspace.interaction_tools || []).filter((tool): tool is ExperimentToolKey =>
    WORKSPACE_TOOL_OPTIONS.some((option) => option.value === tool as ExperimentToolKey),
  )
}

export function findSelectedWorkspaceImage(
  images: DockerImage[],
  formData: ExperimentEditorFormState | null | undefined,
): DockerImage | undefined {
  if (!formData) {
    return undefined
  }

  return images.find((image) => getImageLabel(image) === formData.blueprint.workspace.image)
}

export { getSelectedVisualizationModule }
