import { useMemo } from 'react'
import { Alert, Button, Input, Select, Space, Switch, Tag } from 'antd'
import {
  DeleteOutlined,
  DownloadOutlined,
  ReloadOutlined,
  SearchOutlined,
  VerticalAlignBottomOutlined,
} from '@ant-design/icons'

import { useRuntimeLogs } from '@/hooks'
import type { LogViewerProps } from '@/types/presentation'

const levelTags: Record<string, string> = {
  info: 'blue',
  warn: 'orange',
  error: 'red',
  debug: 'default',
}

export default function LogViewer({ accessUrl, sources = [] }: LogViewerProps) {
  const {
    loading,
    connected,
    selectedSource,
    levelFilter,
    searchText,
    autoScroll,
    containerRef,
    filteredLogs,
    setLogs,
    setSelectedSource,
    setLevelFilter,
    setSearchText,
    setAutoScroll,
    refreshLogs,
  } = useRuntimeLogs(accessUrl)

  const availableSources = useMemo(() => sources.map((source) => ({ label: source, value: source })), [sources])

  const handleClear = () => {
    setLogs([])
  }

  const handleExport = () => {
    if (filteredLogs.length === 0) {
      return
    }

    const content = filteredLogs
      .map((log) => `${log.timestamp} [${log.level.toUpperCase()}] ${log.source}: ${log.message}`)
      .join('\n')

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `logs_${new Date().toISOString().slice(0, 10)}.txt`
    link.click()
    URL.revokeObjectURL(url)
  }

  const formatTime = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleTimeString('zh-CN', { hour12: false })
    } catch {
      return timestamp
    }
  }

  if (!accessUrl) {
    return (
      <div className="flex h-full items-center justify-center bg-slate-50">
        <Alert
          type="warning"
          message="日志查看器不可用"
          description="实验环境未就绪，暂时无法获取日志。"
          showIcon
        />
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col bg-white text-slate-900">
      <div className="flex items-center justify-between border-b border-slate-200 bg-slate-50 p-3">
        <Space>
          <Select
            placeholder="日志来源"
            allowClear
            value={selectedSource || undefined}
            onChange={(value) => setSelectedSource(value || '')}
            style={{ width: 140 }}
            options={availableSources}
          />
          <Select
            placeholder="日志级别"
            mode="multiple"
            allowClear
            value={levelFilter}
            onChange={setLevelFilter}
            style={{ width: 180 }}
            options={[
              { label: 'INFO', value: 'info' },
              { label: 'WARN', value: 'warn' },
              { label: 'ERROR', value: 'error' },
              { label: 'DEBUG', value: 'debug' },
            ]}
          />
          <Input
            placeholder="搜索日志..."
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(event) => setSearchText(event.target.value)}
            style={{ width: 220 }}
            allowClear
          />
        </Space>

        <Space>
          <span className="flex items-center gap-1 text-sm text-text-secondary">
            <VerticalAlignBottomOutlined />
            自动滚动
          </span>
          <Switch size="small" checked={autoScroll} onChange={setAutoScroll} />
          <Button icon={<ReloadOutlined />} onClick={() => void refreshLogs()} loading={loading}>
            刷新
          </Button>
          <Button icon={<DeleteOutlined />} onClick={handleClear}>
            清空
          </Button>
          <Button icon={<DownloadOutlined />} onClick={handleExport} disabled={filteredLogs.length === 0}>
            导出
          </Button>
        </Space>
      </div>

      <div ref={containerRef} className="flex-1 overflow-auto bg-white p-3 font-mono text-sm">
        {!connected && filteredLogs.length === 0 ? (
          <div className="flex h-full items-center justify-center">
            <Alert
              type="info"
              message="等待日志服务"
              description="正在连接实验环境日志接口..."
              showIcon
            />
          </div>
        ) : filteredLogs.length === 0 ? (
          <div className="flex h-full items-center justify-center text-text-secondary">
            当前筛选条件下没有日志记录。
          </div>
        ) : (
          filteredLogs.map((log) => (
            <div
              key={log.id}
              className="mb-1 grid grid-cols-[72px_72px_120px_1fr] gap-3 rounded px-2 py-1 hover:bg-slate-100"
            >
              <span className="text-text-secondary">{formatTime(log.timestamp)}</span>
              <Tag color={levelTags[log.level]} className="m-0 w-fit">
                {log.level.toUpperCase()}
              </Tag>
              <span className="truncate text-primary">{log.source}</span>
              <span className="break-all">{log.message}</span>
            </div>
          ))
        )}
      </div>

      <div className="flex items-center justify-between border-t border-slate-200 bg-slate-50 px-3 py-2 text-xs text-text-secondary">
        <span>共 {filteredLogs.length} 条日志</span>
        <span>{connected ? '已连接 · 自动轮询中' : '未连接 · 等待重试'}</span>
      </div>
    </div>
  )
}
