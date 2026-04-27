import { useMemo } from 'react'
import { Alert, Button, Card, Input, Select, Space, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'

import { useApiDebugger } from '@/hooks'
import type { ApiDebuggerProps } from '@/types/presentation'

const HTTP_METHOD_OPTIONS = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'PATCH', value: 'PATCH' },
  { label: 'DELETE', value: 'DELETE' },
] as const

export default function ApiDebugger({
  accessUrl,
  title = 'API 调试器',
  description = '用于向当前实验环境发送 REST 或 JSON-RPC 请求',
}: ApiDebuggerProps) {
  const {
    mode,
    method,
    path,
    jsonRpcMethod,
    headers,
    body,
    responseState,
    history,
    loading,
    setMode,
    setMethod,
    setPath,
    setJsonRpcMethod,
    setHeaders,
    setBody,
    sendRequest,
  } = useApiDebugger(accessUrl)

  const historyColumns = useMemo<ColumnsType<(typeof history)[number]>>(() => [
    { title: '模式', dataIndex: 'mode', key: 'mode', width: 90 },
    { title: '方法', dataIndex: 'method', key: 'method', width: 90 },
    { title: '路径/方法名', dataIndex: 'path', key: 'path' },
    {
      title: '结果',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (value: string) => (
        <Tag color={value.startsWith('2') ? 'success' : value.startsWith('ERR') ? 'error' : 'processing'}>
          {value}
        </Tag>
      ),
    },
  ], [history])

  const handleSend = async () => {
    try {
      await sendRequest()
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '未知错误'
      if (errorMessage.includes('JSON')) {
        message.error(errorMessage)
      }
    }
  }

  return (
    <div className="h-full overflow-auto bg-slate-50 p-4 text-slate-900">
      <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <Card
          title={title}
          className="border-slate-200 bg-white"
          headStyle={{ color: '#0f172a', borderBottomColor: '#e2e8f0' }}
          bodyStyle={{ color: '#0f172a' }}
        >
          <Alert
            type="info"
            showIcon
            className="mb-4"
            message={description}
          />

          <div className="space-y-4">
            <Space wrap>
              <Select
                value={mode}
                onChange={(value) => setMode(value)}
                options={[
                  { label: 'JSON-RPC', value: 'jsonrpc' },
                  { label: 'REST', value: 'rest' },
                ]}
                style={{ width: 140 }}
              />
              <Select
                value={method}
                onChange={(value) => setMethod(value)}
                options={HTTP_METHOD_OPTIONS as unknown as Array<{ label: string; value: string }>}
                style={{ width: 120 }}
              />
              {mode === 'jsonrpc' ? (
                <Input
                  value={jsonRpcMethod}
                  onChange={(event) => setJsonRpcMethod(event.target.value)}
                  placeholder="eth_blockNumber"
                />
              ) : (
                <Input
                  value={path}
                  onChange={(event) => setPath(event.target.value)}
                  placeholder="/"
                />
              )}
            </Space>

            <div>
              <Typography.Text className="mb-2 block text-slate-600">请求头</Typography.Text>
              <Input.TextArea rows={5} value={headers} onChange={(event) => setHeaders(event.target.value)} />
            </div>

            <div>
              <Typography.Text className="mb-2 block text-slate-600">请求体</Typography.Text>
              <Input.TextArea rows={10} value={body} onChange={(event) => setBody(event.target.value)} />
            </div>

            <Button type="primary" loading={loading} onClick={() => void handleSend()}>
              发送请求
            </Button>
          </div>
        </Card>

        <div className="space-y-4">
          <Card
            title="响应结果"
            className="border-slate-200 bg-white"
            headStyle={{ color: '#0f172a', borderBottomColor: '#e2e8f0' }}
            bodyStyle={{ color: '#0f172a' }}
          >
            <div className="mb-2 text-sm text-slate-600">{responseState.statusLine}</div>
            <Input.TextArea rows={18} value={responseState.body} readOnly />
          </Card>

          <Card
            title="最近请求"
            className="border-slate-200 bg-white"
            headStyle={{ color: '#0f172a', borderBottomColor: '#e2e8f0' }}
            bodyStyle={{ color: '#0f172a' }}
          >
            <Table rowKey="id" size="small" pagination={false} columns={historyColumns} dataSource={history} />
          </Card>
        </div>
      </div>
    </div>
  )
}
