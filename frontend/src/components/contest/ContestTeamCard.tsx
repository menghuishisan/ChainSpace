import { Avatar, Card, Empty, List, Tag } from 'antd'
import { UserOutlined } from '@ant-design/icons'

import type { Team } from '@/types'

interface ContestTeamCardProps {
  team: Team
}

export default function ContestTeamCard({ team }: ContestTeamCardProps) {
  return (
    <Card
      title="我的队伍"
      className="border-0 shadow-sm"
      styles={{
        body: { padding: 14 },
        header: {
          background: 'linear-gradient(90deg, rgba(114,46,209,0.08), rgba(24,144,255,0.08))',
          minHeight: 42,
          padding: '0 14px',
        },
      }}
    >
      <div className="mb-2.5">
        <h4 className="text-sm font-semibold leading-6">{team.name}</h4>
        <Tag color="blue" className="mt-1.5">邀请码：{team.invite_code || '-'}</Tag>
      </div>
      {team.members?.length ? (
        <List
          size="small"
          dataSource={team.members}
          renderItem={(member) => (
            <List.Item className="px-0 py-1.5">
              <List.Item.Meta
                avatar={<Avatar size="small" icon={<UserOutlined />} />}
                title={<span className="text-sm leading-5">{member.real_name || member.display_name || member.student_no || member.phone || '未命名成员'}</span>}
                description={member.is_captain ? <Tag color="gold" className="mt-0.5">队长</Tag> : null}
              />
            </List.Item>
          )}
        />
      ) : (
        <Empty description="暂无队伍成员" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      )}
    </Card>
  )
}
