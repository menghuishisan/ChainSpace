import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Alert, Button, Card, Descriptions, Spin, Tag, message } from 'antd'
import { PlayCircleOutlined, RadarChartOutlined } from '@ant-design/icons'
import DOMPurify from 'dompurify'
import { marked } from 'marked'

import { getExperiment } from '@/api/experiment'
import { getSubmissions } from '@/api/experimentSubmission'
import { PageHeader, StatusTag, SubmissionStatusConfig } from '@/components/common'
import type { Experiment, Submission } from '@/types'
import { ExperimentTypeMap } from '@/types'
import { formatDurationCN } from '@/utils/format'

export default function StudentExperimentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const experimentId = Number.parseInt(id || '0', 10)

  const [loading, setLoading] = useState(true)
  const [experiment, setExperiment] = useState<Experiment | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [starting, setStarting] = useState(false)

  const fetchData = useCallback(async () => {
    if (!experimentId) {
      return
    }

    setLoading(true)
    try {
      const [experimentData, submissionResult] = await Promise.all([
        getExperiment(experimentId),
        getSubmissions({ experiment_id: experimentId, page: 1, page_size: 1 }).catch(() => ({
          list: [],
          total: 0,
          page: 1,
          page_size: 1,
        })),
      ])

      setExperiment(experimentData)
      setSubmission(submissionResult.list[0] || null)
    } catch {
      message.error('获取实验信息失败')
      navigate(-1)
    } finally {
      setLoading(false)
    }
  }, [experimentId, navigate])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const handleEnterWorkbench = async () => {
    setStarting(true)
    try {
      message.success('正在进入实验工作台')
      navigate(`/experiment/${experimentId}/env`)
    } finally {
      setStarting(false)
    }
  }

  if (loading) {
    return <div className="flex h-64 items-center justify-center"><Spin size="large" /></div>
  }

  if (!experiment) {
    return null
  }

  const renderedDescription = experiment.description
    ? DOMPurify.sanitize(marked(experiment.description) as string)
    : ''
  const assetCount = experiment.blueprint.content?.assets?.length || 0
  const checkpointCount = experiment.blueprint.grading?.checkpoints?.length || 0
  const nodeCount = experiment.blueprint.nodes?.length || 0
  const serviceCount = experiment.blueprint.services?.length || 0
  const toolCount = experiment.blueprint.tools?.length || experiment.blueprint.workspace.interaction_tools?.length || 0

  return (
    <div className="space-y-6">
      <PageHeader
        title={experiment.title}
        subtitle="查看实验说明、提交情况，并进入实验环境开始操作"
        showBack
        backPath="/student/experiments"
        tags={<Tag color="blue">{ExperimentTypeMap[experiment.type]}</Tag>}
      />

      <Card className="overflow-hidden border-0 shadow-sm" styles={{ body: { padding: 0 } }}>
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[minmax(0,1.45fr)_340px]">
          <div className="bg-[linear-gradient(135deg,#0b2239_0%,#0f2744_55%,#113d67_100%)] px-6 py-6 text-white">
            <div className="text-xs uppercase tracking-[0.28em] text-cyan-200">实验概览</div>
            <div className="mt-3 text-3xl font-semibold">{experiment.title}</div>
            <p className="mt-4 max-w-3xl text-sm leading-7 text-slate-200">
              {experiment.description || '暂无实验说明'}
            </p>
            <div className="mt-6 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">类型</div>
                <div className="mt-2 text-sm font-medium">{ExperimentTypeMap[experiment.type]}</div>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">工具</div>
                <div className="mt-2 text-sm font-medium">{toolCount} 项可用工具</div>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">时长</div>
                <div className="mt-2 text-sm font-medium">{formatDurationCN(experiment.estimated_time)}</div>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-cyan-200">分值</div>
                <div className="mt-2 text-sm font-medium">{experiment.max_score} 分</div>
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-4 bg-white px-6 py-6">
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-5">
              <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">快速开始</div>
              <div className="mt-3 text-xl font-semibold text-slate-900">进入实验环境</div>
              <p className="mt-2 text-sm leading-6 text-text-secondary">
                进入后即可使用本实验提供的编辑、终端、文件和其他操作工具完成任务。
              </p>
              <div className="mt-4">
                <Button
                  type="primary"
                  block
                  size="large"
                  icon={<PlayCircleOutlined />}
                  onClick={() => void handleEnterWorkbench()}
                  loading={starting}
                >
                  进入实验环境
                </Button>
              </div>
            </div>

            <Alert
              type="info"
              showIcon
              icon={<RadarChartOutlined />}
              message="进入后会自动准备实验内容"
              description={assetCount > 0
                ? `进入环境后会自动准备 ${assetCount} 项实验资料，方便你直接开始操作。`
                : '进入环境后可直接开始实验。'}
            />
          </div>
        </div>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
        <div className="space-y-6">
          <Card title="实验说明" className="border-0 shadow-sm">
            {renderedDescription
              ? <div className="prose max-w-none" dangerouslySetInnerHTML={{ __html: renderedDescription }} />
              : <div className="text-text-secondary">暂无实验说明</div>}
          </Card>

          <Card
            title="实验信息"
            className="border-0 shadow-sm"
            styles={{ header: { background: 'linear-gradient(90deg, rgba(24,144,255,0.08), rgba(114,46,209,0.08))' } }}
          >
            <Descriptions column={{ xs: 1, md: 2 }} size="small">
              <Descriptions.Item label="实验类型">{ExperimentTypeMap[experiment.type]}</Descriptions.Item>
              <Descriptions.Item label="预计时长">{formatDurationCN(experiment.estimated_time)}</Descriptions.Item>
              <Descriptions.Item label="满分">{experiment.max_score} 分</Descriptions.Item>
              <Descriptions.Item label="评分方式">{experiment.auto_grade ? '自动评测' : '人工批改'}</Descriptions.Item>
              <Descriptions.Item label="可用工具">{toolCount} 项</Descriptions.Item>
              <Descriptions.Item label="实验资源">{assetCount} 项</Descriptions.Item>
              <Descriptions.Item label="检查点">{checkpointCount} 项</Descriptions.Item>
              {(nodeCount > 0 || serviceCount > 0) ? (
                <Descriptions.Item label="实验环境">
                  {nodeCount > 0 ? `${nodeCount} 个节点` : '基础环境'}
                  {serviceCount > 0 ? ` / ${serviceCount} 个服务` : ''}
                </Descriptions.Item>
              ) : null}
            </Descriptions>
          </Card>
        </div>

        <div className="space-y-6 xl:sticky xl:top-6 xl:h-fit">
          {submission && (
            <Card
              title="提交状态"
              className="border-0 shadow-sm"
              styles={{ header: { background: 'linear-gradient(90deg, rgba(82,196,26,0.08), rgba(24,144,255,0.08))' } }}
            >
              <Descriptions column={1} size="small">
                <Descriptions.Item label="状态">
                  <StatusTag status={submission.status} statusMap={SubmissionStatusConfig} />
                </Descriptions.Item>
                {submission.score !== undefined && submission.score !== null && (
                  <Descriptions.Item label="得分">
                    <span className={submission.score >= 60 ? 'text-success' : 'text-error'}>
                      {submission.score} 分
                    </span>
                  </Descriptions.Item>
                )}
                {submission.feedback && <Descriptions.Item label="评语">{submission.feedback}</Descriptions.Item>}
              </Descriptions>
            </Card>
          )}

          <Card
            title="实验环境"
            className="border-0 shadow-sm"
            styles={{ header: { background: 'linear-gradient(90deg, rgba(250,173,20,0.08), rgba(24,144,255,0.08))' } }}
          >
            <Alert
              message="进入后将自动准备实验环境"
              description="系统会为你加载当前实验所需的工具与内容，你可以直接开始操作。"
              type="info"
              className="mb-4"
            />
            <Button
              type="primary"
              block
              icon={<PlayCircleOutlined />}
              onClick={() => void handleEnterWorkbench()}
              loading={starting}
            >
              进入实验环境
            </Button>
          </Card>
        </div>
      </div>
    </div>
  )
}
