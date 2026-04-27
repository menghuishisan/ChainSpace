import { useState } from 'react'
import { Button, Collapse, Empty, Form, Input, Modal, Popconfirm, Space, message } from 'antd'
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'

import { useCourseChapters, useCourseMaterials } from '@/hooks'
import type { Chapter } from '@/types'
import type { TeacherCourseContentTabProps } from '@/types/presentation'
import TeacherChapterMaterials from './TeacherChapterMaterials'

export default function TeacherCourseContentTab({
  courseId,
  chapters,
  onRefresh,
}: TeacherCourseContentTabProps) {
  const [chapterModalVisible, setChapterModalVisible] = useState(false)
  const [editingChapter, setEditingChapter] = useState<Chapter | null>(null)
  const [chapterForm] = Form.useForm()
  const [expandedChapters, setExpandedChapters] = useState<string[]>([])
  const { chapterMaterials, fetchMaterials } = useCourseMaterials(courseId)
  const { saveChapter, removeChapter } = useCourseChapters(courseId)

  const handleExpand = (keys: string[]) => {
    setExpandedChapters(keys)
    keys.forEach((key) => {
      const chapterId = parseInt(key, 10)
      if (!chapterMaterials[chapterId]) {
        void fetchMaterials(chapterId)
      }
    })
  }

  const handleEditChapter = (chapter?: Chapter) => {
    setEditingChapter(chapter || null)
    chapterForm.resetFields()
    if (chapter) {
      chapterForm.setFieldsValue(chapter)
    }
    setChapterModalVisible(true)
  }

  const handleSaveChapter = async (values: { title: string; description?: string }) => {
    try {
      await saveChapter(editingChapter?.id || null, values)
      message.success(editingChapter ? '章节更新成功' : '章节创建成功')
      setChapterModalVisible(false)
      onRefresh()
    } catch {
      // 交给请求层统一处理
    }
  }

  const handleDeleteChapter = async (chapterId: number) => {
    try {
      await removeChapter(chapterId)
      message.success('章节删除成功')
      onRefresh()
    } catch {
      // 交给请求层统一处理
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <span className="text-text-secondary">共 {chapters.length} 个章节</span>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => handleEditChapter()}>
          添加章节
        </Button>
      </div>

      {chapters.length === 0 ? (
        <Empty description="暂无章节，点击上方按钮添加" />
      ) : (
        <Collapse
          activeKey={expandedChapters}
          onChange={(keys) => handleExpand(keys as string[])}
          items={chapters.map((chapter, index) => ({
            key: chapter.id.toString(),
            label: (
              <div className="flex w-full items-center justify-between pr-4">
                <span>
                  <span className="mr-2 text-text-secondary">第 {index + 1} 章</span>
                  {chapter.title}
                </span>
                <Space onClick={(event) => event.stopPropagation()}>
                  <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditChapter(chapter)}>
                    编辑
                  </Button>
                  <Popconfirm title="确定要删除该章节吗？" onConfirm={() => void handleDeleteChapter(chapter.id)}>
                    <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                      删除
                    </Button>
                  </Popconfirm>
                </Space>
              </div>
            ),
            children: (
              <TeacherChapterMaterials
                courseId={courseId}
                chapterId={chapter.id}
                materials={chapterMaterials[chapter.id] || []}
              />
            ),
          }))}
        />
      )}

      <Modal
        title={editingChapter ? '编辑章节' : '添加章节'}
        open={chapterModalVisible}
        onCancel={() => setChapterModalVisible(false)}
        footer={null}
      >
        <Form form={chapterForm} layout="vertical" onFinish={(values) => void handleSaveChapter(values)} className="mt-4">
          <Form.Item name="title" label="章节标题" rules={[{ required: true, message: '请输入章节标题' }]}>
            <Input placeholder="请输入章节标题" />
          </Form.Item>
          <Form.Item name="description" label="章节描述">
            <Input.TextArea placeholder="请输入章节描述" rows={3} />
          </Form.Item>
          <Form.Item className="mb-0 text-right">
            <Space>
              <Button onClick={() => setChapterModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit">保存</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
