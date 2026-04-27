/**
 * 学校管理员 - 本校统计页面
 * 使用后端 /system/stats API 获取统计数据
 */
import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Spin } from 'antd'
import { 
  TeamOutlined, BookOutlined, ExperimentOutlined,
  TrophyOutlined, CloudServerOutlined, GlobalOutlined
} from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import type { SystemStats } from '@/types'
import { getSystemStats } from '@/api/admin'

export default function SchoolStatistics() {
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

  return (
    <div>
      <PageHeader title="本校统计" subtitle="查看学校的教学统计数据" />

      <Spin spinning={loading}>
        <Row gutter={[24, 24]}>
          <Col xs={24} sm={12} lg={4}>
            <Card>
              <Statistic
                title="用户总数"
                value={stats?.total_users || 0}
                prefix={<TeamOutlined style={{ color: '#1890FF' }} />}
                valueStyle={{ color: '#1890FF' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={4}>
            <Card>
              <Statistic
                title="学校数量"
                value={stats?.total_schools || 0}
                prefix={<GlobalOutlined style={{ color: '#52C41A' }} />}
                valueStyle={{ color: '#52C41A' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={4}>
            <Card>
              <Statistic
                title="课程数量"
                value={stats?.total_courses || 0}
                prefix={<BookOutlined style={{ color: '#722ED1' }} />}
                valueStyle={{ color: '#722ED1' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={4}>
            <Card>
              <Statistic
                title="实验数量"
                value={stats?.total_experiments || 0}
                prefix={<ExperimentOutlined style={{ color: '#FA8C16' }} />}
                valueStyle={{ color: '#FA8C16' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={4}>
            <Card>
              <Statistic
                title="竞赛数量"
                value={stats?.total_contests || 0}
                prefix={<TrophyOutlined style={{ color: '#EB2F96' }} />}
                valueStyle={{ color: '#EB2F96' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={4}>
            <Card>
              <Statistic
                title="运行中环境"
                value={stats?.active_envs || 0}
                prefix={<CloudServerOutlined style={{ color: '#13C2C2' }} />}
                valueStyle={{ color: '#13C2C2' }}
              />
            </Card>
          </Col>
        </Row>
      </Spin>
    </div>
  )
}
