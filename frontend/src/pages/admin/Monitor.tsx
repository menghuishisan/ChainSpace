/**
 * 平台管理员 - 系统监控页面
 */
import { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Statistic, Progress, Table, Tag, Space, Button, Alert, Spin } from 'antd'
import { 
  CloudServerOutlined, DatabaseOutlined, ClusterOutlined, 
  ReloadOutlined, WarningOutlined, CheckCircleOutlined
} from '@ant-design/icons'
import { PageHeader } from '@/components/common'
import { getSystemMonitor, getContainerStats, getServiceHealth } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { SystemMonitor, ContainerStats, ServiceHealth } from '@/types'

export default function AdminMonitor() {
  const [loading, setLoading] = useState(true)
  const [systemStats, setSystemStats] = useState<SystemMonitor | null>(null)
  const [containerStats, setContainerStats] = useState<ContainerStats | null>(null)
  const [serviceHealth, setServiceHealth] = useState<ServiceHealth[]>([])
  const [refreshing, setRefreshing] = useState(false)

  // 获取数据
  const fetchData = useCallback(async () => {
    try {
      const [system, containers, services] = await Promise.all([
        getSystemMonitor(),
        getContainerStats(),
        getServiceHealth(),
      ])
      setSystemStats(system)
      setContainerStats(containers)
      setServiceHealth(services)
    } catch { /* */ } finally { setLoading(false); setRefreshing(false) }
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  // 刷新
  const handleRefresh = () => {
    setRefreshing(true)
    fetchData()
  }

  // 格式化运行时间
  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    return `${days}天 ${hours}小时 ${mins}分钟`
  }

  // 服务状态表格列
  const serviceColumns = [
    { title: '服务名称', dataIndex: 'name', key: 'name' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: string) => (
      <Tag color={v === 'healthy' ? 'success' : v === 'unhealthy' ? 'error' : 'default'} icon={v === 'healthy' ? <CheckCircleOutlined /> : <WarningOutlined />}>
        {v === 'healthy' ? '健康' : v === 'unhealthy' ? '异常' : '未知'}
      </Tag>
    )},
    { title: '响应延迟', dataIndex: 'latency', key: 'latency', render: (v?: number) => v ? `${v}ms` : '-' },
    { title: '最后检查', dataIndex: 'last_check', key: 'last_check', render: (v: string) => formatDateTime(v) },
    { title: '备注', dataIndex: 'message', key: 'message', render: (v?: string) => v || '-' },
  ]

  // 容器表格列
  const containerColumns = [
    { title: '容器名', dataIndex: 'name', key: 'name', render: (v: string) => <span className="font-mono text-xs">{v}</span> },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: string) => (
      <Tag color={v === 'running' ? 'success' : v === 'paused' ? 'warning' : 'default'}>{v}</Tag>
    )},
    { title: 'CPU', dataIndex: 'cpu_percent', key: 'cpu_percent', render: (v: number) => `${v.toFixed(1)}%` },
    { title: '内存', dataIndex: 'memory_usage', key: 'memory_usage' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (v: string) => formatDateTime(v) },
  ]

  if (loading) return <div className="flex items-center justify-center h-64"><Spin size="large" /></div>

  const hasUnhealthyService = serviceHealth.some(s => s.status === 'unhealthy')

  return (
    <div>
      <PageHeader 
        title="系统监控" 
        subtitle="查看平台运行状态和资源使用情况" 
        extra={<Button icon={<ReloadOutlined />} loading={refreshing} onClick={handleRefresh}>刷新</Button>}
      />

      {hasUnhealthyService && (
        <Alert message="部分服务异常，请及时处理" type="warning" showIcon className="mb-4" />
      )}

      {/* 系统资源 */}
      <Row gutter={16} className="mb-6">
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="CPU使用率" value={systemStats?.cpu_usage || 0} precision={1} suffix="%" prefix={<CloudServerOutlined />} />
            <Progress percent={systemStats?.cpu_usage || 0} showInfo={false} status={systemStats?.cpu_usage && systemStats.cpu_usage > 80 ? 'exception' : 'normal'} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="内存使用" value={systemStats?.memory_usage || 0} suffix="%" prefix={<DatabaseOutlined />} />
            <Progress percent={systemStats?.memory_usage || 0} showInfo={false} status={systemStats?.memory_usage && systemStats.memory_usage > 80 ? 'exception' : 'normal'} />
            <p className="text-xs text-text-secondary mt-2">{systemStats?.memory_used} / {systemStats?.memory_total}</p>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="磁盘使用" value={systemStats?.disk_usage || 0} precision={1} suffix="%" prefix={<DatabaseOutlined />} />
            <Progress percent={systemStats?.disk_usage || 0} showInfo={false} status={systemStats?.disk_usage && systemStats.disk_usage > 80 ? 'exception' : 'normal'} />
            <p className="text-xs text-text-secondary mt-2">{systemStats?.disk_used} / {systemStats?.disk_total}</p>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="系统运行时间" value={formatUptime(systemStats?.uptime || 0)} prefix={<ClusterOutlined />} />
            <p className="text-xs text-text-secondary mt-2">负载：{systemStats?.load_average?.join(' / ') || '-'}</p>
          </Card>
        </Col>
      </Row>

      {/* 服务健康状态 */}
      <Card title="服务健康状态" className="mb-6">
        <Table columns={serviceColumns} dataSource={serviceHealth} rowKey="name" pagination={false} size="small" />
      </Card>

      {/* 容器状态 */}
      <Card 
        title="容器状态" 
        extra={
          <Space>
            <Tag color="success">运行中: {containerStats?.running || 0}</Tag>
            <Tag color="warning">已暂停: {containerStats?.paused || 0}</Tag>
            <Tag>已停止: {containerStats?.stopped || 0}</Tag>
          </Space>
        }
      >
        <Table 
          columns={containerColumns} 
          dataSource={containerStats?.containers || []} 
          rowKey="id" 
          pagination={{ pageSize: 10 }} 
          size="small"
          scroll={{ x: 800 }}
        />
      </Card>
    </div>
  )
}
