import { useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Button, Card, Divider, Form, Input, InputNumber, Select, Space, Tag } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import type {
  ExperimentCheckpointBlueprint,
  ExperimentCollaborationBlueprint,
  ExperimentContentBlueprintAsset,
  DockerImageCapability,
  ExperimentNodeBlueprint,
} from '@/types'
import type { ExperimentBlueprintEditorProps } from '@/types/presentation'
import { getImageCapabilities } from '@/api/admin'
import {
  capabilitySupportsAllTools,
  capabilitySupportsMode,
  describeCapability,
  getCapabilityImageRef,
} from '@/domains/runtime/imageCapabilities'
import {
  SERVICE_TEMPLATES,
  TOPOLOGY_TEMPLATES,
  WORKSPACE_TOOL_OPTIONS,
  describeImage,
  getImageLabel,
} from '@/domains/experiment/blueprint'

function splitScriptBlocks(value: string): string[] {
  return value
    .split('\n---\n')
    .map((item) => item.trim())
    .filter(Boolean)
}

function joinScriptBlocks(scripts?: string[]): string {
  return (scripts || []).join('\n---\n')
}

function updateAssetAt(
  assets: ExperimentContentBlueprintAsset[],
  index: number,
  patch: Partial<ExperimentContentBlueprintAsset>,
): ExperimentContentBlueprintAsset[] {
  return assets.map((asset, currentIndex) => currentIndex === index ? { ...asset, ...patch } : asset)
}

function updateCheckpointAt(
  checkpoints: ExperimentCheckpointBlueprint[],
  index: number,
  patch: Partial<ExperimentCheckpointBlueprint>,
): ExperimentCheckpointBlueprint[] {
  return checkpoints.map((checkpoint, currentIndex) => (
    currentIndex === index ? { ...checkpoint, ...patch } : checkpoint
  ))
}

function updateNodeAt(
  nodes: ExperimentNodeBlueprint[],
  index: number,
  patch: Partial<ExperimentNodeBlueprint>,
): ExperimentNodeBlueprint[] {
  return nodes.map((node, currentIndex) => currentIndex === index ? { ...node, ...patch } : node)
}

function updateRoleAt(
  collaboration: ExperimentCollaborationBlueprint,
  index: number,
  patch: Partial<NonNullable<ExperimentCollaborationBlueprint['roles']>[number]>,
): ExperimentCollaborationBlueprint {
  return {
    ...collaboration,
    roles: (collaboration.roles || []).map((role: NonNullable<ExperimentCollaborationBlueprint['roles']>[number], currentIndex: number) => (
      currentIndex === index ? { ...role, ...patch } : role
    )),
  }
}

