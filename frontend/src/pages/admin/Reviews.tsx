import { Tabs } from 'antd'

import { ChallengePublishReviewList, CrossSchoolReviewList } from '@/components/admin'
import { PageHeader } from '@/components/common'
import { usePersistedTab } from '@/hooks'

export default function AdminReviews() {
  const [activeTab, setActiveTab] = usePersistedTab(
    'admin_reviews',
    'cross_school',
    ['cross_school', 'challenge_publish'],
  )

  return (
    <div>
      <PageHeader
        title="审核管理"
        subtitle="处理跨校比赛和题目公开申请的审核事项"
      />

      <div className="card">
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            { key: 'cross_school', label: '跨校比赛审核', children: <CrossSchoolReviewList /> },
            { key: 'challenge_publish', label: '题目公开审核', children: <ChallengePublishReviewList /> },
          ]}
        />
      </div>
    </div>
  )
}
