import { useNavigate } from 'react-router-dom'
import { Button, Empty, List, Spin, Tag } from 'antd'
import { ExperimentOutlined } from '@ant-design/icons'

import { useCourseExperiments } from '@/hooks'
import { ExperimentTypeMap } from '@/types'
import type { StudentCourseExperimentsTabProps } from '@/types/presentation'

export default function StudentCourseExperimentsTab({ courseId }: StudentCourseExperimentsTabProps) {
  const navigate = useNavigate()
  const { loading, experiments } = useCourseExperiments(courseId)

  return (
    <Spin spinning={loading}>
      {experiments.length === 0 ? (
        <Empty description="暂无实验" />
      ) : (
        <List
          dataSource={experiments}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button
                  key="start"
                  type="primary"
                  onClick={() => navigate(`/student/experiments/${item.id}`)}
                >
                  进入实验
                </Button>,
              ]}
            >
              <List.Item.Meta
                avatar={<ExperimentOutlined className="text-2xl text-primary" />}
                title={item.title}
                description={(
                  <div>
                    <Tag>{ExperimentTypeMap[item.type as keyof typeof ExperimentTypeMap] || item.type}</Tag>
                    <span className="ml-2">满分: {item.max_score} 分</span>
                    {item.estimated_time && <span className="ml-2">预计: {item.estimated_time} 分钟</span>}
                  </div>
                )}
              />
            </List.Item>
          )}
        />
      )}
    </Spin>
  )
}
