import { Button, Descriptions, Input, Modal, Space } from 'antd'
import { CopyOutlined, DownloadOutlined } from '@ant-design/icons'
import CodeViewerModal from '@/components/contest/CodeViewerModal'
import type { ChallengeDetailModalProps } from '@/types/presentation'
import { CategoryMap, ChallengeRuntimeProfileMap, DifficultyMap } from '@/types'
import { useState } from 'react'

function downloadTextFile(filename: string, content: string) {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function extractContractName(code: string): string {
  const match = code.match(/contract\s+(\w+)/)
  return match ? match[1] : 'contract'
}

export default function ChallengeDetailModal({
  open,
  challenge,
  onCancel,
}: ChallengeDetailModalProps) {
  const [codeViewerOpen, setCodeViewerOpen] = useState(false)

  if (!challenge) {
    return null
  }

  return (
    <>
      <Modal
        title={challenge.title || '题目详情'}
        open={open}
        onCancel={onCancel}
        footer={<Button onClick={onCancel}>关闭</Button>}
        width={780}
      >
        <Descriptions bordered size="small" column={2}>
          <Descriptions.Item label="题目名称" span={2}>{challenge.title}</Descriptions.Item>
          <Descriptions.Item label="知识主题">
            {CategoryMap[challenge.category] || challenge.category}
          </Descriptions.Item>
          <Descriptions.Item label="环境类型">
            {ChallengeRuntimeProfileMap[challenge.runtime_profile] || challenge.runtime_profile}
          </Descriptions.Item>
          <Descriptions.Item label="难度">
            {DifficultyMap[challenge.difficulty]?.text || challenge.difficulty}
          </Descriptions.Item>
          <Descriptions.Item label="基础分值">{challenge.base_points}</Descriptions.Item>
          <Descriptions.Item label="环境类型">
            {challenge.challenge_orchestration.needs_environment ? '环境题' : '静态题'}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">{challenge.created_at}</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>
            <div style={{ whiteSpace: 'pre-wrap', maxHeight: 220, overflow: 'auto' }}>
              {challenge.description || '暂无描述'}
            </div>
          </Descriptions.Item>
        </Descriptions>

        {challenge.challenge_orchestration.scenario?.attack_goal && (
          <div className="mt-4">
            <div className="mb-2 font-medium">题目目标</div>
            <Input.TextArea value={challenge.challenge_orchestration.scenario.attack_goal} rows={3} readOnly />
          </div>
        )}

        {challenge.contract_code && (
          <div className="mt-4">
            <div className="mb-2 flex items-center justify-between">
              <div className="font-medium">合约代码</div>
              <Space>
                <Button
                  type="text"
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => {
                    navigator.clipboard.writeText(challenge.contract_code || '')
                  }}
                >
                  复制
                </Button>
                <Button
                  type="text"
                  size="small"
                  icon={<DownloadOutlined />}
                  onClick={() => {
                    const code = challenge.contract_code || ''
                    downloadTextFile(`${extractContractName(code)}.sol`, code)
                  }}
                >
                  下载
                </Button>
              </Space>
            </div>
            <Button type="primary" block onClick={() => setCodeViewerOpen(true)}>
              在 Monaco 编辑器中查看代码
            </Button>
            <div className="mt-2 rounded bg-gray-50 p-2 font-mono text-xs text-text-secondary line-clamp-4">
              {challenge.contract_code.slice(0, 300)}
              {challenge.contract_code.length > 300 && '...'}
            </div>
          </div>
        )}
      </Modal>

      <CodeViewerModal
        open={codeViewerOpen}
        code={challenge.contract_code || ''}
        title={`${challenge.title} - 合约代码`}
        onClose={() => setCodeViewerOpen(false)}
      />
    </>
  )
}
