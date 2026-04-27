/**
 * 学生 - 我的队伍页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Button, Tag, Modal, List, Avatar, Empty, message, Divider } from 'antd'
import { TeamOutlined, UserOutlined, CopyOutlined, UserDeleteOutlined, LogoutOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import type { TeamWithContest } from '@/types/presentation'
import { getMyTeams, leaveTeam, removeTeamMember } from '@/api/contest'
import { formatDateTime } from '@/utils/format'
import { useUserStore } from '@/store'

export default function StudentTeams() {
  const { user } = useUserStore()
  const [, setLoading] = useState(false)
  const [teams, setTeams] = useState<TeamWithContest[]>([])
  const [selectedTeam, setSelectedTeam] = useState<TeamWithContest | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)

  // 获取我的队伍
  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getMyTeams()
      setTeams(result || [])
    } catch { /* */ } finally { setLoading(false) }
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  // 复制邀请码
  const copyInviteCode = (code: string) => {
    navigator.clipboard.writeText(code)
    message.success('邀请码已复制')
  }

  // 查看队伍详情
  const handleViewTeam = (team: TeamWithContest) => {
    setSelectedTeam(team)
    setDetailVisible(true)
  }

  // 退出队伍
  const handleLeaveTeam = (teamId: number) => {
    Modal.confirm({
      title: '确认退出',
      content: '确定要退出该队伍吗？',
      onOk: async () => {
        try {
          await leaveTeam(teamId)
          message.success('已退出队伍')
          fetchData()
          setDetailVisible(false)
        } catch { /* */ }
      }
    })
  }

  // 移除成员（仅队长）
  const handleRemoveMember = async (teamId: number, userId: number) => {
    try {
      await removeTeamMember(teamId, userId)
      message.success('已移除成员')
      fetchData()
    } catch { /* */ }
  }

  return (
    <div>
      <PageHeader title="我的队伍" subtitle="管理你参与的竞赛队伍" />

      {teams.length === 0 ? (
        <Card>
          <Empty description="暂无队伍，参加团队比赛时会自动创建或加入队伍" />
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {teams.map(team => (
            <Card key={team.id} hoverable onClick={() => handleViewTeam(team)}>
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-medium mb-2">{team.name}</h3>
                  <p className="text-text-secondary text-sm mb-2">
                    <TeamOutlined className="mr-1" />
                    {team.members?.length || 0} 名成员
                  </p>
                  {team.contest && (
                    <Tag color="blue">{team.contest.title}</Tag>
                  )}
                </div>
                <Tag color={team.status === 'competing' ? 'success' : team.status === 'ready' ? 'processing' : 'default'}>
                  {team.status === 'competing' ? '比赛中' : team.status === 'ready' ? '已就绪' : '组队中'}
                </Tag>
              </div>
              {team.captain_id === user?.id && (
                <div className="mt-4 pt-4 border-t">
                  <span className="text-xs text-warning">你是队长</span>
                </div>
              )}
            </Card>
          ))}
        </div>
      )}

      {/* 队伍详情弹窗 */}
      <Modal
        title="队伍详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={
          selectedTeam && selectedTeam.captain_id !== user?.id ? (
            <Button danger icon={<LogoutOutlined />} onClick={() => handleLeaveTeam(selectedTeam.id)}>退出队伍</Button>
          ) : null
        }
        width={500}
      >
        {selectedTeam && (
          <div>
            <div className="mb-4">
              <h4 className="text-lg font-medium">{selectedTeam.name}</h4>
              {selectedTeam.contest && <Tag color="blue" className="mt-2">{selectedTeam.contest.title}</Tag>}
            </div>

            <Divider />

            {/* 邀请码 */}
            {selectedTeam.invite_code && selectedTeam.captain_id === user?.id && (
              <div className="mb-4 p-3 bg-gray-50 rounded">
                <span className="text-text-secondary">邀请码：</span>
                <span className="font-mono font-medium mx-2">{selectedTeam.invite_code}</span>
                <Button type="link" size="small" icon={<CopyOutlined />} onClick={() => copyInviteCode(selectedTeam.invite_code!)}>复制</Button>
              </div>
            )}

            {/* 成员列表 */}
            <h5 className="font-medium mb-2">队伍成员</h5>
            <List
              dataSource={selectedTeam.members || []}
              renderItem={member => (
                <List.Item
                  actions={
                    selectedTeam.captain_id === user?.id && member.user_id !== user?.id ? [
                      <Button type="link" size="small" danger icon={<UserDeleteOutlined />} onClick={() => handleRemoveMember(selectedTeam.id, member.user_id)}>移除</Button>
                    ] : []
                  }
                >
                  <List.Item.Meta
                    avatar={<Avatar icon={<UserOutlined />} />}
                    title={
                      <span>
                        {member.real_name || member.display_name || member.student_no || member.phone || `成员 ${member.user_id}`}
                        {member.is_captain && <Tag color="gold" className="ml-2">队长</Tag>}
                      </span>
                    }
                    description={`加入时间：${formatDateTime(member.joined_at)}`}
                  />
                </List.Item>
              )}
            />
          </div>
        )}
      </Modal>
    </div>
  )
}
