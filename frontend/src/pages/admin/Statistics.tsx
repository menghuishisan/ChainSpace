/**
 * 平台管理员 - 运营统计页面
 * 使用后端 /system/stats API
 */
import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Spin } from 'antd'
import { TeamOutlined, BookOutlined, ExperimentOutlined, CloudServerOutlined, BankOutlined, TrophyOutlined } from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import type { SystemStats } from '@/types'
import { getSystemStats } from '@/api/admin'

export default function Statistics() {
  const [loading, setLoading] = useState(false)
  const [stats, setStats] = useState<SystemStats | null>(null)

  useEffect(() => {
    const fetchStats = async () => {
      setLoading(true)
      try {
        const data = await getSystemStats()
        setStats(data)
      } catch {
        // 错误由拦截器处理
      } finally {
        setLoading(false)
      }
    }
    fetchStats()
  }, [])

  const statItems = [
    { title: '学校数量', value: stats?.total_schools || 0, icon: <BankOutlined />, color: '#1890FF' },
    { title: '用户总数', value: stats?.total_users || 0, icon: <TeamOutlined />, color: '#52C41A' },
    { title: '课程数量', value: stats?.total_courses || 0, icon: <BookOutlined />, color: '#722ED1' },
    { title: '实验数量', value: stats?.total_experiments || 0, icon: <ExperimentOutlined />, color: '#FA8C16' },
    { title: '竞赛数量', value: stats?.total_contests || 0, icon: <TrophyOutlined />, color: '#EB2F96' },
    { title: '运行中环境', value: stats?.active_envs || 0, icon: <CloudServerOutlined />, color: '#13C2C2' },
  ]

  return (
    <div>
      <PageHeader title="运营统计" subtitle="查看平台整体运营数据" />

      <Spin spinning={loading}>
        <Row gutter={[24, 24]}>
          {statItems.map((item, index) => (
            <Col xs={24} sm={12} md={8} lg={4} key={index}>
              <Card>
                <Statistic
                  title={item.title}
                  value={item.value}
                  prefix={<span style={{ color: item.color }}>{item.icon}</span>}
                  valueStyle={{ color: item.color }}
                />
              </Card>
            </Col>
          ))}
        </Row>

        <Row gutter={[24, 24]} className="mt-6">
          <Col span={24}>
            <Card title="系统信息">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="text-center p-4 bg-gray-50 rounded">
                  <div className="text-2xl font-bold text-primary">{stats?.online_users || 0}</div>
                  <div className="text-text-secondary">在线用户</div>
                </div>
                <div className="text-center p-4 bg-gray-50 rounded">
                  <div className="text-2xl font-bold text-success">{stats?.active_envs || 0}</div>
                  <div className="text-text-secondary">运行环境</div>
                </div>
                <div className="text-center p-4 bg-gray-50 rounded">
                  <div className="text-sm font-mono text-gray-600">{stats?.server_uptime || '-'}</div>
                  <div className="text-text-secondary">服务运行时间</div>
                </div>
                <div className="text-center p-4 bg-gray-50 rounded">
                  <div className="text-sm font-mono text-gray-600">{stats?.go_version || '-'}</div>
                  <div className="text-text-secondary">Go版本</div>
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </Spin>
    </div>
  )
}
