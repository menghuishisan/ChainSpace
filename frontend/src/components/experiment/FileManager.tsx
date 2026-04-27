import { useState } from 'react'
import { Alert, Button, Dropdown, Modal, Space, Table, Upload, message } from 'antd'
import {
  ArrowLeftOutlined,
  DeleteOutlined,
  DownloadOutlined,
  ExclamationCircleOutlined,
  FileOutlined,
  FolderOutlined,
  MoreOutlined,
  PlusOutlined,
  ReloadOutlined,
  UploadOutlined,
} from '@ant-design/icons'
import type { MenuProps, UploadProps } from 'antd'
import type { ColumnsType } from 'antd/es/table'

import { useRuntimeFileManager } from '@/hooks'
import type { FileItem } from '@/types/presentation'
import type { FileManagerProps } from '@/types/presentation'
import { formatDateTime, formatFileSize } from '@/utils/format'

export default function FileManager({ accessUrl, onFileOpen }: FileManagerProps) {
  const {
    loading,
    files,
    currentPath,
    error,
    refreshFiles,
    enterDirectory,
    goBack,
    createDirectory,
    removeFiles,
  } = useRuntimeFileManager(accessUrl)
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])
  const [newFolderModalVisible, setNewFolderModalVisible] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')

  const handleEnterDirectory = (item: FileItem) => {
    if (item.type === 'directory') {
      enterDirectory(item.path)
      setSelectedKeys([])
      return
    }

    onFileOpen?.(item)
  }

  const handleCreateFolder = async () => {
    if (!newFolderName.trim()) {
      message.warning('请输入文件夹名称')
      return
    }

    try {
      await createDirectory(newFolderName.trim())
      message.success('文件夹创建成功')
      setNewFolderModalVisible(false)
      setNewFolderName('')
    } catch {
      message.error('创建文件夹失败')
    }
  }

  const handleDelete = async (items: FileItem[]) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除选中的 ${items.length} 项吗？`,
      onOk: async () => {
        try {
          await removeFiles(items.map((item) => item.path))
          message.success('删除成功')
          setSelectedKeys([])
        } catch {
          message.error('删除失败')
        }
      },
    })
  }

  const handleDownload = (item: FileItem) => {
    if (!accessUrl) {
      message.info('下载功能需要实验环境支持')
      return
    }

    window.open(`${accessUrl}/api/files/download?path=${encodeURIComponent(item.path)}`)
  }

  const uploadProps: UploadProps = {
    name: 'file',
    action: accessUrl ? `${accessUrl}/api/files/upload?path=${encodeURIComponent(currentPath)}` : undefined,
    showUploadList: false,
    onChange(info) {
      if (info.file.status === 'done') {
        message.success(`${info.file.name} 上传成功`)
        void refreshFiles()
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} 上传失败`)
      }
    },
  }

  const getActionMenu = (record: FileItem): MenuProps['items'] => [
    record.type === 'file'
      ? {
          key: 'download',
          icon: <DownloadOutlined />,
          label: '下载',
          onClick: () => handleDownload(record),
        }
      : null,
    {
      key: 'delete',
      icon: <DeleteOutlined />,
      label: '删除',
      danger: true,
      onClick: () => void handleDelete([record]),
    },
  ].filter(Boolean) as MenuProps['items']

  const columns: ColumnsType<FileItem> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: FileItem) => (
        <span className="flex cursor-pointer items-center hover:text-primary" onClick={() => handleEnterDirectory(record)}>
          {record.type === 'directory' ? (
            <FolderOutlined className="mr-2 text-warning" />
          ) : (
            <FileOutlined className="mr-2 text-text-secondary" />
          )}
          {name}
        </span>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: FileItem['type']) => (type === 'directory' ? '文件夹' : '文件'),
    },
    {
      title: '大小',
      dataIndex: 'size',
      key: 'size',
      width: 100,
      render: (size: number, record: FileItem) => (record.type === 'directory' ? '-' : formatFileSize(size)),
    },
    {
      title: '修改时间',
      dataIndex: 'modified_at',
      key: 'modified_at',
      width: 140,
      render: (time: string) => formatDateTime(time, 'MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: unknown, record: FileItem) => (
        <Dropdown menu={{ items: getActionMenu(record) }} trigger={['click']}>
          <Button type="text" size="small" icon={<MoreOutlined />} />
        </Dropdown>
      ),
    },
  ]

  const selectedItems = files.filter((file) => selectedKeys.includes(file.path))

  return (
    <div className="flex h-full flex-col bg-gray-800 text-white">
      {error && (
        <Alert
          message="无法访问文件系统"
          description={error}
          type="error"
          showIcon
          icon={<ExclamationCircleOutlined />}
          className="m-2"
          action={<Button size="small" onClick={() => void refreshFiles()}>重试</Button>}
        />
      )}

      <div className="flex items-center justify-between border-b border-gray-700 p-3">
        <Space>
          <Button size="small" icon={<ArrowLeftOutlined />} onClick={goBack} disabled={currentPath === '/workspace'} />
          <span className="text-sm text-gray-300">{currentPath}</span>
        </Space>
        <Space>
          <Upload {...uploadProps}>
            <Button size="small" icon={<UploadOutlined />}>上传</Button>
          </Upload>
          <Button size="small" icon={<PlusOutlined />} onClick={() => setNewFolderModalVisible(true)}>
            新建文件夹
          </Button>
          <Button size="small" icon={<ReloadOutlined />} onClick={() => void refreshFiles()}>
            刷新
          </Button>
        </Space>
      </div>

      <div className="flex-1 overflow-auto">
        <Table
          columns={columns}
          dataSource={files}
          rowKey="path"
          loading={loading}
          pagination={false}
          size="small"
          className="file-table"
          rowSelection={{
            selectedRowKeys: selectedKeys,
            onChange: (keys) => setSelectedKeys(keys as string[]),
          }}
        />
      </div>

      <div className="flex items-center justify-between border-t border-gray-700 p-3">
        <span className="text-xs text-gray-400">
          共 {files.length} 项，已选择 {selectedKeys.length} 项
        </span>
        <Space>
          <Button size="small" disabled={selectedItems.length === 0} onClick={() => void handleDelete(selectedItems)}>
            批量删除
          </Button>
        </Space>
      </div>

      <Modal
        title="新建文件夹"
        open={newFolderModalVisible}
        onOk={() => void handleCreateFolder()}
        onCancel={() => setNewFolderModalVisible(false)}
        okText="创建"
        cancelText="取消"
      >
        <input
          className="w-full rounded border border-gray-300 px-3 py-2 text-black"
          value={newFolderName}
          onChange={(event) => setNewFolderName(event.target.value)}
          placeholder="请输入文件夹名称"
        />
      </Modal>
    </div>
  )
}
