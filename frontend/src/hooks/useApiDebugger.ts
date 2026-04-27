import { useCallback, useMemo, useState } from 'react'

import { sendRuntimeRequest } from '@/api/experimentRuntime'

export type RequestMode = 'jsonrpc' | 'rest'
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export interface DebugHistoryItem {
  id: string
  mode: RequestMode
  method: HttpMethod
  path: string
  status: string
  createdAt: string
}

export interface DebugResponseState {
  statusLine: string
  body: string
}

export function useApiDebugger(accessUrl?: string) {
  const [mode, setMode] = useState<RequestMode>('jsonrpc')
  const [method, setMethod] = useState<HttpMethod>('POST')
  const [path, setPath] = useState('/')
  const [jsonRpcMethod, setJsonRpcMethod] = useState('eth_blockNumber')
  const [headers, setHeaders] = useState('{\n  "Content-Type": "application/json"\n}')
  const [body, setBody] = useState('{\n  "jsonrpc": "2.0",\n  "method": "eth_blockNumber",\n  "params": [],\n  "id": 1\n}')
  const [responseState, setResponseState] = useState<DebugResponseState>({
    statusLine: '尚未发送请求',
    body: '',
  })
  const [history, setHistory] = useState<DebugHistoryItem[]>([])
  const [loading, setLoading] = useState(false)

  const requestTarget = useMemo(() => (
    mode === 'jsonrpc' ? '/' : (path.startsWith('/') ? path : `/${path}`)
  ), [mode, path])

  const parseHeaders = useCallback(() => {
    if (!headers.trim()) {
      return {}
    }

    return JSON.parse(headers) as Record<string, string>
  }, [headers])

  const sendRequest = useCallback(async () => {
    if (!accessUrl) {
      throw new Error('当前没有可用的 API 调试入口')
    }

    const parsedHeaders = parseHeaders()
    let requestBody = body

    if (mode === 'jsonrpc') {
      const parsedBody = JSON.parse(body) as Record<string, unknown>
      parsedBody.method = jsonRpcMethod.trim() || parsedBody.method
      requestBody = JSON.stringify(parsedBody)
    }

    setLoading(true)
    try {
      const response = await sendRuntimeRequest(accessUrl, {
        path: requestTarget,
        method,
        headers: parsedHeaders,
        body: method === 'GET' || method === 'DELETE' ? undefined : requestBody,
      })

      setResponseState({
        statusLine: `${response.status} ${response.statusText}`,
        body: response.body,
      })
      setHistory((current) => [
        {
          id: `${Date.now()}`,
          mode,
          method,
          path: mode === 'jsonrpc' ? jsonRpcMethod : requestTarget,
          status: `${response.status}`,
          createdAt: new Date().toISOString(),
        },
        ...current,
      ].slice(0, 8))
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '未知错误'
      setResponseState({
        statusLine: 'ERR Request Failed',
        body: errorMessage,
      })
      setHistory((current) => [
        {
          id: `${Date.now()}`,
          mode,
          method,
          path: mode === 'jsonrpc' ? jsonRpcMethod : requestTarget,
          status: 'ERR',
          createdAt: new Date().toISOString(),
        },
        ...current,
      ].slice(0, 8))
      throw error
    } finally {
      setLoading(false)
    }
  }, [accessUrl, body, jsonRpcMethod, method, mode, parseHeaders, requestTarget])

  return {
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
  }
}
