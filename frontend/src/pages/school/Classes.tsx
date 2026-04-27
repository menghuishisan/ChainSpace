/**
 * 学校管理员 - 班级管理页面
 */
import { useState, useEffect } from 'react'
import { Card, List, Tag, Spin } from 'antd'
import { TeamOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import { getClasses, getClassStudents } from '@/api/school'
import type { PaginatedData, User } from '@/types'
import type { SchoolClassItem } from '@/types/presentation'

export default function SchoolClasses() {
  const [loading, setLoading] = useState(false)
  const [classes, setClasses] = useState<SchoolClassItem[]>([])
  const [selectedClass, setSelectedClass] = useState<SchoolClassItem | null>(null)
  const [students, setStudents] = useState<User[]>([])
  const [loadingStudents, setLoadingStudents] = useState(false)

  useEffect(() => {
    const fetchClasses = async () => {
      setLoading(true)
      try {
        const data = await getClasses() as PaginatedData<SchoolClassItem>
        setClasses(data.list)
      } catch {
        // 错误由拦截器处理
      } finally {
        setLoading(false)
      }
    }
    fetchClasses()
  }, [])

  const handleSelectClass = async (classItem: SchoolClassItem) => {
    setSelectedClass(classItem)
    setLoadingStudents(true)
    try {
      const data = await getClassStudents(classItem.id)
      setStudents(data.list)
    } catch {
      // 错误由拦截器处理
    } finally {
      setLoadingStudents(false)
    }
  }

  return (
    <div>
      <PageHeader title="班级管理" subtitle="查看和管理学校的班级信息" />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 班级列表 */}
        <Card title="班级列表" className="lg:col-span-1">
          <Spin spinning={loading}>
            {classes.length === 0 ? (
              <div className="text-center text-text-secondary py-8">暂无班级</div>
            ) : (
              <List
                dataSource={classes}
                renderItem={(item) => (
                  <List.Item
                    className={`cursor-pointer hover:bg-gray-50 ${selectedClass === item ? 'bg-blue-50' : ''}`}
                    onClick={() => handleSelectClass(item)}
                  >
                    <div className="flex items-center">
                      <TeamOutlined className="mr-2 text-primary" />
                      <span>{item.name}</span>
                    </div>
                  </List.Item>
                )}
              />
            )}
          </Spin>
        </Card>

        {/* 学生列表 */}
        <Card title={selectedClass ? `${selectedClass.name} - 学生列表` : '请选择班级'} className="lg:col-span-2">
          <Spin spinning={loadingStudents}>
            {!selectedClass ? (
              <div className="text-center text-text-secondary py-8">请先选择一个班级</div>
            ) : students.length === 0 ? (
              <div className="text-center text-text-secondary py-8">该班级暂无学生</div>
            ) : (
              <List
                dataSource={students}
                renderItem={(student) => (
                  <List.Item>
                    <List.Item.Meta
                      title={student.real_name}
                      description={`学号：${student.student_no || '-'}`}
                    />
                    <Tag color={student.status === 'active' ? 'success' : 'error'}>
                      {student.status === 'active' ? '正常' : '已禁用'}
                    </Tag>
                  </List.Item>
                )}
              />
            )}
          </Spin>
        </Card>
      </div>
    </div>
  )
}
