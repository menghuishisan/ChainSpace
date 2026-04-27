import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, Select, Steps, message } from 'antd'
import { ArrowLeftOutlined, ArrowRightOutlined, PlusOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import { ExperimentBlueprintEditor } from '@/components/experiment'
import { getAllImages } from '@/api/admin'
import { getChapters, getCourses } from '@/api/course'
import { createExperiment } from '@/api/experiment'
import { uploadExperimentAsset } from '@/api/upload'
import { usePersistedStep } from '@/hooks'
import type { Chapter, Course, DockerImage, ExperimentEditorFormState, ExperimentType } from '@/types'
import { ExperimentTypeMap } from '@/types'
import {
  buildExperimentEditorFormState,
  changeEditorExperimentType,
  findSelectedWorkspaceImage,
  getEditorSelectedServiceKeys,
  getSelectedVisualizationModule,
  normalizeEditorBlueprint,
  setEditorCheckpoints,
  setEditorCollaborationConfig,
  setEditorContentAssets,
  setEditorContentInitScripts,
  setEditorGradingStrategy,
  setEditorInteractionTools,
  setEditorNodes,
  setEditorServices,
  setEditorTopologyTemplate,
  setEditorVisualizationModule,
  setEditorWorkspaceImage,
  setEditorWorkspaceInitScripts,
  setEditorWorkspaceResource,
} from '@/domains/experiment/editor'
import { VISUALIZATION_TOOL_OPTIONS } from '@/domains/experiment/blueprint'

const { Step } = Steps

function buildInitialState(type: ExperimentType = 'code_dev'): ExperimentEditorFormState {
  return buildExperimentEditorFormState(type)
}

export default function TeacherExperimentCreate() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const defaultCourseId = searchParams.get('course_id')
  const { currentStep, goNext, goPrev, resetStep } = usePersistedStep('experiment_create', 4)

  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [images, setImages] = useState<DockerImage[]>([])
  const [formData, setFormData] = useState<ExperimentEditorFormState>(() => {
    const initial = buildInitialState()
    if (defaultCourseId) {
      initial.course_id = Number.parseInt(defaultCourseId, 10)
    }
    return initial
  })

  useEffect(() => {
    const init = async () => {
      try {
        const [courseData, imageData] = await Promise.all([
          getCourses({ page: 1, page_size: 100 }),
          getAllImages(),
        ])
        setCourses(courseData.list)
        setImages(imageData)
      } catch {
        message.error('加载实验配置数据失败')
      }
    }

    void init()
  }, [])

  useEffect(() => {
    if (!formData.course_id) {
      setChapters([])
      return
    }

    getChapters(formData.course_id).then(setChapters).catch(() => setChapters([]))
  }, [formData.course_id])

  const selectedWorkspaceImage = useMemo(
    () => findSelectedWorkspaceImage(images, formData),
    [formData, images],
  )
  const selectedServiceKeys = getEditorSelectedServiceKeys(formData)
  const selectedVisualizationModule = useMemo(
    () => getSelectedVisualizationModule(formData),
    [formData],
  )

  const updateField = <K extends keyof ExperimentEditorFormState>(key: K, value: ExperimentEditorFormState[K]) => {
    setFormData((prev) => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!formData.chapter_id || !formData.title.trim()) {
      message.error('请先填写完整的实验基本信息')
      return
    }

    const blueprint = normalizeEditorBlueprint(formData)
    const autoGrade = (blueprint.grading?.strategy || 'checkpoint') !== 'manual'

    setLoading(true)
    try {
      await createExperiment({
        chapter_id: formData.chapter_id,
        title: formData.title,
        description: formData.description,
        type: formData.type,
        estimated_time: formData.estimated_time,
        max_score: formData.max_score,
        auto_grade: autoGrade,
        blueprint,
      })
      message.success('实验创建成功')
      resetStep()
      navigate(`/teacher/courses/${formData.course_id}`)
    } finally {
      setLoading(false)
    }
  }

  const steps = [
    {
      title: '基本信息',
      content: (
        <div className="max-w-3xl">
          <Form layout="vertical">
            <Form.Item label="所属课程" required>
              <Select
                value={formData.course_id}
                onChange={(value) => updateField('course_id', value)}
                placeholder="请选择课程"
                options={courses.map((course) => ({ label: course.title, value: course.id }))}
                notFoundContent={<Empty description="暂无课程，请先创建课程" />}
              />
            </Form.Item>
            <Form.Item label="所属章节" required>
              <Select
                value={formData.chapter_id}
                onChange={(value) => updateField('chapter_id', value)}
                placeholder="请选择章节"
                options={chapters.map((chapter) => ({ label: chapter.title, value: chapter.id }))}
                notFoundContent={<Empty description="当前课程暂无章节" />}
              />
            </Form.Item>
            <Form.Item label="实验名称" required>
              <Input value={formData.title} onChange={(event) => updateField('title', event.target.value)} />
            </Form.Item>
            <Form.Item label="实验说明">
              <Input.TextArea
                value={formData.description}
                rows={4}
                onChange={(event) => updateField('description', event.target.value)}
              />
            </Form.Item>
            <Form.Item label="预计时长（分钟）">
              <InputNumber
                min={10}
                max={480}
                value={formData.estimated_time}
                onChange={(value) => updateField('estimated_time', Number(value || 60))}
              />
            </Form.Item>
          </Form>
        </div>
      ),
    },
    {
      title: '实验类型',
      content: (
        <div>
          <p className="mb-4 text-text-secondary">
            实验类型决定默认工具和环境配置，后续仍可继续细化节点、内容和评测规则。
          </p>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            {(Object.entries(ExperimentTypeMap) as [ExperimentType, string][]).map(([key, label]) => (
              <Card
                key={key}
                hoverable
                className={formData.type === key ? 'border-primary border-2' : ''}
                onClick={() => setFormData((prev) => changeEditorExperimentType(prev, key))}
              >
                <div className="mb-2 font-medium">{label}</div>
                <div className="text-sm text-text-secondary">
                  {key === 'visualization'
                    ? '适合通过可视化方式展示实验过程与状态变化。'
                    : key === 'collaboration'
                      ? '支持会话成员、协作角色与节点分工的实验模式。'
                      : '可配置工作区、节点、服务、内容和评分规则。'}
                </div>
              </Card>
            ))}
          </div>
        </div>
      ),
    },
    {
      title: '环境配置',
      content: (
        <ExperimentBlueprintEditor
          formData={formData}
          images={images}
          selectedWorkspaceImage={selectedWorkspaceImage}
          selectedServiceKeys={selectedServiceKeys}
          selectedVisualizationModule={selectedVisualizationModule}
          visualizationModuleOptions={VISUALIZATION_TOOL_OPTIONS}
          infoMessage="实验环境配置"
          infoDescription="在这里配置工作区、节点、服务、内容、评测规则和协作角色，学生进入实验后会直接使用这些内容。"
          showWorkspaceResources
          onWorkspaceImageChange={(image) => setFormData((prev) => setEditorWorkspaceImage(prev, image))}
          onWorkspaceResourceChange={(key, value) => setFormData((prev) => setEditorWorkspaceResource(prev, key, value))}
          onInteractionToolsChange={(tools) => setFormData((prev) => setEditorInteractionTools(prev, tools))}
          onServicesChange={(keys) => setFormData((prev) => setEditorServices(prev, keys))}
          onTopologyTemplateChange={(template) => setFormData((prev) => setEditorTopologyTemplate(prev, template))}
          onVisualizationModuleChange={(moduleKey) => setFormData((prev) => setEditorVisualizationModule(prev, moduleKey))}
          onWorkspaceInitScriptsChange={(scripts) => setFormData((prev) => setEditorWorkspaceInitScripts(prev, scripts))}
          onContentInitScriptsChange={(scripts) => setFormData((prev) => setEditorContentInitScripts(prev, scripts))}
          onContentAssetsChange={(assets) => setFormData((prev) => setEditorContentAssets(prev, assets))}
          onNodesChange={(nodes) => setFormData((prev) => setEditorNodes(prev, nodes))}
          onGradingStrategyChange={(strategy) => setFormData((prev) => setEditorGradingStrategy(prev, strategy))}
          onCheckpointsChange={(checkpoints) => setFormData((prev) => setEditorCheckpoints(prev, checkpoints))}
          onCollaborationChange={(collaboration) => setFormData((prev) => setEditorCollaborationConfig(prev, collaboration))}
          onAssetUpload={async (file) => {
            const result = await uploadExperimentAsset(file)
            return {
              key: result.filename,
              name: result.filename,
              source_type: 'object_storage',
              bucket: result.bucket,
              object_path: result.path,
              mount_path: `/workspace/${result.filename}`,
              required: true,
            }
          }}
        />
      ),
    },
    {
      title: '发布设置',
      content: (
        <div className="max-w-4xl">
          {formData.type === 'visualization' && (
            <Alert
              type="info"
              showIcon
              className="mb-4"
              message="可视化实验说明"
              description="当前实验会按所选模块展示可视化内容，学生进入后可直接查看和交互。"
            />
          )}

          <Form layout="vertical">
            <Form.Item label="满分">
              <InputNumber
                min={10}
                max={1000}
                value={formData.max_score}
                onChange={(value) => updateField('max_score', Number(value || 100))}
              />
            </Form.Item>
            <Form.Item label="评分方式">
              <Select
                value={formData.blueprint.grading?.strategy || 'checkpoint'}
                onChange={(value) => setFormData((prev) => setEditorGradingStrategy(prev, value))}
                options={[
                  { label: '检查点评测', value: 'checkpoint' },
                  { label: '手动评分', value: 'manual' },
                ]}
              />
            </Form.Item>
          </Form>
        </div>
      ),
    },
  ]

  return (
    <div>
      <PageHeader
        title="创建实验"
        subtitle="配置实验环境、内容和评测规则"
        extra={(
          <Button icon={<PlusOutlined />} type="primary" onClick={handleSubmit} loading={loading}>
            创建实验
          </Button>
        )}
      />

      <Card>
        <Steps current={currentStep} className="mb-8">
          {steps.map((step, index) => (
            <Step key={index} title={step.title} />
          ))}
        </Steps>

        {steps[currentStep].content}

        <div className="mt-8 flex justify-between border-t pt-4">
          <div>
            {currentStep > 0 && (
              <Button icon={<ArrowLeftOutlined />} onClick={goPrev}>
                上一步
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            {currentStep < steps.length - 1 ? (
              <Button type="primary" icon={<ArrowRightOutlined />} onClick={goNext}>
                下一步
              </Button>
            ) : (
              <Button type="primary" onClick={handleSubmit} loading={loading}>
                创建实验
              </Button>
            )}
          </div>
        </div>
      </Card>
    </div>
  )
}
