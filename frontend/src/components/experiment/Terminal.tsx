/**
 * 实验环境 - 终端组件
 * 基于xterm.js实现WebSocket终端连接
 */
import { useEffect, useRef, useState } from 'react'
import { Spin, Alert, Button } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import type { TerminalConnectionStatus } from '@/types/presentation'
import type { TerminalProps } from '@/types/presentation'

export default function Terminal({ experimentId, accessUrl, wsUrl }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<unknown>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const [status, setStatus] = useState<TerminalConnectionStatus>('connecting')
  const [errorMsg, setErrorMsg] = useState('')
  const [reconnectKey, setReconnectKey] = useState(0)

  useEffect(() => {
    let term: unknown = null
    let fitAddon: unknown = null

    const initTerminal = async () => {
      if (!terminalRef.current) return

      try {
        // 动态导入xterm（避免SSR问题）
        const { Terminal } = await import('xterm')
        const { FitAddon } = await import('xterm-addon-fit')
        // xterm CSS will be loaded by the xterm package

        // 创建终端实例
        term = new Terminal({
          cursorBlink: true,
          fontSize: 14,
          fontFamily: 'Consolas, "Courier New", monospace',
          theme: {
            background: '#1a1a2e',
            foreground: '#e0e0e0',
            cursor: '#4ecdc4',
            cursorAccent: '#1a1a2e',
            selectionBackground: 'rgba(78, 205, 196, 0.3)',
            black: '#000000',
            red: '#ff6b6b',
            green: '#4ecdc4',
            yellow: '#ffd93d',
            blue: '#6c5ce7',
            magenta: '#a29bfe',
            cyan: '#74b9ff',
            white: '#e0e0e0',
          },
          allowTransparency: true,
        })

        fitAddon = new FitAddon()
        ;(term as { loadAddon: (addon: unknown) => void }).loadAddon(fitAddon)
        ;(term as { open: (container: HTMLElement) => void }).open(terminalRef.current)
        ;(fitAddon as { fit: () => void }).fit()

        xtermRef.current = term

        // 连接WebSocket
        const wsEndpoint = wsUrl || (accessUrl ? `${accessUrl.replace('http', 'ws')}/ws/terminal` : null)
        
        if (wsEndpoint) {
          connectWebSocket(wsEndpoint, term as { write: (data: string) => void; onData: (cb: (data: string) => void) => void })
        } else {
          // 无连接时显示错误状态
          setStatus('error')
          setErrorMsg('实验环境未就绪，无法连接终端')
        }

        // 监听窗口大小变化
        const handleResize = () => {
          ;(fitAddon as { fit: () => void })?.fit()
        }
        window.addEventListener('resize', handleResize)

        return () => {
          window.removeEventListener('resize', handleResize)
        }
      } catch (error) {
        console.error('Terminal init error:', error)
        setStatus('error')
        setErrorMsg('终端初始化失败')
      }
    }

    // WebSocket连接
    const connectWebSocket = (url: string, t: { write: (data: string) => void; onData: (cb: (data: string) => void) => void }) => {
      setStatus('connecting')
      
      try {
        const ws = new WebSocket(url)
        wsRef.current = ws

        ws.onopen = () => {
          setStatus('connected')
          // 发送实验ID进行认证
          ws.send(JSON.stringify({ type: 'auth', experimentId }))
        }

        ws.onmessage = (event) => {
          t.write(event.data)
        }

        ws.onclose = () => {
          setStatus('disconnected')
        }

        ws.onerror = () => {
          setStatus('error')
          setErrorMsg('WebSocket连接失败')
        }

        // 发送输入到服务器
        t.onData((data: string) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'input', data }))
          }
        })
      } catch {
        setStatus('error')
        setErrorMsg('无法建立WebSocket连接')
      }
    }

    initTerminal()

    return () => {
      wsRef.current?.close()
      if (xtermRef.current) {
        ;(xtermRef.current as { dispose: () => void }).dispose()
      }
    }
  }, [experimentId, accessUrl, wsUrl, reconnectKey])

  // 重新连接
  const handleReconnect = () => {
    setStatus('connecting')
    setErrorMsg('')
    setReconnectKey(prev => prev + 1)
  }

  if (status === 'error') {
    return (
      <div className="h-full flex items-center justify-center bg-gray-900">
        <Alert
          type="error"
          message="终端连接失败"
          description={errorMsg || '无法连接到实验环境终端'}
          action={
            <Button size="small" icon={<ReloadOutlined />} onClick={handleReconnect}>
              重试
            </Button>
          }
        />
      </div>
    )
  }

  return (
    <div className="h-full bg-gray-900 relative flex flex-col">
            
      {status === 'connecting' && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-900 z-10">
          <Spin tip="正在连接终端..."><div /></Spin>
        </div>
      )}
      {status === 'disconnected' && (
        <div className="absolute top-2 right-2 z-10">
          <Button size="small" type="primary" icon={<ReloadOutlined />} onClick={handleReconnect}>
            重新连接
          </Button>
        </div>
      )}
      <div ref={terminalRef} className="flex-1 w-full p-2" />
    </div>
  )
}
