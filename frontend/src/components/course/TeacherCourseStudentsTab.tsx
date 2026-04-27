import { useState } from 'react'
import { Alert, Button, Empty, Input, List, Modal, Popconfirm, Space, Spin, Upload, message } from 'antd'
import { DownloadOutlined, PlusOutlined, UploadOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'

import { useCourseStudents } from '@/hooks'
import type { TeacherCourseStudentsTabProps } from '@/types/presentation'

function downloadStudentTemplate() {
  const templateData = [
    ['phone', 'student_no', 'real_name', 'class_name'],
    ['13800138001', '2024001', '张三', '计算机 1 班'],
    ['13800138002', '2024002', '李四', '计算机 1 班'],
    ['13800138003', '2024003', '王五', '计算机 1 班'],
  ]

  const csvContent = templateData.map((row) => row.join(',')).join('\n')
  const blob = new Blob([`\ufeff${csvContent}`], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = '学生导入模板.csv'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

export default function TeacherCourseStudentsTab({ courseId }: TeacherCourseStudentsTabProps) {
  const [importModalVisible, setImportModalVisible] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<{ success: number; failed: number; errors?: string[] } | null>(null)
  const { loading, students, addStudentsByPhones, removeStudent, importStudents } = useCourseStudents(courseId)

  const handleRemove = async (studentId: number) => {
    try {
      await removeStudent(studentId)
      message.success('已移除学生')
    } catch {
      // 交给请求层统一处理
    }
  }

  const handleAddStudents = () => {
    let inputValue = ''
    Modal.confirm({
      title: '添加学生',
      content: (
        <div>
          <p className="mb-2 text-text-secondary">请输入学生手机号，多个用逗号或换行分隔</p>
          <Input.TextArea
            rows={4}
            placeholder="例如: 13800138001, 13900139002"
            onChange={(event) => { inputValue = event.target.value }}
          />
        </div>
      ),
      okText: '添加',
      cancelText: '取消',
      onOk: async () => {
        const phones = inputValue.split(/[,\n]/).map((item) => item.trim()).filter((item) => item.length > 0)
        if (phones.length === 0) {
          message.warning('请输入有效的手机号')
          return
        }

        try {
          await addStudentsByPhones(phones)
          message.success(`成功添加 ${phones.length} 名学生`)
        } catch {
          // 交给请求层统一处理
        }
      },
    })
  }

  const handleExcelUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options
    setImporting(true)
    setImportResult(null)
    try {
      const result = await importStudents(file as File)
      setImportResult(result)
      if (result.failed === 0) {
        message.success(`成功导入 ${result.success} 名学生`)
      } else if (result.success > 0) {
        message.warning(`部分成功：导入 ${result.success} 名，失败 ${result.failed} 名`)
      } else {
        message.error('导入失败，请检查 Excel 格式')
      }
      onSuccess?.(result)
    } catch (error) {
      onError?.(error as Error)
      message.error('导入失败')
    } finally {
      setImporting(false)
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <span className="text-text-secondary">共 {students.total} 名学生</span>
        <Space>
          <Button icon={<UploadOutlined />} onClick={() => setImportModalVisible(true)}>
            批量导入
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAddStudents}>
            添加学生
          </Button>
        </Space>
      </div>

      <Spin spinning={loading}>
        {students.list.length === 0 ? (
          <Empty description="暂无学生加入课程" />
        ) : (
          <List
            dataSource={students.list}
            renderItem={(item) => (
              <List.Item
                actions={[
                  <span key="progress">{item.progress || 0}%</span>,
                  (
                    <Popconfirm
                      key="remove"
                      title="确定要移除该学生吗？"
                      onConfirm={() => void handleRemove(item.student_id)}
                    >
                      <Button type="link" danger size="small">移除</Button>
                    </Popconfirm>
                  ),
                ]}
              >
                <List.Item.Meta
                  title={item.real_name || item.student_no || item.phone || '未命名学生'}
                  description={(
                    <Space>
                      <span>学号: {item.student_no}</span>
                      {item.class_name && <span>班级: {item.class_name}</span>}
                      <span>加入时间: {item.joined_at}</span>
                    </Space>
                  )}
                />
              </List.Item>
            )}
          />
        )}
      </Spin>

      <Modal
        title="批量导入学生"
        open={importModalVisible}
        onCancel={() => {
          setImportModalVisible(false)
          setImportResult(null)
        }}
        footer={null}
        width={600}
      >
        <div className="py-4">
          <Alert
            message="导入说明"
            description={(
              <div className="text-sm">
                <p>1. 请先下载 Excel 模板，按模板格式填写学生信息。</p>
                <p>2. 必填字段：`phone`、`student_no`、`real_name`、`class_name`。</p>
                <p>3. 手机号用于登录，班级不存在时会自动创建。</p>
                <p>4. 导入后学生将自动加入当前课程，首次登录需要修改密码。</p>
              </div>
            )}
            type="info"
            showIcon
            className="mb-4"
          />

          <div className="mb-4">
            <Button icon={<DownloadOutlined />} onClick={downloadStudentTemplate}>
              下载导入模板
            </Button>
          </div>

          <Upload.Dragger
            accept=".xlsx,.xls,.csv"
            customRequest={handleExcelUpload}
            showUploadList={false}
            disabled={importing}
          >
            <p className="ant-upload-drag-icon">
              <UploadOutlined style={{ fontSize: 48, color: '#1890ff' }} />
            </p>
            <p className="ant-upload-text">点击或拖拽上传 Excel 文件</p>
            <p className="ant-upload-hint">支持 .xlsx、.xls、.csv 格式</p>
          </Upload.Dragger>

          {importing && (
            <div className="mt-4 text-center">
              <Spin tip="正在导入..." />
            </div>
          )}

          {importResult && (
            <div className="mt-4">
              <Alert
                message="导入完成"
                description={(
                  <div>
                    <p>成功导入: {importResult.success} 名学生</p>
                    <p>导入失败: {importResult.failed} 名学生</p>
                    {importResult.errors && importResult.errors.length > 0 && (
                      <div className="mt-2">
                        <p className="font-medium">错误详情:</p>
                        <ul className="list-inside list-disc text-sm">
                          {importResult.errors.slice(0, 10).map((err, idx) => (
                            <li key={idx}>{err}</li>
                          ))}
                          {importResult.errors.length > 10 && (
                            <li>...还有 {importResult.errors.length - 10} 条错误</li>
                          )}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
                type={importResult.failed === 0 ? 'success' : importResult.success > 0 ? 'warning' : 'error'}
                showIcon
              />
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}
