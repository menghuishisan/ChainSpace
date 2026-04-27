import {
  CheckCircleOutlined,
  CopyOutlined,
  DownloadOutlined,
  EnvironmentOutlined,
  FileTextOutlined,
  FlagOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  QuestionCircleOutlined,
  ReloadOutlined,
  StopOutlined,
} from '@ant-design/icons'
import {
  Alert,
  Badge,
  Button,
  Card,
  Descriptions,
  Empty,
  Input,
  List,
  Space,
  Tag,
  Typography,
} from 'antd'

import {
  extractContractName,
  getRuntimeDescription,
  getRuntimeHeadline,
} from '@/domains/contest/jeopardy'
import { normalizeRuntimeWorkbenchToolKind } from '@/domains/runtime/workbench'
import type { ChallengeOrchestration } from '@/types'
import type { JeopardySidebarProps } from '@/types/presentation'
import { formatDuration } from '@/utils/format'

const { Paragraph, Title } = Typography

function renderRuntimePanel({
  runtimeProfile,
  challengeOrchestration,
  accessUrl,
  rpcAvailable,
  serviceEntries,
}: {
  runtimeProfile: string
  challengeOrchestration: ChallengeOrchestration
  accessUrl?: string
  rpcAvailable?: boolean
  serviceEntries?: Array<{
    key: string
    label: string
    description?: string
    purpose?: string
    access_url?: string
    protocol?: string
    port?: number
    expose_as?: string
  }>
}) {
  switch (runtimeProfile) {
    case 'single_chain_instance':
      return (
        <Descriptions size="small" column={1}>
          <Descriptions.Item label="环境状态">{accessUrl ? '环境已准备' : '启动后可用'}</Descriptions.Item>
          <Descriptions.Item label="可用能力">独立链实例、工作区、链上交互调试</Descriptions.Item>
          <Descriptions.Item label="使用时长">{challengeOrchestration.lifecycle?.time_limit_minutes || 0} 分钟</Descriptions.Item>
        </Descriptions>
      )
    case 'fork_replay':
      return (
        <Descriptions size="small" column={1}>
          <Descriptions.Item label="环境状态">{accessUrl ? '环境已准备' : '启动后可用'}</Descriptions.Item>
          <Descriptions.Item label="目标链">
            {challengeOrchestration.fork?.label || challengeOrchestration.fork?.chain || '未配置'}
          </Descriptions.Item>
          <Descriptions.Item label="目标区块">
            {challengeOrchestration.fork?.block_number || '未配置'}
          </Descriptions.Item>
          <Descriptions.Item label="调试方式">{rpcAvailable ? '已提供链上调试与日志能力' : '启动后可进行链上调试'}</Descriptions.Item>
        </Descriptions>
      )
    case 'multi_service_lab':
      return (
        <div className="space-y-3">
          <Descriptions size="small" column={1}>
            <Descriptions.Item label="环境状态">{accessUrl ? '环境已准备' : '启动后可用'}</Descriptions.Item>
            <Descriptions.Item label="关键组件">{serviceEntries?.length || challengeOrchestration.services?.length || 0} 个</Descriptions.Item>
            <Descriptions.Item label="使用方式">启动后按题目要求使用环境和服务完成分析。</Descriptions.Item>
          </Descriptions>

          <List
            size="small"
            bordered
            dataSource={serviceEntries?.length
              ? serviceEntries
              : (challengeOrchestration.services || []).map((service) => ({
                key: service.key,
                label: service.key,
                description: service.description,
                purpose: service.purpose,
                protocol: service.ports?.[0]?.protocol,
                port: service.ports?.[0]?.port,
                expose_as: service.ports?.[0]?.expose_as,
              }))}
            locale={{ emptyText: '当前没有附加服务' }}
            renderItem={(service) => (
              <List.Item>
                <div className="w-full">
                  <div className="font-medium">{service.label || service.key}</div>
                  <div className="text-xs text-text-secondary">
                    {service.description || service.purpose || '暂无说明'}
                  </div>
                </div>
              </List.Item>
            )}
          />
        </div>
      )
    default:
      return (
        <Paragraph type="secondary" className="mb-0">
          本题不需要启动环境，请结合题面、附件和代码完成分析与提交。
        </Paragraph>
      )
  }
}

