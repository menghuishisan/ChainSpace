import { useCallback, useEffect, useState } from 'react'

import {
  createRuntimeDirectory,
  deleteRuntimeFile,
  listRuntimeFiles,
} from '@/api/experimentRuntime'
import type { FileItem } from '@/types/presentation'

export function useRuntimeFileManager(accessUrl?: string) {
  const [loading, setLoading] = useState(false)
  const [files, setFiles] = useState<FileItem[]>([])
  const [currentPath, setCurrentPath] = useState('/workspace')
  const [error, setError] = useState<string | null>(null)

  const refreshFiles = useCallback(async () => {
    if (!accessUrl) {
      setError('实验环境未就绪')
      setFiles([])
      return
    }

    setLoading(true)
    setError(null)
    try {
      const nextFiles = await listRuntimeFiles(accessUrl, currentPath)
      setFiles(nextFiles)
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取文件列表失败')
      setFiles([])
    } finally {
      setLoading(false)
    }
  }, [accessUrl, currentPath])

  useEffect(() => {
    void refreshFiles()
  }, [refreshFiles])

  const enterDirectory = useCallback((path: string) => {
    setCurrentPath(path)
  }, [])

  const goBack = useCallback(() => {
    setCurrentPath((previous) => {
      if (previous === '/workspace') {
        return previous
      }
      return previous.split('/').slice(0, -1).join('/') || '/workspace'
    })
  }, [])

  const createDirectory = useCallback(async (name: string) => {
    if (!accessUrl) {
      throw new Error('实验环境未就绪')
    }
    await createRuntimeDirectory(accessUrl, `${currentPath}/${name}`)
    await refreshFiles()
  }, [accessUrl, currentPath, refreshFiles])

  const removeFiles = useCallback(async (paths: string[]) => {
    if (!accessUrl) {
      throw new Error('实验环境未就绪')
    }
    await Promise.all(paths.map((path) => deleteRuntimeFile(accessUrl, path)))
    await refreshFiles()
  }, [accessUrl, refreshFiles])

  return {
    loading,
    files,
    currentPath,
    error,
    refreshFiles,
    enterDirectory,
    goBack,
    createDirectory,
    removeFiles,
  }
}