export default function ExperimentBlueprintEditor({
  formData,
  images,
  selectedWorkspaceImage,
  selectedServiceKeys,
  selectedVisualizationModule,
  visualizationModuleOptions = [],
  infoMessage,
  infoDescription,
  showWorkspaceResources = false,
  onWorkspaceImageChange,
  onWorkspaceResourceChange,
  onInteractionToolsChange,
  onServicesChange,
  onTopologyTemplateChange,
  onVisualizationModuleChange,
  onWorkspaceInitScriptsChange,
  onContentInitScriptsChange,
  onContentAssetsChange,
  onNodesChange,
  onGradingStrategyChange,
  onCheckpointsChange,
  onCollaborationChange,
  onAssetUpload,
}: ExperimentBlueprintEditorProps) {
  const hasVisualizationService = selectedServiceKeys.includes('simulation')
  const contentAssets = formData.blueprint.content?.assets || []
  const nodes = formData.blueprint.nodes || []
  const grading = formData.blueprint.grading
  const collaboration = formData.blueprint.collaboration || { max_members: 2, roles: [] }
  const assetInputRef = useRef<HTMLInputElement | null>(null)
  const [uploadingAsset, setUploadingAsset] = useState(false)
  const [imageCapabilities, setImageCapabilities] = useState<DockerImageCapability[]>([])

  useEffect(() => {
    let disposed = false
    const load = async () => {
      try {
        const list = await getImageCapabilities()
        if (!disposed) {
          setImageCapabilities(list || [])
        }
      } catch {
        if (!disposed) {
          setImageCapabilities([])
        }
      }
    }
    void load()
    return () => {
      disposed = true
    }
  }, [])

  const modeCompatibility = useMemo(() => {
    switch (formData.blueprint.mode) {
      case 'multi_node':
        return ['single_user_multi_node', 'multi_service_lab']
      case 'collaboration':
        return ['collaborative', 'single_user_multi_node']
      default:
        return ['single_user']
    }
  }, [formData.blueprint.mode])

  const capabilityWorkspaceOptions = useMemo(() => {
    const requiredTools = formData.blueprint.workspace.interaction_tools || []
    return imageCapabilities
      .filter((capability) => capabilitySupportsMode(capability, modeCompatibility))
      .filter((capability) => capabilitySupportsAllTools(capability, requiredTools))
      .map((capability) => ({
        label: getCapabilityImageRef(capability),
        value: getCapabilityImageRef(capability),
        title: describeCapability(capability),
      }))
  }, [formData.blueprint.workspace.interaction_tools, imageCapabilities, modeCompatibility])

  const handleAssetUpload = async (file?: File) => {
    if (!file || !onAssetUpload || !onContentAssetsChange) {
      return
    }

    setUploadingAsset(true)
    try {
      const asset = await onAssetUpload(file)
      if (!asset) {
        return
      }

      onContentAssetsChange([
        ...contentAssets,
        asset,
      ])
    } finally {
      setUploadingAsset(false)
      if (assetInputRef.current) {
        assetInputRef.current.value = ''
      }
    }
  }

  return (
    <div className="max-w-5xl">
      <Alert
        type="info"
        showIcon
        className="mb-4"
        message={infoMessage}
        description={infoDescription}
      />

      <Form layout="vertical">
        <Form.Item label="工作区镜像" required>
          <Select
            value={formData.blueprint.workspace.image}
            onChange={onWorkspaceImageChange}
            options={capabilityWorkspaceOptions.length > 0
              ? capabilityWorkspaceOptions
              : images.map((image) => ({
                label: `${getImageLabel(image)} | ${image.category}`,
                value: getImageLabel(image),
              }))}
            placeholder="请选择工作区镜像"
            showSearch
            optionFilterProp="label"
          />
        </Form.Item>
        {capabilityWorkspaceOptions.length === 0 ? (
          <Alert
            type="warning"
            showIcon
            className="mb-4"
            message="当前镜像能力与工具组合不匹配"
            description="请调整工具集合，或在镜像管理中启用支持这些工具能力的镜像。"
          />
        ) : null}

        {selectedWorkspaceImage && (
          <Card size="small" className="mb-4">
            <div className="font-medium">{getImageLabel(selectedWorkspaceImage)}</div>
            <div className="mt-2 text-sm text-text-secondary">{describeImage(selectedWorkspaceImage)}</div>
            <div className="mt-2 flex flex-wrap gap-2">
              {selectedWorkspaceImage.features?.map((feature) => (
                <Tag key={feature}>{feature}</Tag>
              ))}
            </div>
          </Card>
        )}

        {showWorkspaceResources && onWorkspaceResourceChange && (
          <Form.Item label="工作区资源">
            <Space wrap>
              <Input
                value={formData.blueprint.workspace.resources.cpu}
                addonBefore="CPU"
                onChange={(event) => onWorkspaceResourceChange('cpu', event.target.value)}
                style={{ width: 180 }}
              />
              <Input
                value={formData.blueprint.workspace.resources.memory}
                addonBefore="内存"
                onChange={(event) => onWorkspaceResourceChange('memory', event.target.value)}
                style={{ width: 180 }}
              />
              <Input
                value={formData.blueprint.workspace.resources.storage}
                addonBefore="存储"
                onChange={(event) => onWorkspaceResourceChange('storage', event.target.value)}
                style={{ width: 180 }}
              />
            </Space>
          </Form.Item>
        )}

        <Form.Item label="交互工具">
          <Select
            mode="multiple"
            value={formData.blueprint.workspace.interaction_tools}
            onChange={onInteractionToolsChange}
            options={WORKSPACE_TOOL_OPTIONS}
          />
        </Form.Item>

        <Form.Item label="服务组件">
          <Select
            mode="multiple"
            value={selectedServiceKeys}
            onChange={onServicesChange}
            options={SERVICE_TEMPLATES.map((item) => ({
              label: `${item.name} | ${item.purpose}`,
              value: item.key,
            }))}
          />
        </Form.Item>

        {hasVisualizationService && onVisualizationModuleChange && (
          <Form.Item label="可视化模块">
            <Select
              value={selectedVisualizationModule}
              onChange={onVisualizationModuleChange}
              options={visualizationModuleOptions}
              showSearch
              optionFilterProp="label"
              placeholder="请选择要暴露给学生的可视化模块"
            />
          </Form.Item>
        )}

        {(formData.blueprint.services || []).length > 0 && (
          <div className="mb-4 grid grid-cols-1 gap-3 md:grid-cols-2">
            {(formData.blueprint.services || []).map((service) => (
              <Card key={service.key} size="small">
                <div className="font-medium">{service.name}</div>
                <div className="mt-1 text-sm text-text-secondary">{service.purpose}</div>
                <div className="mt-2 text-xs">{service.image}</div>
              </Card>
            ))}
          </div>
        )}

        <Form.Item label="拓扑模板">
          <Select
            value={formData.blueprint.topology?.template}
            onChange={onTopologyTemplateChange}
            options={TOPOLOGY_TEMPLATES.map((item) => ({
              label: `${item.label} | ${item.description}`,
              value: item.key,
            }))}
          />
        </Form.Item>

        {(onWorkspaceInitScriptsChange || onContentInitScriptsChange) && <Divider>内容注入</Divider>}

        {onWorkspaceInitScriptsChange && (
          <Form.Item label="工作区初始化脚本">
            <Input.TextArea
              rows={6}
              value={joinScriptBlocks(formData.blueprint.workspace.init_scripts)}
              onChange={(event) => onWorkspaceInitScriptsChange(splitScriptBlocks(event.target.value))}
              placeholder={'使用 "\\n---\\n" 分隔多段脚本'}
            />
          </Form.Item>
        )}

        {onContentAssetsChange && (
          <Form.Item label="实验资源文件">
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <input
                  ref={assetInputRef}
                  type="file"
                  className="hidden"
                  onChange={(event) => void handleAssetUpload(event.target.files?.[0])}
                />
                <Button
                  icon={<UploadOutlined />}
                  loading={uploadingAsset}
                  onClick={() => assetInputRef.current?.click()}
                >
                  上传资源文件
                </Button>
                <Button
                  onClick={() => onContentAssetsChange([
                    ...contentAssets,
                    {
                      key: `asset-${contentAssets.length + 1}`,
                      name: '',
                      source_type: 'object_storage',
                      bucket: '',
                      object_path: '',
                      target: 'workspace',
                      mount_path: '/workspace/',
                      required: true,
                    },
                  ])}
                >
                  手动添加资源
                </Button>
              </div>

              {contentAssets.length > 0 ? contentAssets.map((asset, index) => (
                <Card key={`${asset.key}-${index}`} size="small">
                  <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                    <Input
                      value={asset.key}
                      addonBefore="资源键"
                      onChange={(event) => onContentAssetsChange(updateAssetAt(contentAssets, index, { key: event.target.value }))}
                    />
                    <Input
                      value={asset.name}
                      addonBefore="文件名"
                      onChange={(event) => onContentAssetsChange(updateAssetAt(contentAssets, index, { name: event.target.value }))}
                    />
                    <Input
                      value={asset.bucket}
                      addonBefore="Bucket"
                      onChange={(event) => onContentAssetsChange(updateAssetAt(contentAssets, index, { bucket: event.target.value }))}
                    />
                    <Input
                      value={asset.object_path}
                      addonBefore="对象路径"
                      onChange={(event) => onContentAssetsChange(updateAssetAt(contentAssets, index, { object_path: event.target.value }))}
                    />
                    <Input
                      value={asset.target}
                      addonBefore="Target"
                      placeholder="workspace | all_nodes | node:peer1 | service:geth"
                      onChange={(event) => onContentAssetsChange(updateAssetAt(contentAssets, index, { target: event.target.value }))}
                    />
                    <Input
                      value={asset.mount_path}
                      addonBefore="挂载路径"
                      onChange={(event) => onContentAssetsChange(updateAssetAt(contentAssets, index, { mount_path: event.target.value }))}
                    />
                    <Select
                      value={asset.required ? 'required' : 'optional'}
                      onChange={(value) => onContentAssetsChange(updateAssetAt(contentAssets, index, { required: value === 'required' }))}
                      options={[
                        { label: '必需', value: 'required' },
                        { label: '可选', value: 'optional' },
                      ]}
                    />
                  </div>
                  <div className="mt-3 flex justify-end">
                    <Button
                      danger
                      onClick={() => onContentAssetsChange(contentAssets.filter((_, currentIndex) => currentIndex !== index))}
                    >
                      删除资源
                    </Button>
                  </div>
                </Card>
              )) : (
                <div className="text-sm text-text-secondary">当前没有配置实验资源文件。</div>
              )}
            </div>
          </Form.Item>
        )}

        {onContentInitScriptsChange && (
          <Form.Item label="内容初始化脚本">
            <Input.TextArea
              rows={6}
              value={joinScriptBlocks(formData.blueprint.content?.init_scripts)}
              onChange={(event) => onContentInitScriptsChange(splitScriptBlocks(event.target.value))}
              placeholder={'使用 "\\n---\\n" 分隔多段脚本'}
            />
          </Form.Item>
        )}

        {onNodesChange && (
          <>
            <Divider>多节点配置</Divider>
            <Form.Item label="节点实例">
              <div className="space-y-3">
                <Button
                  onClick={() => onNodesChange([
                    ...nodes,
                    {
                      key: `node-${nodes.length + 1}`,
                      name: '',
                      image: formData.blueprint.workspace.image,
                      role: '',
                      ports: [],
                      resources: {
                        cpu: '500m',
                        memory: '512Mi',
                        storage: '1Gi',
                      },
                      student_facing: true,
                      interaction_tools: ['terminal', 'logs'],
                      init_scripts: [],
                    },
                  ])}
                >
                  新增节点
                </Button>

                {nodes.length > 0 ? nodes.map((node, index) => (
                  <Card key={`${node.key}-${index}`} size="small">
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                      <Input
                        value={node.key}
                        addonBefore="节点键"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, { key: event.target.value }))}
                      />
                      <Input
                        value={node.name}
                        addonBefore="节点名"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, { name: event.target.value }))}
                      />
                      <Input
                        value={node.role}
                        addonBefore="角色"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, { role: event.target.value }))}
                      />
                      <Input
                        value={node.image}
                        addonBefore="镜像"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, { image: event.target.value }))}
                      />
                      <Input
                        value={(node.ports || []).join(',')}
                        addonBefore="端口"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, {
                          ports: event.target.value.split(',').map((item) => Number(item.trim())).filter((item) => Number.isFinite(item)),
                        }))}
                      />
                      <Select
                        mode="multiple"
                        value={node.interaction_tools || []}
                        onChange={(value) => onNodesChange(updateNodeAt(nodes, index, { interaction_tools: value }))}
                        options={WORKSPACE_TOOL_OPTIONS}
                      />
                      <Input
                        value={node.resources?.cpu}
                        addonBefore="CPU"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, {
                          resources: {
                            cpu: event.target.value,
                            memory: node.resources?.memory || '512Mi',
                            storage: node.resources?.storage || '1Gi',
                          },
                        }))}
                      />
                      <Input
                        value={node.resources?.memory}
                        addonBefore="内存"
                        onChange={(event) => onNodesChange(updateNodeAt(nodes, index, {
                          resources: {
                            cpu: node.resources?.cpu || '500m',
                            memory: event.target.value,
                            storage: node.resources?.storage || '1Gi',
                          },
                        }))}
                      />
                    </div>
                    <Input.TextArea
                      className="mt-3"
                      rows={4}
                      value={joinScriptBlocks(node.init_scripts)}
                      placeholder={'节点初始化脚本，使用 "\\n---\\n" 分隔多段脚本'}
                      onChange={(event) => onNodesChange(updateNodeAt(nodes, index, {
                        init_scripts: splitScriptBlocks(event.target.value),
                      }))}
                    />
                    <div className="mt-3 flex justify-end">
                      <Button
                        danger
                        onClick={() => onNodesChange(nodes.filter((_, currentIndex) => currentIndex !== index))}
                      >
                        删除节点
                      </Button>
                    </div>
                  </Card>
                )) : (
                  <div className="text-sm text-text-secondary">当前没有配置节点实例。</div>
                )}
              </div>
            </Form.Item>
          </>
        )}

        {(onGradingStrategyChange || onCheckpointsChange) && <Divider>评测规则</Divider>}

        {onGradingStrategyChange && (
          <Form.Item label="评分策略">
            <Select
              value={grading?.strategy || 'checkpoint'}
              onChange={onGradingStrategyChange}
              options={[
                { label: '检查点评测', value: 'checkpoint' },
                { label: '手动评分', value: 'manual' },
              ]}
            />
          </Form.Item>
        )}

        {onCheckpointsChange && (
          <Form.Item label="检查点">
            <div className="space-y-3">
              <Button
                onClick={() => onCheckpointsChange([
                  ...(grading?.checkpoints || []),
                  {
                    key: `checkpoint-${(grading?.checkpoints || []).length + 1}`,
                    type: 'command_exec',
                    target: 'workspace',
                    command: '',
                    expected: '',
                    score: 10,
                  },
                ])}
              >
                新增检查点
              </Button>

              {(grading?.checkpoints || []).length > 0 ? (grading?.checkpoints || []).map((checkpoint, index) => (
                <Card key={`${checkpoint.key}-${index}`} size="small">
                  <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                    <Input
                      value={checkpoint.key}
                      addonBefore="键"
                      onChange={(event) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { key: event.target.value }))}
                    />
                    <Select
                      value={checkpoint.type}
                      onChange={(value) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { type: value }))}
                      options={[
                        { label: '命令执行', value: 'command_exec' },
                        { label: '文件存在', value: 'file_exists' },
                        { label: '文件内容', value: 'file_content' },
                        { label: '测试通过', value: 'test_pass' },
                        { label: '自定义脚本', value: 'custom_script' },
                        { label: '合约已部署', value: 'contract_deployed' },
                      ]}
                    />
                    <Input
                      value={checkpoint.target}
                      addonBefore="目标"
                      onChange={(event) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { target: event.target.value }))}
                    />
                    <InputNumber
                      className="w-full"
                      value={checkpoint.score}
                      addonBefore="分值"
                      min={0}
                      onChange={(value) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { score: Number(value || 0) }))}
                    />
                    <Input
                      value={checkpoint.path}
                      addonBefore="路径"
                      onChange={(event) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { path: event.target.value }))}
                    />
                    <Input
                      value={checkpoint.expected}
                      addonBefore="期望"
                      onChange={(event) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { expected: event.target.value }))}
                    />
                  </div>
                  <Input.TextArea
                    className="mt-3"
                    rows={2}
                    value={checkpoint.command}
                    placeholder="命令"
                    onChange={(event) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { command: event.target.value }))}
                  />
                  <Input.TextArea
                    className="mt-3"
                    rows={3}
                    value={checkpoint.script}
                    placeholder="自定义脚本"
                    onChange={(event) => onCheckpointsChange(updateCheckpointAt(grading?.checkpoints || [], index, { script: event.target.value }))}
                  />
                  <div className="mt-3 flex justify-end">
                    <Button
                      danger
                      onClick={() => onCheckpointsChange((grading?.checkpoints || []).filter((_, currentIndex) => currentIndex !== index))}
                    >
                      删除检查点
                    </Button>
                  </div>
                </Card>
              )) : (
                <div className="text-sm text-text-secondary">当前没有配置检查点。</div>
              )}
            </div>
          </Form.Item>
        )}

        {formData.type === 'collaboration' && onCollaborationChange && (
          <>
            <Divider>协作配置</Divider>
            <Form.Item label="最大成员数">
              <InputNumber
                min={2}
                max={10}
                value={collaboration.max_members || 2}
                onChange={(value) => onCollaborationChange({
                  ...collaboration,
                  max_members: Number(value || 2),
                })}
              />
            </Form.Item>

            <Form.Item label="协作角色">
              <div className="space-y-3">
                <Button
                  onClick={() => onCollaborationChange({
                    ...collaboration,
                    roles: [
                      ...(collaboration.roles || []),
                      {
                        key: `role-${(collaboration.roles || []).length + 1}`,
                        label: '',
                        node_keys: [],
                        tool_keys: [],
                      },
                    ],
                  })}
                >
                  新增角色
                </Button>

                {(collaboration.roles || []).length > 0 ? (collaboration.roles || []).map((role, index) => (
                  <Card key={`${role.key}-${index}`} size="small">
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                      <Input
                        value={role.key}
                        addonBefore="角色键"
                        onChange={(event) => onCollaborationChange(updateRoleAt(collaboration, index, { key: event.target.value }))}
                      />
                      <Input
                        value={role.label}
                        addonBefore="角色名"
                        onChange={(event) => onCollaborationChange(updateRoleAt(collaboration, index, { label: event.target.value }))}
                      />
                      <Input
                        value={(role.node_keys || []).join(',')}
                        addonBefore="节点"
                        onChange={(event) => onCollaborationChange(updateRoleAt(collaboration, index, {
                          node_keys: event.target.value.split(',').map((item) => item.trim()).filter(Boolean),
                        }))}
                      />
                      <Input
                        value={(role.tool_keys || []).join(',')}
                        addonBefore="工具"
                        onChange={(event) => onCollaborationChange(updateRoleAt(collaboration, index, {
                          tool_keys: event.target.value.split(',').map((item) => item.trim()).filter(Boolean),
                        }))}
                      />
                    </div>
                    <div className="mt-3 flex justify-end">
                      <Button
                        danger
                        onClick={() => onCollaborationChange({
                          ...collaboration,
                          roles: (collaboration.roles || []).filter((_, currentIndex) => currentIndex !== index),
                        })}
                      >
                        删除角色
                      </Button>
                    </div>
                  </Card>
                )) : (
                  <div className="text-sm text-text-secondary">当前没有配置协作角色。</div>
                )}
              </div>
            </Form.Item>
          </>
        )}
      </Form>
    </div>
  )
}