export default function JeopardySidebar({
  selectedChallenge,
  categoryMap,
  difficultyMap,
  renderedDescription,
  showDescriptionCard = true,
  firstBloodBonus,
  challengeEnv,
  envLoading,
  envRemaining,
  fetchingEnv,
  flagInput,
  submitting,
  onRefreshEnv,
  onStartEnv,
  onStopEnv,
  onOpenCodeViewer,
  onCopyCode,
  onDownloadCode,
  onOpenAttachment,
  onFlagInputChange,
  onSubmitFlag,
}: JeopardySidebarProps) {
  if (!selectedChallenge) {
    return (
      <Card className="py-12 text-center text-text-secondary">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择一道题目查看详情" />
      </Card>
    )
  }

  const difficulty = difficultyMap[selectedChallenge.difficulty]
  const needsEnvironment = selectedChallenge.challenge_orchestration?.needs_environment
  const rpcAvailable = (challengeEnv?.tools || []).some((tool) => normalizeRuntimeWorkbenchToolKind(tool.kind || tool.key) === 'rpc')

  return (
    <div className="space-y-3">
      <Card
        size="small"
        title={(
          <span className="flex items-center gap-2">
            <Badge count={selectedChallenge.is_solved ? <CheckCircleOutlined className="text-base text-success" /> : 0} offset={[-4, 4]}>
              <span className="text-base font-semibold">{selectedChallenge.title}</span>
            </Badge>
          </span>
        )}
        extra={(
          <Button
            type="text"
            size="small"
            icon={<ReloadOutlined />}
            loading={fetchingEnv}
            onClick={onRefreshEnv}
          >
            刷新
          </Button>
        )}
      >
        <Descriptions column={2} size="small">
          <Descriptions.Item label="知识主题">
            <Tag>{categoryMap[selectedChallenge.category] || selectedChallenge.category}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="难度">
            <Tag color={difficulty?.color}>{difficulty?.text}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="分值">
            <span className="text-base font-bold text-primary">
              {selectedChallenge.points ?? selectedChallenge.base_points}
            </span>
          </Descriptions.Item>
          <Descriptions.Item label="已解人数">
            {selectedChallenge.solve_count ?? 0} 人
          </Descriptions.Item>
          <Descriptions.Item label="题目类型">
            {needsEnvironment ? <Tag color="cyan">环境题</Tag> : <Tag>静态题</Tag>}
          </Descriptions.Item>
        </Descriptions>
        {selectedChallenge.first_blood ? (
          <Alert
            type="success"
            showIcon
            className="mt-3"
            message={`一血：${selectedChallenge.first_blood}`}
            description={firstBloodBonus ? `当前题目首解奖励 ${firstBloodBonus} 分` : '当前题目已产生首解'}
          />
        ) : firstBloodBonus ? (
          <Alert
            type="info"
            showIcon
            className="mt-3"
            message="一血奖励未触发"
            description={`本场比赛首解奖励为 ${firstBloodBonus} 分`}
          />
        ) : null}
      </Card>

      <Alert
        type={needsEnvironment ? 'info' : 'success'}
        showIcon
        icon={needsEnvironment ? <EnvironmentOutlined /> : <PauseCircleOutlined />}
        message={getRuntimeHeadline(selectedChallenge.runtime_profile)}
        description={getRuntimeDescription(selectedChallenge.runtime_profile)}
      />

      <Card size="small" className={needsEnvironment ? 'border-blue-200' : ''}>
        <div className="mb-2 flex items-center">
          <EnvironmentOutlined className="mr-2 text-blue-500" />
          <strong>{needsEnvironment ? '题目环境' : '解题说明'}</strong>
        </div>

        {selectedChallenge.challenge_orchestration.scenario?.attack_goal ? (
          <Alert
            type="warning"
            showIcon
            className="mb-3"
            message="攻击目标"
            description={selectedChallenge.challenge_orchestration.scenario.attack_goal}
          />
        ) : null}

        {renderRuntimePanel({
          runtimeProfile: selectedChallenge.runtime_profile,
          challengeOrchestration: selectedChallenge.challenge_orchestration,
          accessUrl: challengeEnv?.access_url,
          rpcAvailable,
          serviceEntries: challengeEnv?.service_entries,
        })}

        {needsEnvironment ? (
          <div className="mt-4">
            {challengeEnv?.status === 'running' ? (
              <div>
                <Alert
                  type="success"
                  showIcon
                  message={`环境运行中，剩余 ${formatDuration(envRemaining)}`}
                  className="mb-3"
                />
                <Alert
                  type="info"
                  showIcon
                  className="mb-3"
                  message="可在下方继续操作"
                  description="题目环境已经就绪，你可以继续使用下方区域完成分析和操作。"
                />
                <div className="mt-3">
                  <Button danger block icon={<StopOutlined />} loading={envLoading} onClick={onStopEnv}>
                    停止环境
                  </Button>
                </div>
              </div>
            ) : challengeEnv?.status === 'creating' ? (
              <div className="py-4 text-center text-text-secondary">环境创建中，请稍候...</div>
            ) : challengeEnv?.status === 'failed' ? (
              <div>
                <Alert
                  type="error"
                  showIcon
                  className="mb-3"
                  message="环境启动失败"
                  description={challengeEnv.error_message || '当前没有返回更详细的错误信息，请稍后重试。'}
                />
                <Button type="primary" block icon={<PlayCircleOutlined />} loading={envLoading} onClick={onStartEnv}>
                  重新启动环境
                </Button>
              </div>
            ) : challengeEnv?.status === 'expired' ? (
              <div>
                <Alert
                  type="warning"
                  showIcon
                  className="mb-3"
                  message="环境已过期"
                  description="你可以重新启动题目环境，继续完成漏洞复现或系统分析。"
                />
                <Button type="primary" block icon={<PlayCircleOutlined />} loading={envLoading} onClick={onStartEnv}>
                  重新启动环境
                </Button>
              </div>
            ) : (
              <Button type="primary" block icon={<PlayCircleOutlined />} loading={envLoading} onClick={onStartEnv}>
                启动环境
              </Button>
            )}
          </div>
        ) : null}
      </Card>

      {showDescriptionCard ? (
        <Card
          size="small"
          title={<><FileTextOutlined className="mr-1" />题目描述</>}
          styles={{ body: { maxHeight: 300, overflow: 'auto' } }}
        >
          <div className="prose prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: renderedDescription }} />
        </Card>
      ) : null}

      {selectedChallenge.contract_code ? (
        <Card
          size="small"
          title="合约代码"
          extra={(
            <Space size="small">
              <Button type="text" size="small" icon={<CopyOutlined />} onClick={() => onCopyCode(selectedChallenge.contract_code || '')}>
                复制
              </Button>
              <Button
                type="text"
                size="small"
                icon={<DownloadOutlined />}
                onClick={() => {
                  const code = selectedChallenge.contract_code || ''
                  onDownloadCode(`${extractContractName(code)}.sol`, code)
                }}
              >
                下载
              </Button>
            </Space>
          )}
        >
          <Button
            type="primary"
            block
            onClick={() => onOpenCodeViewer(selectedChallenge.contract_code || '', `${selectedChallenge.title} - 合约代码`)}
          >
            在 Monaco 编辑器中查看代码
          </Button>
          <div className="mt-2 line-clamp-4 font-mono text-xs text-text-secondary">
            {selectedChallenge.contract_code.slice(0, 200)}...
          </div>
        </Card>
      ) : null}

      {selectedChallenge.hints && selectedChallenge.hints.length > 0 ? (
        <Card size="small" title={<><QuestionCircleOutlined className="mr-1" />题目提示（{selectedChallenge.hints.length}）</>}>
          <List
            size="small"
            dataSource={selectedChallenge.hints}
            renderItem={(hint, index) => (
              <List.Item key={`${selectedChallenge.id}-${index}`}>
                <div className="flex w-full items-start gap-2">
                  <Badge count={index + 1} style={{ backgroundColor: '#1677ff' }} />
                  <span className="text-sm text-text-secondary">
                    {typeof hint === 'string' ? hint : hint.content}
                  </span>
                </div>
              </List.Item>
            )}
          />
        </Card>
      ) : null}

      {selectedChallenge.attachments && selectedChallenge.attachments.length > 0 ? (
        <Card size="small" title="题目附件">
          <List
            size="small"
            dataSource={selectedChallenge.attachments}
            renderItem={(attachment, index) => (
              <List.Item>
                <Button
                  type="link"
                  className="truncate px-0"
                  onClick={() => onOpenAttachment(index)}
                >
                  {attachment.split('/').pop() || `附件 ${index + 1}`}
                </Button>
              </List.Item>
            )}
          />
        </Card>
      ) : null}

      {selectedChallenge.is_solved ? (
        <Card size="small" className="border-green-200 bg-green-50">
          <div className="py-4 text-center">
            <CheckCircleOutlined className="mb-2 text-3xl text-success" />
            <Paragraph className="mb-0 font-medium text-success">该题已完成</Paragraph>
            <Paragraph type="secondary" className="mb-0 text-xs">
              已获得 {selectedChallenge.points ?? selectedChallenge.base_points} 分
            </Paragraph>
          </div>
        </Card>
      ) : (
        <Card size="small">
          <Space direction="vertical" className="w-full" size="middle">
            <Title level={5} className="mb-0">
              <FlagOutlined className="mr-1" />
              提交 Flag
            </Title>
            <Input.Search
              placeholder="输入 Flag 并提交"
              value={flagInput}
              onChange={(event) => onFlagInputChange(event.target.value)}
              onSearch={onSubmitFlag}
              enterButton="提交"
              loading={submitting}
              allowClear
            />
          </Space>
        </Card>
      )}
    </div>
  )
}
