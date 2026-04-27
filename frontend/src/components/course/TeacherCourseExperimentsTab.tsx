import { useNavigate } from 'react-router-dom'
import { Button, Empty, List, Space, Spin, Tag } from 'antd'
import { ExperimentOutlined, PlusOutlined } from '@ant-design/icons'

import { useCourseExperiments } from '@/hooks'
import { ExperimentTypeMap } from '@/types'
import type { TeacherCourseExperimentsTabProps } from '@/types/presentation'

export default function TeacherCourseExperimentsTab({ courseId }: TeacherCourseExperimentsTabProps) {
  const navigate = useNavigate()
  const { loading, experiments } = useCourseExperiments(courseId)

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <span className="text-text-secondary">共 {experiments.length} 个实验</span>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/teacher/experiments/create?course_id=${courseId}`)}
        >
          添加实验
        </Button>
      </div>

      <Spin spinning={loading}>
        {experiments.length === 0 ? (
          <Empty description="暂无实验" />
        ) : (
          <List
            dataSource={experiments}
            renderItem={(item) => (
              <List.Item
                actions={[
                  <Button key="edit" type="link" onClick={() => navigate(`/teacher/experiments/${item.id}/edit`)}>编辑</Button>,
                  <Button key="grade" type="link" onClick={() => navigate('/teacher/grading')}>批改</Button>,
                ]}
              >
                <List.Item.Meta
                  avatar={<ExperimentOutlined className="text-2xl text-primary" />}
                  title={item.title}
                  description={(
                    <Space>
                      <Tag>{ExperimentTypeMap[item.type as keyof typeof ExperimentTypeMap] || item.type}</Tag>
                      <span>满分: {item.max_score} 分</span>
                      {item.estimated_time && <span>预计: {item.estimated_time} 分钟</span>}
                    </Space>
                  )}
                />
              </List.Item>
            )}
          />
        )}
      </Spin>
    </div>
  )
}
