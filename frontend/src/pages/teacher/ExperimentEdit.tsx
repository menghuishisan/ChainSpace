import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, Select, Space, Spin, message } from 'antd'

import { getAllImages } from '@/api/admin'
import { getChapters, getCourses } from '@/api/course'
import { getExperiment, publishExperiment, updateExperiment } from '@/api/experiment'
import { uploadExperimentAsset } from '@/api/upload'
import { PageHeader } from '@/components/common'
import { ExperimentBlueprintEditor } from '@/components/experiment'
import {
  changeEditorExperimentType,
  findSelectedWorkspaceImage,
  getEditorSelectedServiceKeys,
  getSelectedVisualizationModule,
  mapExperimentToEditorFormState,
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
import type { Chapter, Course, DockerImage, Experiment, ExperimentEditorFormState, ExperimentType } from '@/types'
import { ExperimentTypeMap } from '@/types'

export default function TeacherExperimentEdit() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const experimentId = Number.parseInt(id || '0', 10)

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [courses, setCourses] = useState<Course[]>([])
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [images, setImages] = useState<DockerImage[]>([])
  const [experiment, setExperiment] = useState<Experiment | null>(null)
  const [formData, setFormData] = useState<ExperimentEditorFormState | null>(null)

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      try {
        const [experimentData, courseData, imageData] = await Promise.all([
          getExperiment(experimentId),
          getCourses({ page: 1, page_size: 100 }),
          getAllImages(),
        ])

        setExperiment(experimentData)
        setCourses(courseData.list)
        setImages(imageData)
        setFormData(mapExperimentToEditorFormState(experimentData))

        if (experimentData.course_id) {
          const chapterData = await getChapters(experimentData.course_id)
          setChapters(chapterData)
        }
      } catch {
        message.error('加载实验失败')
        navigate(-1)
      } finally {
        setLoading(false)
      }
    }

    void init()
  }, [experimentId, navigate])

  useEffect(() => {
    if (!formData?.course_id) {
      setChapters([])
      if (formData?.chapter_id) {
        setFormData((previous) => previous ? { ...previous, chapter_id: undefined } : previous)
      }
      return
    }

    getChapters(formData.course_id)
      .then((chapterList) => {
        setChapters(chapterList)
        if (!chapterList.some((chapter) => chapter.id === formData.chapter_id)) {
          setFormData((previous) => previous ? { ...previous, chapter_id: undefined } : previous)
        }
      })
      .catch(() => setChapters([]))
  }, [formData?.chapter_id, formData?.course_id])

  const selectedWorkspaceImage = useMemo(
    () => findSelectedWorkspaceImage(images, formData),
    [formData, images],
  )
  const selectedVisualizationModule = useMemo(
    () => formData ? getSelectedVisualizationModule(formData) : undefined,
    [formData],
  )

  if (loading || !formData) {
    return <div className="flex h-64 items-center justify-center"><Spin size="large" /></div>
  }

  const updateField = <K extends keyof ExperimentEditorFormState>(key: K, value: ExperimentEditorFormState[K]) => {
    setFormData((previous) => previous ? { ...previous, [key]: value } : previous)
  }

  const handleCourseChange = (courseId?: number) => {
    setFormData((previous) => {
      if (!previous) {
        return previous
      }
      return {
        ...previous,
        course_id: courseId,
        chapter_id: undefined,
      }
    })
  }

  const handleSave = async () => {
    if (!formData.chapter_id) {
      message.error('请选择所属章节')
      return
    }

    const blueprint = normalizeEditorBlueprint(formData)
    const autoGrade = (blueprint.grading?.strategy || 'checkpoint') !== 'manual'

    setSaving(true)
    try {
      await updateExperiment(experimentId, {
        chapter_id: formData.chapter_id,
        title: formData.title,
        description: formData.description,
        type: formData.type,
        estimated_time: formData.estimated_time,
        max_score: formData.max_score,
        auto_grade: autoGrade,
        blueprint,
      })
      message.success('实验已保存')
    } finally {
      setSaving(false)
    }
  }

  const handlePublish = async () => {
    await publishExperiment(experimentId)
    message.success('实验已发布')
    navigate('/teacher/experiments')
  }

  return (
    <div>
      <PageHeader
        title="编辑实验"
        subtitle={experiment?.title}
        showBack
        extra={(
          <Space>
            <Button onClick={() => void handleSave()} loading={saving}>保存</Button>
            {experiment?.status === 'draft' && (
              <Button type="primary" onClick={() => void handlePublish()}>发布实验</Button>
            )}
          </Space>
        )}
      />

      <Card className="mb-4">
        <Alert
          type="info"
          showIcon
          className="mb-4"
          message="归属关系说明"
          description="实验最终归属于章节。课程选择仅用于筛选可选章节，保存时会提交 chapter_id。"
        />

        <Form layout="vertical">
          <Form.Item label="所属课程">
            <Select
              value={formData.course_id}
              onChange={handleCourseChange}
              options={courses.map((course) => ({ label: course.title, value: course.id }))}
              notFoundContent={<Empty description="暂无课程" />}
            />
          </Form.Item>
          <Form.Item label="所属章节">
            <Select
              value={formData.chapter_id}
              onChange={(value) => updateField('chapter_id', value)}
              options={chapters.map((chapter) => ({ label: chapter.title, value: chapter.id }))}
              notFoundContent={<Empty description="暂无章节" />}
            />
          </Form.Item>
          <Form.Item label="实验名称">
            <Input value={formData.title} onChange={(event) => updateField('title', event.target.value)} />
          </Form.Item>
          <Form.Item label="实验说明">
            <Input.TextArea
              value={formData.description}
              rows={4}
              onChange={(event) => updateField('description', event.target.value)}
            />
          </Form.Item>
          <Form.Item label="实验类型">
            <Select
              value={formData.type}
              onChange={(value) => setFormData((previous) => (
                previous ? changeEditorExperimentType(previous, value as ExperimentType) : previous
              ))}
              options={(Object.entries(ExperimentTypeMap) as [ExperimentType, string][]).map(([value, label]) => ({ label, value }))}
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
              onChange={(value) => setFormData((previous) => previous ? setEditorGradingStrategy(previous, value) : previous)}
              options={[
                { label: '检查点评测', value: 'checkpoint' },
                { label: '手动评分', value: 'manual' },
              ]}
            />
          </Form.Item>
        </Form>
      </Card>

      <Card className="mb-4" title="环境配置">
        <ExperimentBlueprintEditor
          formData={formData}
          images={images}
          selectedWorkspaceImage={selectedWorkspaceImage}
          selectedServiceKeys={getEditorSelectedServiceKeys(formData)}
          selectedVisualizationModule={selectedVisualizationModule}
          visualizationModuleOptions={VISUALIZATION_TOOL_OPTIONS}
          infoMessage="实验环境配置"
          infoDescription="编辑页与创建页使用同一套配置方式，用于调整工作区、节点、服务、内容、评测规则和协作角色。"
          showWorkspaceResources
          onWorkspaceImageChange={(value) => setFormData((previous) => previous ? setEditorWorkspaceImage(previous, value) : previous)}
          onWorkspaceResourceChange={(key, value) => setFormData((previous) => previous ? setEditorWorkspaceResource(previous, key, value) : previous)}
          onInteractionToolsChange={(value) => setFormData((previous) => previous ? setEditorInteractionTools(previous, value) : previous)}
          onServicesChange={(keys) => setFormData((previous) => previous ? setEditorServices(previous, keys) : previous)}
          onTopologyTemplateChange={(value) => setFormData((previous) => previous ? setEditorTopologyTemplate(previous, value) : previous)}
          onVisualizationModuleChange={(moduleKey) => setFormData((previous) => previous ? setEditorVisualizationModule(previous, moduleKey) : previous)}
          onWorkspaceInitScriptsChange={(scripts) => setFormData((previous) => previous ? setEditorWorkspaceInitScripts(previous, scripts) : previous)}
          onContentInitScriptsChange={(scripts) => setFormData((previous) => previous ? setEditorContentInitScripts(previous, scripts) : previous)}
          onContentAssetsChange={(assets) => setFormData((previous) => previous ? setEditorContentAssets(previous, assets) : previous)}
          onNodesChange={(nodes) => setFormData((previous) => previous ? setEditorNodes(previous, nodes) : previous)}
          onGradingStrategyChange={(strategy) => setFormData((previous) => previous ? setEditorGradingStrategy(previous, strategy) : previous)}
          onCheckpointsChange={(checkpoints) => setFormData((previous) => previous ? setEditorCheckpoints(previous, checkpoints) : previous)}
          onCollaborationChange={(collaboration) => setFormData((previous) => previous ? setEditorCollaborationConfig(previous, collaboration) : previous)}
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

        {formData.type === 'visualization' && (
          <Alert
            type="info"
            showIcon
            message="可视化实验说明"
            description="当前实验会按所选模块展示可视化内容，学生进入后可直接查看和交互。"
          />
        )}
      </Card>
    </div>
  )
}
