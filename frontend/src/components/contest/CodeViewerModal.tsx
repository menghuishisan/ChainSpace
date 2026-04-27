/**
 * 合约代码查看器弹窗。
 * 使用 Monaco Editor 提供 Solidity 语法高亮，支持复制和下载。
 */
import { Button, Modal, Space } from 'antd'
import {
  CopyOutlined,
  DownloadOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
} from '@ant-design/icons'
import Editor, { OnMount } from '@monaco-editor/react'
import { useCallback, useRef, useState } from 'react'

interface CodeViewerModalProps {
  open: boolean
  code: string
  title?: string
  onClose: () => void
}

function downloadTextFile(filename: string, content: string) {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function extractContractName(code: string): string {
  const match = code.match(/contract\s+(\w+)/)
  return match ? match[1] : 'contract'
}

function detectLanguage(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'sol':
      return 'sol'
    case 'js':
    case 'mjs':
      return 'javascript'
    case 'ts':
    case 'mts':
      return 'typescript'
    case 'py':
      return 'python'
    case 'json':
      return 'json'
    case 'yaml':
    case 'yml':
      return 'yaml'
    case 'md':
      return 'markdown'
    case 'sh':
      return 'shell'
    case 'go':
      return 'go'
    case 'rs':
      return 'rust'
    case 'java':
      return 'java'
    case 'cpp':
    case 'cc':
    case 'cxx':
      return 'cpp'
    case 'c':
    case 'h':
      return 'c'
    case 'sql':
      return 'sql'
    default:
      return 'plaintext'
  }
}

export default function CodeViewerModal({ open, code, title, onClose }: CodeViewerModalProps) {
  const editorRef = useRef<Parameters<OnMount>[0] | null>(null)
  const [fullscreen, setFullscreen] = useState(false)

  const handleEditorMount: OnMount = useCallback((editor) => {
    editorRef.current = editor
  }, [])

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(code)
    } catch {
      // fallback
      const textarea = document.createElement('textarea')
      textarea.value = code
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
  }, [code])

  const handleDownload = useCallback(() => {
    const filename = `${extractContractName(code)}.sol`
    downloadTextFile(filename, code)
  }, [code])

  const handleFullscreen = useCallback(() => {
    const container = document.getElementById('code-viewer-editor-container')
    if (!container) return

    if (!fullscreen) {
      container.requestFullscreen?.()
    } else {
      document.exitFullscreen?.()
    }
    setFullscreen(!fullscreen)
  }, [fullscreen])

  const contractName = extractContractName(code)
  const filename = `${contractName}.sol`
  const language = detectLanguage(filename)

  return (
    <Modal
      title={title || `查看代码 - ${filename}`}
      open={open}
      onCancel={onClose}
      footer={null}
      width={fullscreen ? '100vw' : 900}
      style={{ top: fullscreen ? 0 : 40 }}
      bodyStyle={{ padding: 0, height: fullscreen ? 'calc(100vh - 56px)' : 560 }}
      destroyOnClose
    >
      <div className="flex items-center justify-between border-b border-gray-200 bg-gray-50 px-4 py-2">
        <span className="font-mono text-sm text-gray-600">{filename}</span>
        <Space>
          <Button
            type="text"
            size="small"
            icon={fullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
            onClick={handleFullscreen}
            title={fullscreen ? '退出全屏' : '全屏'}
          />
          <Button type="text" size="small" icon={<CopyOutlined />} onClick={handleCopy}>
            复制
          </Button>
          <Button type="text" size="small" icon={<DownloadOutlined />} onClick={handleDownload}>
            下载
          </Button>
        </Space>
      </div>

      <div id="code-viewer-editor-container" className="h-[calc(100%-44px)]">
        <Editor
          height="100%"
          language={language}
          value={code}
          theme="vs-dark"
          onMount={handleEditorMount}
          options={{
            readOnly: true,
            minimap: { enabled: true },
            fontSize: 13,
            fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, monospace",
            fontLigatures: true,
            lineNumbers: 'on',
            renderLineHighlight: 'line',
            scrollBeyondLastLine: false,
            wordWrap: 'on',
            folding: true,
            automaticLayout: true,
            padding: { top: 12 },
            scrollbar: {
              verticalScrollbarSize: 8,
              horizontalScrollbarSize: 8,
            },
          }}
        />
      </div>
    </Modal>
  )
}
