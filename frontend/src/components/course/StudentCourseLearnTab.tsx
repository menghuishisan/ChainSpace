import { Collapse, Empty, List, Spin, Tag } from 'antd'
import { FileTextOutlined, LinkOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { useState } from 'react'

import { useCourseMaterials } from '@/hooks'
import { MaterialTypeMap } from '@/types'
import type { StudentCourseLearnTabProps } from '@/types/presentation'

export default function StudentCourseLearnTab({ courseId, chapters }: StudentCourseLearnTabProps) {
  const [expandedKeys, setExpandedKeys] = useState<string[]>([])
  const { chapterMaterials, fetchMaterials } = useCourseMaterials(courseId)

  const handleExpand = (keys: string[]) => {
    setExpandedKeys(keys)
    keys.forEach((key) => {
      const chapterId = parseInt(key, 10)
      if (!chapterMaterials[chapterId]) {
        void fetchMaterials(chapterId)
      }
    })
  }

  const getMaterialIcon = (type: string) => {
    switch (type) {
      case 'video':
        return <PlayCircleOutlined className="text-blue-500" />
      case 'link':
        return <LinkOutlined className="text-green-500" />
      default:
        return <FileTextOutlined className="text-orange-500" />
    }
  }

  if (chapters.length === 0) {
    return <Empty description="暂无章节内容" />
  }

  return (
    <Collapse
      activeKey={expandedKeys}
      onChange={(keys) => handleExpand(keys as string[])}
      items={chapters.map((chapter, index) => ({
        key: chapter.id.toString(),
        label: (
          <span>
            <span className="mr-2 text-text-secondary">第 {index + 1} 章</span>
            {chapter.title}
          </span>
        ),
        children: !chapterMaterials[chapter.id]
          ? <Spin />
          : chapterMaterials[chapter.id].length === 0
            ? <div className="py-4 text-center text-text-secondary">暂无资料</div>
            : (
                <List
                  size="small"
                  dataSource={chapterMaterials[chapter.id]}
                  renderItem={(item) => (
                    <List.Item className="cursor-pointer hover:bg-gray-50">
                      <List.Item.Meta
                        avatar={getMaterialIcon(item.type)}
                        title={item.title}
                        description={<Tag>{MaterialTypeMap[item.type as keyof typeof MaterialTypeMap]}</Tag>}
                      />
                    </List.Item>
                  )}
                />
              ),
      }))}
    />
  )
}
