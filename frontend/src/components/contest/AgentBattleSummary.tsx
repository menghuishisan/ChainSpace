import type { ChangeEvent } from 'react'
import {
  Alert,
  Button,
  Card,
  Descriptions,
  Input,
  Progress,
  Tabs,
  Tag,
  Upload,
} from 'antd'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CodeOutlined,
  EyeOutlined,
  RocketOutlined,
  UploadOutlined,
} from '@ant-design/icons'
import type { AgentBattleSummaryProps } from '@/types/presentation'
import { ROUND_STATUS_MAP, STRATEGY_TEMPLATE } from '@/domains/contest/battle'
import { BattleStatusMap } from '@/types'
import { formatDateTime, formatDuration } from '@/utils/format'

const { TextArea } = Input

export default function AgentBattleSummary({
  contest,
  battleStatus,
  battleConfig,
  currentRound,
  contractInfo,
  sourceCode,
  deploying,
  uploading,
  remainingTime,
  onSourceCodeChange,
  onDeploy,
  onUploadFile,
  onSpectate,
}: AgentBattleSummaryProps) {
  const scoreWeights = battleConfig.judge.score_weights || {}
  const allowedActions = battleConfig.judge.allowed_actions || []
  const isTimeLow = remainingTime <= 600
  const contestStatusLabel = (battleStatus?.status && BattleStatusMap[battleStatus.status])
    ? BattleStatusMap[battleStatus.status].text
    : '未开始'
  const contractStatusLabel = (() => {
    switch (contractInfo?.status) {
      case 'active':
      case 'deployed':
        return '已生效'
      case 'pending':
      case 'deploying':
      case 'upgrading':
        return '部署中'
      case 'failed':
      case 'upgrade_failed':
        return '部署失败'
      default:
        return contractInfo?.status || '未部署'
    }
  })()

  return (
    <div className="space-y-6">
      <Card
        className="overflow-hidden border-0 shadow-sm"
        styles={{ body: { padding: 0 } }}
      >
        <div className="grid gap-px bg-slate-200 xl:grid-cols-[minmax(0,1.35fr)_240px]">
          <div className="bg-[linear-gradient(135deg,#eef6ff_0%,#f5f9ff_55%,#eefaf7_100%)] px-6 py-6 text-slate-900">
            <div className="text-xs uppercase tracking-[0.28em] text-sky-600">对局总览</div>
            <div className="mt-3 flex flex-wrap items-center gap-3">
              <div className="text-2xl font-semibold">{contest.title}</div>
              <Tag color={battleStatus?.status === 'running' ? 'success' : 'default'}>
                {contestStatusLabel}
              </Tag>
            </div>
            <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">轮次</div>
                <div className="mt-2 text-lg font-semibold">
                  {battleStatus ? `${battleStatus.current_round} / ${battleStatus.total_rounds}` : '-'}
                </div>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">排名</div>
                <div className="mt-2 text-lg font-semibold">{battleStatus?.my_rank || '-'}</div>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">得分</div>
                <div className="mt-2 text-lg font-semibold">{battleStatus?.my_score || 0}</div>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-slate-500">区块高度</div>
                <div className="mt-2 text-lg font-semibold">{battleStatus?.current_block || 0}</div>
              </div>
            </div>
            {battleStatus?.status === 'running' ? (
              <div className="mt-5">
                <Progress
                  percent={Math.round(((battleStatus.current_round || 0) / (battleStatus.total_rounds || 1)) * 100)}
                  status="active"
                />
              </div>
            ) : null}
          </div>

          <div className="flex flex-col justify-between gap-4 bg-white px-6 py-6">
            <div>
              <div className="text-xs uppercase tracking-[0.18em] text-text-secondary">剩余时间</div>
              <div className={`mt-3 flex items-center text-2xl font-semibold ${isTimeLow ? 'text-error' : 'text-slate-900'}`}>
                <ClockCircleOutlined className="mr-2" />
                {formatDuration(remainingTime)}
              </div>
              <div className="mt-3 space-y-2 text-sm text-text-secondary">
                <div>轮次状态：{currentRound ? ROUND_STATUS_MAP[currentRound.status]?.text || '待开始' : '待开始'}</div>
                <div>当前策略：{contractInfo ? `v${contractInfo.version}（${contractStatusLabel}）` : '尚未部署'}</div>
              </div>
            </div>
            <Button icon={<EyeOutlined />} onClick={onSpectate}>
              进入观战
            </Button>
          </div>
        </div>
      </Card>

      <Card
        title="比赛规则说明"
        className="border-0 shadow-sm"
        styles={{ header: { background: 'linear-gradient(90deg, rgba(250,173,20,0.08), rgba(24,144,255,0.08))' } }}
      >
        <Descriptions bordered size="small" column={2}>
          <Descriptions.Item label="策略接口">{battleConfig.judge.strategy_interface || '-'}</Descriptions.Item>
          <Descriptions.Item label="资源模型">{battleConfig.judge.resource_model || '-'}</Descriptions.Item>
          <Descriptions.Item label="评分模型">{battleConfig.judge.scoring_model || '-'}</Descriptions.Item>
          <Descriptions.Item label="总轮数">{battleConfig.lifecycle.total_rounds || 0}</Descriptions.Item>
          <Descriptions.Item label="单轮时长">{battleConfig.lifecycle.round_duration_seconds || 0} 秒</Descriptions.Item>
          <Descriptions.Item label="升级窗口">{battleConfig.lifecycle.upgrade_window_seconds || 0} 秒</Descriptions.Item>
          <Descriptions.Item label="允许动作" span={2}>
            {allowedActions.length > 0 ? allowedActions.map((action) => <Tag key={action}>{action}</Tag>) : '未配置'}
          </Descriptions.Item>
        </Descriptions>

        <Alert
          className="mt-4"
          type="info"
          showIcon
          message="提交说明"
          description="本场比赛提交的是策略代码，系统会按比赛规则执行并结算结果。"
        />
      </Card>

      <Card
        title="我的策略智能体"
        styles={{ header: { background: 'linear-gradient(90deg, rgba(114,46,209,0.08), rgba(24,144,255,0.08))' } }}
        className="border-0 shadow-sm"
      >
        {contractInfo && (
          <Alert
            className="mb-4"
            type={contractInfo.status === 'active' || contractInfo.status === 'deployed' ? 'success' : contractInfo.status === 'pending' ? 'info' : 'warning'}
            showIcon
            icon={<CheckCircleOutlined />}
            message={(
              <div className="flex items-center justify-between">
                <span>
                  策略版本 v{contractInfo.version}
                  {contractInfo.contract_address && (
                    <span className="ml-2 font-mono text-xs">{contractInfo.contract_address}</span>
                  )}
                </span>
                  <Tag color={contractInfo.status === 'active' || contractInfo.status === 'deployed' ? 'success' : 'processing'}>
                    {contractStatusLabel}
                  </Tag>
              </div>
            )}
            description={contractInfo.deployed_at ? `部署时间：${formatDateTime(contractInfo.deployed_at)}` : undefined}
          />
        )}

        <Tabs
          defaultActiveKey="code"
          items={[
            {
              key: 'code',
              label: <span><CodeOutlined /> 策略编写与部署</span>,
              children: (
                <div>
                  <p className="mb-3 text-text-secondary">
                    请围绕平台给定的策略接口实现资源调度、目标选择与攻防决策逻辑。
                  </p>

                  <TextArea
                    value={sourceCode}
                    onChange={(event: ChangeEvent<HTMLTextAreaElement>) => onSourceCodeChange(event.target.value)}
                    placeholder={STRATEGY_TEMPLATE}
                    rows={18}
                    className="mb-4 font-mono text-sm"
                    style={{ background: '#fafafa' }}
                  />

                  <Button type="primary" icon={<RocketOutlined />} loading={deploying} onClick={onDeploy} size="large">
                    {contractInfo ? '提交新策略版本' : '部署首个策略版本'}
                  </Button>
                </div>
              ),
            },
            {
              key: 'upload',
              label: <span><UploadOutlined /> 文件上传</span>,
              children: (
                <div>
                  <p className="mb-4 text-text-secondary">
                    你也可以直接上传完整的策略工程文件，例如 `.sol` 或 `.zip`。
                  </p>
                  <Upload accept=".sol,.zip" showUploadList={false} customRequest={onUploadFile}>
                    <Button icon={<UploadOutlined />} loading={uploading}>
                      上传策略代码
                    </Button>
                  </Upload>
                </div>
              ),
            },
          ]}
        />
      </Card>

      <Card title="评分权重" size="small" className="border-0 shadow-sm" styles={{ header: { background: 'linear-gradient(90deg, rgba(82,196,26,0.08), rgba(24,144,255,0.08))' } }}>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between"><span>资源控制</span><Tag color="green">{scoreWeights.resource || 0}</Tag></div>
          <div className="flex justify-between"><span>攻击收益</span><Tag color="green">{scoreWeights.attack || 0}</Tag></div>
          <div className="flex justify-between"><span>防御保全</span><Tag color="blue">{scoreWeights.defense || 0}</Tag></div>
          <div className="flex justify-between"><span>生存稳定</span><Tag color="purple">{scoreWeights.survival || 0}</Tag></div>
        </div>
      </Card>
    </div>
  )
}
