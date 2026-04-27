import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Descriptions,
  Drawer,
  Input,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'

import { sendRuntimeJsonRpc } from '@/api/experimentRuntime'
import type { BlockExplorerProps } from '@/types/presentation'

interface RpcTransaction {
  hash: string
  from?: string
  to?: string | null
  value?: string
  gas?: string
  gasPrice?: string
  nonce?: string
  input?: string
}

interface RpcBlock {
  number?: string
  hash?: string
  parentHash?: string
  timestamp?: string
  miner?: string
  gasLimit?: string
  gasUsed?: string
  size?: string
  transactions?: Array<RpcTransaction | string>
}

interface BlockRow {
  key: string
  numberText: string
  hash: string
  timestamp: string
  txCount: number
  miner: string
}

function hexToNumber(value?: string): number {
  if (!value) {
    return 0
  }
  try {
    return Number(BigInt(value))
  } catch {
    return 0
  }
}

function formatHexNumber(value?: string): string {
  if (!value) {
    return '--'
  }
  const parsed = hexToNumber(value)
  return Number.isFinite(parsed) ? String(parsed) : value
}

function formatTimestamp(value?: string): string {
  if (!value) {
    return '--'
  }
  const seconds = hexToNumber(value)
  if (!seconds) {
    return '--'
  }
  return new Date(seconds * 1000).toLocaleString('zh-CN')
}

function shortHash(value?: string, head = 10, tail = 8): string {
  if (!value) {
    return '--'
  }
  if (value.length <= head + tail + 3) {
    return value
  }
  return `${value.slice(0, head)}...${value.slice(-tail)}`
}

function normalizeBlockTag(input: string): string {
  const trimmed = input.trim()
  if (!trimmed) {
    return ''
  }
  if (trimmed.startsWith('0x')) {
    return trimmed
  }
  const asNumber = Number(trimmed)
  if (Number.isNaN(asNumber) || asNumber < 0) {
    return ''
  }
  return `0x${asNumber.toString(16)}`
}

export default function BlockExplorer({ accessUrl }: BlockExplorerProps) {
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [latestBlockNumber, setLatestBlockNumber] = useState<string>('')
  const [recentBlocks, setRecentBlocks] = useState<RpcBlock[]>([])
  const [selectedBlock, setSelectedBlock] = useState<RpcBlock | null>(null)
  const [selectedTransaction, setSelectedTransaction] = useState<RpcTransaction | null>(null)
  const [query, setQuery] = useState('')
  const [error, setError] = useState<string>('')

  const fetchBlockByNumber = useCallback(async (blockNumberHex: string) => (
    sendRuntimeJsonRpc<RpcBlock | null>(accessUrl!, 'eth_getBlockByNumber', [blockNumberHex, true])
  ), [accessUrl])

  const fetchLatestBlocks = useCallback(async () => {
    if (!accessUrl) {
      return
    }

    setError('')
    const latest = await sendRuntimeJsonRpc<string>(accessUrl, 'eth_blockNumber')
    const latestNumber = hexToNumber(latest)
    const requests: Promise<RpcBlock | null>[] = []
    const maxBlocks = Math.min(latestNumber + 1, 10)
    for (let offset = 0; offset < maxBlocks; offset += 1) {
      requests.push(fetchBlockByNumber(`0x${(latestNumber - offset).toString(16)}`))
    }

    const blocks = (await Promise.all(requests)).filter((item): item is RpcBlock => Boolean(item))
    setLatestBlockNumber(latest)
    setRecentBlocks(blocks)
    setSelectedBlock((current) => current || blocks[0] || null)
  }, [accessUrl, fetchBlockByNumber])

  const handleRefresh = useCallback(async () => {
    if (!accessUrl) {
      return
    }
    setRefreshing(true)
    try {
      await fetchLatestBlocks()
    } catch (rpcError) {
      const errorMessage = rpcError instanceof Error ? rpcError.message : '区块链节点暂时不可用'
      setError(errorMessage)
    } finally {
      setRefreshing(false)
      setLoading(false)
    }
  }, [accessUrl, fetchLatestBlocks])

  useEffect(() => {
    let active = true

    if (!accessUrl) {
      setLoading(false)
      return () => {
        active = false
      }
    }

    void (async () => {
      try {
        await fetchLatestBlocks()
        if (!active) {
          return
        }
      } catch (rpcError) {
        if (!active) {
          return
        }
        const errorMessage = rpcError instanceof Error ? rpcError.message : '区块链节点暂时不可用'
        setError(errorMessage)
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    })()

    return () => {
      active = false
    }
  }, [accessUrl, fetchLatestBlocks])

  const blockRows = useMemo<BlockRow[]>(() => recentBlocks.map((block, index) => ({
    key: block.hash || block.number || `${index}`,
    numberText: formatHexNumber(block.number),
    hash: block.hash || '--',
    timestamp: formatTimestamp(block.timestamp),
    txCount: block.transactions?.length || 0,
    miner: block.miner || '--',
  })), [recentBlocks])

  const blockColumns = useMemo<ColumnsType<BlockRow>>(() => [
    {
      title: '区块号',
      dataIndex: 'numberText',
      key: 'numberText',
      width: 110,
      render: (value: string) => <Tag color="processing">#{value}</Tag>,
    },
    {
      title: '区块哈希',
      dataIndex: 'hash',
      key: 'hash',
      render: (value: string) => <Typography.Text code>{shortHash(value)}</Typography.Text>,
    },
    { title: '交易数', dataIndex: 'txCount', key: 'txCount', width: 90 },
    { title: '出块时间', dataIndex: 'timestamp', key: 'timestamp', width: 180 },
    {
      title: '矿工/节点',
      dataIndex: 'miner',
      key: 'miner',
      render: (value: string) => <Typography.Text ellipsis>{shortHash(value, 12, 6)}</Typography.Text>,
    },
  ], [])

  const transactionRows = useMemo(() => (
    (selectedBlock?.transactions || []).map((item, index) => {
      const tx = typeof item === 'string' ? { hash: item } : item
      return {
        key: tx.hash || `${index}`,
        ...tx,
      }
    })
  ), [selectedBlock])

  const handleSearch = useCallback(async () => {
    if (!accessUrl) {
      return
    }

    const trimmed = query.trim()
    if (!trimmed) {
      message.warning('请输入区块号、区块哈希或交易哈希')
      return
    }

    setRefreshing(true)
    setError('')
    try {
      let block: RpcBlock | null = null
      const normalizedBlockTag = normalizeBlockTag(trimmed)

      if (normalizedBlockTag) {
        block = await fetchBlockByNumber(normalizedBlockTag)
      } else if (trimmed.startsWith('0x')) {
        block = await sendRuntimeJsonRpc<RpcBlock | null>(accessUrl, 'eth_getBlockByHash', [trimmed, true])
        if (!block) {
          const tx = await sendRuntimeJsonRpc<RpcTransaction | null>(accessUrl, 'eth_getTransactionByHash', [trimmed])
          if (tx) {
            setSelectedTransaction(tx)
            message.success('已定位到交易详情')
            return
          }
        }
      }

      if (!block) {
        message.warning('未找到对应的区块或交易')
        return
      }

      setSelectedBlock(block)
      if (block.hash && !recentBlocks.some((item) => item.hash === block.hash)) {
        setRecentBlocks((current) => [block, ...current].slice(0, 10))
      }
    } catch (rpcError) {
      const errorMessage = rpcError instanceof Error ? rpcError.message : '查询失败'
      setError(errorMessage)
    } finally {
      setRefreshing(false)
    }
  }, [accessUrl, fetchBlockByNumber, query, recentBlocks])

  if (!accessUrl) {
    return (
      <div className="flex h-full items-center justify-center bg-slate-50">
        <Alert
          type="warning"
          message="区块浏览器不可用"
          description="当前实验没有可用的链上 RPC 入口，因此无法加载原生浏览器。"
          showIcon
        />
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center bg-slate-50">
        <Spin size="large" tip="正在连接链上节点并加载区块数据..." />
      </div>
    )
  }

  return (
    <div className="h-full overflow-auto bg-slate-50 p-4 text-slate-900">
      <div className="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        <div className="space-y-4">
          <Card
            title="链上原生区块浏览器"
            extra={(
              <Space>
                <Input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  onPressEnter={() => void handleSearch()}
                  placeholder="输入区块号 / 区块哈希 / 交易哈希"
                  suffix={<SearchOutlined />}
                  style={{ width: 280 }}
                />
                <Button onClick={() => void handleSearch()} loading={refreshing}>查询</Button>
                <Button icon={<ReloadOutlined />} onClick={() => void handleRefresh()} loading={refreshing}>刷新</Button>
              </Space>
            )}
            className="border-slate-200 bg-white"
            headStyle={{ color: '#0f172a', borderBottomColor: '#e2e8f0' }}
            bodyStyle={{ color: '#0f172a' }}
          >
            <Alert
              type="info"
              showIcon
              className="mb-4"
              message="浏览器直接通过实验环境代理访问链节点 JSON-RPC。"
            />
            {error ? (
              <Alert type="error" showIcon className="mb-4" message="链上接口异常" description={error} />
            ) : null}

            <Descriptions size="small" bordered column={2}>
              <Descriptions.Item label="最新区块">{formatHexNumber(latestBlockNumber)}</Descriptions.Item>
              <Descriptions.Item label="最近同步状态">
                <Tag color={error ? 'error' : 'success'}>{error ? '异常' : '正常'}</Tag>
              </Descriptions.Item>
            </Descriptions>

            <div className="mt-4">
              <Table
                rowKey="key"
                size="small"
                pagination={false}
                columns={blockColumns}
                dataSource={blockRows}
                onRow={(record) => ({
                  onClick: () => {
                    const matched = recentBlocks.find((item) => item.hash === record.hash)
                    if (matched) {
                      setSelectedBlock(matched)
                    }
                  },
                })}
              />
            </div>
          </Card>
        </div>

        <Card
          title="区块详情"
          className="border-slate-200 bg-white"
          headStyle={{ color: '#0f172a', borderBottomColor: '#e2e8f0' }}
          bodyStyle={{ color: '#0f172a' }}
        >
          {selectedBlock ? (
            <>
              <Descriptions size="small" bordered column={1}>
                <Descriptions.Item label="区块号">#{formatHexNumber(selectedBlock.number)}</Descriptions.Item>
                <Descriptions.Item label="区块哈希">
                  <Typography.Text copyable>{selectedBlock.hash || '--'}</Typography.Text>
                </Descriptions.Item>
                <Descriptions.Item label="父区块哈希">
                  <Typography.Text copyable>{selectedBlock.parentHash || '--'}</Typography.Text>
                </Descriptions.Item>
                <Descriptions.Item label="出块时间">{formatTimestamp(selectedBlock.timestamp)}</Descriptions.Item>
                <Descriptions.Item label="矿工/节点">{selectedBlock.miner || '--'}</Descriptions.Item>
                <Descriptions.Item label="Gas Used / Limit">
                  {formatHexNumber(selectedBlock.gasUsed)} / {formatHexNumber(selectedBlock.gasLimit)}
                </Descriptions.Item>
                <Descriptions.Item label="区块大小">{formatHexNumber(selectedBlock.size)}</Descriptions.Item>
              </Descriptions>

              <div className="mt-4">
                <Typography.Title level={5}>交易列表</Typography.Title>
                <div className="space-y-2">
                  {transactionRows.length > 0 ? transactionRows.map((tx) => (
                    <button
                      type="button"
                      key={tx.key}
                      className="w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-left transition hover:border-primary hover:bg-sky-50"
                      onClick={() => setSelectedTransaction(tx)}
                    >
                      <div className="text-sm font-medium text-slate-900">{shortHash(tx.hash, 14, 10)}</div>
                      <div className="mt-1 text-xs text-slate-500">
                        {shortHash(tx.from, 12, 8)} → {shortHash(tx.to || '合约创建', 12, 8)}
                      </div>
                    </button>
                  )) : (
                    <Alert type="info" showIcon message="该区块没有可展示的交易详情。" />
                  )}
                </div>
              </div>
            </>
          ) : (
            <Alert type="info" showIcon message="请从左侧列表中选择一个区块。" />
          )}
        </Card>
      </div>

      <Drawer
        title="交易详情"
        width={640}
        open={Boolean(selectedTransaction)}
        onClose={() => setSelectedTransaction(null)}
      >
        {selectedTransaction ? (
          <Descriptions size="small" bordered column={1}>
            <Descriptions.Item label="交易哈希">
              <Typography.Text copyable>{selectedTransaction.hash}</Typography.Text>
            </Descriptions.Item>
            <Descriptions.Item label="From">{selectedTransaction.from || '--'}</Descriptions.Item>
            <Descriptions.Item label="To">{selectedTransaction.to || '合约创建'}</Descriptions.Item>
            <Descriptions.Item label="Nonce">{formatHexNumber(selectedTransaction.nonce)}</Descriptions.Item>
            <Descriptions.Item label="Value">{formatHexNumber(selectedTransaction.value)}</Descriptions.Item>
            <Descriptions.Item label="Gas">{formatHexNumber(selectedTransaction.gas)}</Descriptions.Item>
            <Descriptions.Item label="Gas Price">{formatHexNumber(selectedTransaction.gasPrice)}</Descriptions.Item>
            <Descriptions.Item label="Input">
              <Typography.Paragraph copyable className="!mb-0 break-all">
                {selectedTransaction.input || '--'}
              </Typography.Paragraph>
            </Descriptions.Item>
          </Descriptions>
        ) : null}
      </Drawer>
    </div>
  )
}
